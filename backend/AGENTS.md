# Backend instructions

Inherit ../AGENTS.md. This module is a Go 1.23 application using net/http and GORM with the Postgres driver.

## Code style

- Follow idiomatic Go: short package names, MixedCaps identifiers, explicit errors, early returns and context propagation for database operations.
- Prefer modern features supported by Go 1.23, such as `any` and integer ranges when they improve clarity. Use standard-library helpers instead of custom equivalents. Avoid clever abstractions and unnecessary packages.
- Use gofmt structure with all tabs expanded to four spaces. The root formatter applies this deliberate exception to Go's default style; do not finish with raw gofmt output.
- Group standard-library imports separately from third-party imports. Keep HTTP decoding bounded, validate input, and return consistent errors without leaking database details.
- Amounts use int64 integer minor units; odds use integer hundredths. Keep wallet locking, idempotency lookup, bet insertion and debit within the same transaction. Check database errors and preserve rollback paths.
- Protect shared line data with the existing mutex. Do not introduce multiple API instances without shared line/version coordination.
- Prefer focused table-driven tests for domain rules and integration tests for database/concurrency guarantees. Do not substitute mocks for transaction correctness.

## Database conventions

- Use GORM models and the Postgres driver. Keep explicit table/column names compatible with existing data; Wallet maps to wallet and Bet to bets.
- Pass request contexts with WithContext. Check Error on every operation; distinguish gorm.ErrRecordNotFound from database failures.
- Acquire clause.Locking with UPDATE on the wallet inside the same transaction as the idempotency check, bet insertion and debit. Use only the transaction handle until commit or rollback.
- Use parameterized conditions and gorm.Expr for atomic balance updates. Preserve database CHECK constraints and primary keys.
- AutoMigrate is acceptable for this demo; production schema changes require reviewed versioned migrations. Never reset data to migrate.

## Commands

From the repository root:

```powershell
Push-Location backend
go mod tidy
go vet ./...
go test ./...
go test -race ./...
go build ./...
Pop-Location
pwsh -File scripts/lint.ps1 -Fix
pwsh -File scripts/lint.ps1
```

The race detector requires a supported platform and C toolchain. `go mod tidy` is for dependency changes and can restore tabs in go.mod; expand them to four spaces before completing. The full validation script configures Go caches under .cache. Run the full root validation after every change.


## Money safety

- Execute acceptance with a GORM Transaction callback; rollback on every error and panic, and return success only after commit.
- Lock the wallet before idempotency lookup. Replay identical requests before checking current odds; reject conflicting payloads.
- Debit with a guarded balance predicate, check RowsAffected equals one, and keep insertion/debit atomic. Use bounded database contexts.
- Reject malformed, noncanonical or out-of-range monetary input. Retain database checks and unique keys as independent safeguards.
