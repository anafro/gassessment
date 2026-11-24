package main

import "time"

// Wallet stores the demo player balance in integer minor units. The database
// CHECK constraint prevents negative balances independently of application logic.
type Wallet struct {
    ID      int   `gorm:"primaryKey;autoIncrement:false"`
    Balance int64 `gorm:"not null;check:balance >= 0"`
}

// TableName preserves the original singular table name across the GORM migration.
func (Wallet) TableName() string { return "wallet" }

// Bet is an accepted stake: Key is the unique idempotency key, Amount is minor
// units, and Odds is the accepted price in hundredths. It has no settlement state.
type Bet struct {
    Key       string    `gorm:"primaryKey"`
    MatchID   int       `gorm:"not null"`
    Outcome   int       `gorm:"not null"`
    Amount    int64     `gorm:"not null;check:amount > 0"`
    Odds      int       `gorm:"not null"`
    CreatedAt time.Time `gorm:"default:now()"`
}
