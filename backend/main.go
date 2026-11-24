package main

import (
    "encoding/json"
    "errors"
    "fmt"
    "io"
    "log"
    "math"
    "math/rand"
    "net/http"
    "os"
    "sync"
    "time"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/clause"
)

type Match struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Odds [3]int `json:"odds"`
}

var lineMutex sync.RWMutex
var line []Match
var database *gorm.DB

func generateLine() []Match {
    names := []string{"Зенит — Спартак", "Динамо — Рубин", "Краснодар — ЦСКА", "Ростов — Локомотив", "Балтика — Ахмат", "Оренбург — Сочи"}
    result := make([]Match, len(names))
    for i, n := range names {
        var p [3]float64
        for range 2 {
            var raw [3]float64
            sum := 0.0
            for j := range raw {
                raw[j] = 0.2 + rand.Float64()
                sum += raw[j]
            }
            for j := range p {
                p[j] += raw[j] / sum / 2
            }
        }
        m := Match{ID: i + 1, Name: n}
        for j := range p {
            m.Odds[j] = int(math.Floor(100 / (p[j] * 1.05)))
        }
        result[i] = m
    }
    return result
}
func respond(w http.ResponseWriter, status int, v any) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    if err := json.NewEncoder(w).Encode(v); err != nil {
        log.Printf("encode response: %v", err)
    }
}
func main() {
    var err error
    database, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    connection, err := database.DB()
    if err != nil {
        log.Fatal(err)
    }
    defer connection.Close()
    if err := database.AutoMigrate(&Wallet{}, &Bet{}); err != nil {
        log.Fatal(err)
    }
    if err := database.Clauses(clause.OnConflict{DoNothing: true}).Create(&Wallet{ID: 1, Balance: 100000}).Error; err != nil {
        log.Fatal(err)
    }
    line = generateLine()
    go func() {
        for range time.Tick(5 * time.Second) {
            lineMutex.Lock()
            line = generateLine()
            lineMutex.Unlock()
        }
    }()
    http.HandleFunc("/api/line", func(w http.ResponseWriter, r *http.Request) {
        lineMutex.RLock()
        defer lineMutex.RUnlock()
        respond(w, http.StatusOK, line)
    })
    http.HandleFunc("/api/events", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "text/event-stream")
        w.Header().Set("Cache-Control", "no-cache")
        f, ok := w.(http.Flusher)
        if !ok {
            return
        }
        t := time.NewTicker(time.Second)
        defer t.Stop()
        for {
            select {
            case <-r.Context().Done():
                return
            case <-t.C:
                lineMutex.RLock()
                payload, err := json.Marshal(line)
                lineMutex.RUnlock()
                if err != nil {
                    log.Printf("encode line: %v", err)
                    return
                }
                if _, err := fmt.Fprintf(w, "data: %s\n\n", payload); err != nil {
                    return
                }
                f.Flush()
            }
        }
    })
    http.HandleFunc("/api/wallet", func(w http.ResponseWriter, r *http.Request) {
        var wallet Wallet
        if err := database.WithContext(r.Context()).First(&wallet, 1).Error; err != nil {
            respond(w, http.StatusServiceUnavailable, "База недоступна")
            return
        }
        respond(w, http.StatusOK, map[string]int64{"balance": wallet.Balance})
    })
    http.HandleFunc("/api/bets", placeBet)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
func placeBet(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        w.Header().Set("Allow", http.MethodPost)
        w.WriteHeader(http.StatusMethodNotAllowed)
        return
    }
    var request BetRequest
    decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
    decoder.DisallowUnknownFields()
    if err := decoder.Decode(&request); err != nil {
        respond(w, http.StatusBadRequest, "Неверный запрос")
        return
    }
    if err := decoder.Decode(new(any)); err != io.EOF {
        respond(w, http.StatusBadRequest, "Ожидается один JSON объект")
        return
    }
    balance, err := acceptBet(r.Context(), database, request)
    if err != nil {
        var rejection *betRejection
        if errors.As(err, &rejection) {
            respond(w, rejection.status, rejection.message)
            return
        }
        log.Printf("accept bet: %v", err)
        respond(w, http.StatusServiceUnavailable, "Результат неизвестен. Повторите тот же запрос.")
        return
    }
    respond(w, http.StatusOK, map[string]any{"balance": balance, "accepted": true})
}
