package main

import "time"

// Wallet keeps the existing singular table name so upgrades preserve balances.
type Wallet struct {
    ID      int   `gorm:"primaryKey;autoIncrement:false"`
    Balance int64 `gorm:"not null;check:balance >= 0"`
}

func (Wallet) TableName() string { return "wallet" }

type Bet struct {
    Key       string    `gorm:"primaryKey"`
    MatchID   int       `gorm:"not null"`
    Outcome   int       `gorm:"not null"`
    Amount    int64     `gorm:"not null;check:amount > 0"`
    Odds      int       `gorm:"not null"`
    CreatedAt time.Time `gorm:"default:now()"`
}
