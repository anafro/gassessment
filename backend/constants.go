package main

import "time"

const (
    playerWalletID                = 1
    initialBalanceMinor     int64 = 100000
    maxBetAmountMinor       int64 = 100000000
    oddsScale                     = 100
    maxAcceptedOdds               = 10000
    outcomeCount                  = 3
    oddsSourceCount               = 2
    lineOverround                 = 1.05
    minimumSourceWeight           = 0.2
    lineRefreshInterval           = 5 * time.Second
    eventPublishInterval          = time.Second
    betTransactionTimeout         = 10 * time.Second
    minIdempotencyKeyLength       = 16
    maxIdempotencyKeyLength       = 100
    maxBetRequestBytes            = 4096
    listenAddress                 = ":8080"
)

var matchNames = [...]string{
    "Зенит — Спартак",
    "Динамо — Рубин",
    "Краснодар — ЦСКА",
    "Ростов — Локомотив",
    "Балтика — Ахмат",
    "Оренбург — Сочи",
}
