package main

import (
    "context"
    "encoding/json"
    "fmt"
    "github.com/jackc/pgx/v5/pgxpool"
    "log"
    "math"
    "math/rand"
    "net/http"
    "os"
    "strconv"
    "sync"
    "time"
)

type Match struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
    Odds [3]int `json:"odds"`
}

var mu sync.RWMutex
var line []Match
var db *pgxpool.Pool

func generate() []Match {
    names := []string{"Зенит — Спартак", "Динамо — Рубин", "Краснодар — ЦСКА", "Ростов — Локомотив", "Балтика — Ахмат", "Оренбург — Сочи"}
    result := make([]Match, len(names))
    for i, n := range names {
        var p [3]float64
        for source := 0; source < 2; source++ {
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
    json.NewEncoder(w).Encode(v)
}
func main() {
    ctx := context.Background()
    var err error
    db, err = pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
    if err != nil {
        log.Fatal(err)
    }
    _, err = db.Exec(ctx, `CREATE TABLE IF NOT EXISTS wallet(id int PRIMARY KEY,balance bigint NOT NULL CHECK(balance>=0)); INSERT INTO wallet VALUES(1,100000) ON CONFLICT DO NOTHING; CREATE TABLE IF NOT EXISTS bets(key text PRIMARY KEY, match_id int NOT NULL, outcome int NOT NULL, amount bigint NOT NULL CHECK(amount>0), odds int NOT NULL, created_at timestamptz DEFAULT now());`)
    if err != nil {
        log.Fatal(err)
    }
    line = generate()
    go func() {
        for range time.Tick(5 * time.Second) {
            mu.Lock()
            line = generate()
            mu.Unlock()
        }
    }()
    http.HandleFunc("/api/line", func(w http.ResponseWriter, r *http.Request) { mu.RLock(); defer mu.RUnlock(); respond(w, 200, line) })
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
                mu.RLock()
                b, _ := json.Marshal(line)
                mu.RUnlock()
                fmt.Fprintf(w, "data: %s\n\n", b)
                f.Flush()
            }
        }
    })
    http.HandleFunc("/api/wallet", func(w http.ResponseWriter, r *http.Request) {
        var balance int64
        if err := db.QueryRow(r.Context(), "SELECT balance FROM wallet WHERE id=1").Scan(&balance); err != nil {
            respond(w, 503, "База недоступна")
            return
        }
        respond(w, 200, map[string]int64{"balance": balance})
    })
    http.HandleFunc("/api/bets", bet)
    log.Fatal(http.ListenAndServe(":8080", nil))
}
func bet(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        w.WriteHeader(405)
        return
    }
    var b struct {
        Key     string `json:"key"`
        Match   int    `json:"match"`
        Outcome int    `json:"outcome"`
        Amount  string `json:"amount"`
        Odds    int    `json:"odds"`
    }
    if json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&b) != nil {
        respond(w, 400, "Неверный запрос")
        return
    }
    amount, err := strconv.ParseInt(b.Amount, 10, 64)
    if err != nil || amount <= 0 || amount > 100000000 || len(b.Key) < 16 || len(b.Key) > 100 || b.Match < 1 || b.Match > 6 || b.Outcome < 0 || b.Outcome > 2 {
        respond(w, 400, "Неверные параметры")
        return
    }
    tx, err := db.Begin(r.Context())
    if err != nil {
        respond(w, 503, "Попробуйте снова")
        return
    }
    defer tx.Rollback(r.Context())
    var balance int64
    if tx.QueryRow(r.Context(), "SELECT balance FROM wallet WHERE id=1 FOR UPDATE").Scan(&balance) != nil {
        respond(w, 503, "Попробуйте снова")
        return
    }
    var oldMatch, oldOutcome, oldOdds int
    var oldAmount int64
    err = tx.QueryRow(r.Context(), "SELECT match_id,outcome,amount,odds FROM bets WHERE key=$1", b.Key).Scan(&oldMatch, &oldOutcome, &oldAmount, &oldOdds)
    if err == nil {
        if oldMatch != b.Match || oldOutcome != b.Outcome || oldAmount != amount || oldOdds != b.Odds {
            respond(w, 409, "Ключ уже использован")
            return
        }
        respond(w, 200, map[string]any{"balance": balance, "accepted": true})
        return
    }
    mu.RLock()
    defer mu.RUnlock()
    if line[b.Match-1].Odds[b.Outcome] != b.Odds {
        respond(w, 409, "Коэффициент изменился. Выберите исход снова.")
        return
    }
    if balance < amount {
        respond(w, 422, "Недостаточно средств")
        return
    }
    _, err = tx.Exec(r.Context(), "INSERT INTO bets(key,match_id,outcome,amount,odds) VALUES($1,$2,$3,$4,$5)", b.Key, b.Match, b.Outcome, amount, b.Odds)
    if err == nil {
        _, err = tx.Exec(r.Context(), "UPDATE wallet SET balance=balance-$1 WHERE id=1", amount)
    }
    if err == nil {
        err = tx.Commit(r.Context())
    }
    if err != nil {
        respond(w, 503, "Результат неизвестен. Повторите тот же запрос.")
        return
    }
    respond(w, 200, map[string]any{"balance": balance - amount, "accepted": true})
}
