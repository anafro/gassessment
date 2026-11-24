package main

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "strconv"

    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type BetRequest struct {
    Key     string `json:"key"`
    Match   int    `json:"match"`
    Outcome int    `json:"outcome"`
    Amount  string `json:"amount"`
    Odds    int    `json:"odds"`
}

type betRejection struct {
    status  int
    message string
}

func (e *betRejection) Error() string { return e.message }

func acceptBet(ctx context.Context, db *gorm.DB, request BetRequest) (int64, error) {
    amount, err := strconv.ParseInt(request.Amount, 10, 64)
    validAmount := err == nil && Between(amount, 1, maxBetAmountMinor) &&
        strconv.FormatInt(amount, 10) == request.Amount
    validKey := Between(len(request.Key), minIdempotencyKeyLength, maxIdempotencyKeyLength)
    validMatch := Between(request.Match, 1, len(matchNames))
    validOutcome := Between(request.Outcome, 0, outcomeCount-1)
    validOdds := Between(request.Odds, oddsScale+1, maxAcceptedOdds)
    if !validAmount || !validKey || !validMatch || !validOutcome || !validOdds {
        return 0, &betRejection{http.StatusBadRequest, "Неверные параметры"}
    }
    ctx, cancel := context.WithTimeout(ctx, betTransactionTimeout)
    defer cancel()
    var balance int64
    err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        var wallet Wallet
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, playerWalletID).Error; err != nil {
            return err
        }
        var existing Bet
        err := tx.Where("key = ?", request.Key).First(&existing).Error
        if err == nil {
            if existing.MatchID != request.Match || existing.Outcome != request.Outcome || existing.Amount != amount || existing.Odds != request.Odds {
                return &betRejection{http.StatusConflict, "Ключ уже использован"}
            }
            balance = wallet.Balance
            return nil
        }
        if !errors.Is(err, gorm.ErrRecordNotFound) {
            return err
        }
        lineMutex.RLock()
        defer lineMutex.RUnlock()
        if len(line) < request.Match || line[request.Match-1].Odds[request.Outcome] != request.Odds {
            return &betRejection{http.StatusConflict, "Коэффициент изменился. Выберите исход снова."}
        }
        if wallet.Balance < amount {
            return &betRejection{http.StatusUnprocessableEntity, "Недостаточно средств"}
        }
        bet := Bet{Key: request.Key, MatchID: request.Match, Outcome: request.Outcome, Amount: amount, Odds: request.Odds}
        if err := tx.Create(&bet).Error; err != nil {
            return err
        }
        result := tx.Model(&Wallet{}).Where("id = ? AND balance >= ?", playerWalletID, amount).UpdateColumn("balance", gorm.Expr("balance - ?", amount))
        if result.Error != nil {
            return result.Error
        }
        if result.RowsAffected != 1 {
            return fmt.Errorf("wallet debit affected %d rows", result.RowsAffected)
        }
        balance = wallet.Balance - amount
        return nil
    })
    return balance, err
}
