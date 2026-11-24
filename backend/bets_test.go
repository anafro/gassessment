package main

import (
    "context"
    "fmt"
    "os"
    "sync"
    "testing"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
)

func TestMoneyPostgres(t *testing.T) {
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("set TEST_DATABASE_URL to an isolated disposable Postgres database")
    }
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    connection, err := db.DB()
    if err != nil {
        t.Fatal(err)
    }
    defer connection.Close()
    // This test owns only its unique schema, never application tables.
    schema := fmt.Sprintf("money_test_%d", os.Getpid())
    if err := db.Exec("CREATE SCHEMA " + schema).Error; err != nil {
        t.Fatal(err)
    }
    defer db.Exec("DROP SCHEMA " + schema + " CASCADE")
    isolated, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{})
    if err != nil {
        t.Fatal(err)
    }
    pool, _ := isolated.DB()
    defer pool.Close()
    if err := isolated.AutoMigrate(&Wallet{}, &Bet{}); err != nil {
        t.Fatal(err)
    }
    if err := isolated.Create(&Wallet{ID: 1, Balance: 1000}).Error; err != nil {
        t.Fatal(err)
    }
    line = []Match{{ID: 1, Odds: [3]int{200, 300, 400}}}
    request := BetRequest{Key: "same-request-key-123", Match: 1, Outcome: 0, Amount: "600", Odds: 200}
    var group sync.WaitGroup
    for range 10 {
        group.Add(1)
        go func() {
            defer group.Done()
            if _, err := acceptBet(context.Background(), isolated, request); err != nil {
                t.Error(err)
            }
        }()
    }
    group.Wait()
    var wallet Wallet
    isolated.First(&wallet, 1)
    var count int64
    isolated.Model(&Bet{}).Count(&count)
    if wallet.Balance != 400 || count != 1 {
        t.Fatalf("balance=%d bets=%d", wallet.Balance, count)
    }
    conflicting := request
    conflicting.Amount = "100"
    if _, err := acceptBet(context.Background(), isolated, conflicting); err == nil {
        t.Fatal("conflicting replay accepted")
    }
    request.Key = "another-request-key"
    if _, err := acceptBet(context.Background(), isolated, request); err == nil {
        t.Fatal("overdraft accepted")
    }
    request.Amount = "100"
    // Failure after insertion must roll back both the inserted bet and debit.
    if err := isolated.Exec("CREATE FUNCTION reject_debit() RETURNS trigger LANGUAGE plpgsql AS $$ BEGIN RAISE EXCEPTION 'injected failure'; END $$").Error; err != nil {
        t.Fatal(err)
    }
    if err := isolated.Exec("CREATE TRIGGER fail_debit BEFORE UPDATE ON wallet FOR EACH ROW EXECUTE FUNCTION reject_debit()").Error; err != nil {
        t.Fatal(err)
    }
    if _, err := acceptBet(context.Background(), isolated, request); err == nil {
        t.Fatal("failed debit accepted")
    }
    isolated.First(&wallet, 1)
    isolated.Model(&Bet{}).Count(&count)
    if wallet.Balance != 400 || count != 1 {
        t.Fatalf("rollback: balance=%d bets=%d", wallet.Balance, count)
    }
    if err := isolated.Exec("DROP TRIGGER fail_debit ON wallet").Error; err != nil {
        t.Fatal(err)
    }
    for index := range 10 {
        group.Add(1)
        go func() {
            defer group.Done()
            candidate := request
            candidate.Key = fmt.Sprintf("distinct-request-%03d", index)
            _, _ = acceptBet(context.Background(), isolated, candidate)
        }()
    }
    group.Wait()
    isolated.First(&wallet, 1)
    isolated.Model(&Bet{}).Count(&count)
    if wallet.Balance != 0 || count != 5 {
        t.Fatalf("concurrent debit: balance=%d bets=%d", wallet.Balance, count)
    }
}
