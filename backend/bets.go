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

// BetRequest identifies one logical attempt. Amount is a canonical decimal string
// of minor units; Odds is integer hundredths and Outcome is a zero-based index.
// Retrying requires the original key and every original payload field.
type BetRequest struct {
    Key     string `json:"key"`
    Match   int    `json:"match"`
    Outcome int    `json:"outcome"`
    Amount  string `json:"amount"`
    Odds    int    `json:"odds"`
}

// betRejection is a definitive input/domain rejection safe to show to the player.
type betRejection struct {
    status  int
    message string
}

// Error returns the public rejection message.
func (e *betRejection) Error() string { return e.message }

// acceptBet atomically records a bet and debits the single demo wallet.
// A successful return follows commit; a replay returns the current locked balance.
// On infrastructure errors, callers must not infer whether commit reached Postgres.
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
        // Serialize requests before checking the key so concurrent retries see
        // the committed bet instead of attempting a second debit.
        var wallet Wallet
        if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&wallet, playerWalletID).Error; err != nil {
            return err
        }
        // Replays precede live-odds checks: a changed line must not invalidate
        // recovery of a bet already accepted under this key.
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
        // Freeze the snapshot through validation and writes in this callback.
        // The read-lock ends before GORM performs the final commit.
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
        // The guard and row count add protection beyond the wallet lock. Any
        // failure rolls back the preceding INSERT along with this transaction.
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
