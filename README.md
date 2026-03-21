# gbkr

A Go client library for the [Interactive Brokers](https://www.interactivebrokers.com/) Client Portal Gateway REST API.

## Features

- **Two-tier session model** — `gbkr.Client` for gateway access, `brokerage.NewSession` for brokerage session capabilities
- **Handle-based API** — capabilities accessed via purpose-built handle types (`Portfolio`, `Analysis`, `Accounts`, `MarketData`, `Contracts`, `Trades`)
- **Automatic API pacing** — built-in rate limiting and concurrency control matching IBKR's documented limits
- **Strongly-typed domain aliases** — `AccountID`, `ConID`, `Currency`, `BarSize`, `TimePeriod` prevent parameter confusion
- **Structured error model** — const sentinel errors with `errors.Is`/`errors.As` support

## Install

```bash
go get github.com/trippwill/gbkr
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/trippwill/gbkr"
    "github.com/trippwill/gbkr/brokerage"
)

func main() {
    client, err := gbkr.NewClient(
        gbkr.WithBaseURL("https://localhost:5000/v1/api"),
        gbkr.WithInsecureTLS(),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Elevate to a brokerage session (SSO/DH handshake).
    bc, err := brokerage.NewSession(context.Background(), client, &brokerage.SSOInitRequest{})
    if err != nil {
        log.Fatal(err)
    }

    // Use handle-based API.
    accounts, err := bc.Accounts().List(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(accounts)
}
```

## Session Model

gbkr mirrors the IBKR gateway's two-phase session lifecycle:

| Tier | Type | How to get | Capabilities |
|------|------|------------|--------------|
| Gateway | `*gbkr.Client` | `gbkr.NewClient(opts...)` | `SessionStatus`, `Portfolio`, `Analysis` |
| Brokerage | `*brokerage.Client` | `brokerage.NewSession(ctx, client, req)` | `Accounts`, `MarketData`, `Contracts`, `SecurityDefinitions`, `Trades` |

### Capability-to-Path Mapping

| Capability | Access Point | IBKR Path Prefix |
|------------|-------------|-----------------|
| Portfolio | `Client.Portfolio()` | `/portfolio/{accountId}/*` |
| Analysis | `Client.Analysis()` | `/pa/*` |
| Accounts | `brokerage.Client.Accounts()` | `/iserver/accounts` |
| MarketData | `brokerage.Client.MarketData()` | `/iserver/marketdata/*` |
| Contracts | `brokerage.Client.Contracts()` | `/iserver/contract/{conid}/*` |
| SecurityDefinitions | `brokerage.Client.SecurityDefinitions()` | `/iserver/secdef/*` |
| Trades | `brokerage.Client.Trades()` | `/iserver/account/trades` |

## Capability Separation (ADR-008)

Types are co-located with their capabilities — no separate `models` package. Shared
primitives (`AccountID`, `ConID`, `Currency`, etc.) live in `types.go` in the root package;
brokerage-specific types (`BarSize`, `SnapshotField`, etc.) live in `brokerage/types.go`.

Dangerous capabilities (trading, banking) will live in **separate Go modules**
(`gbkr/trading`, `gbkr/banking`) so that consuming applications must explicitly import
them — making the dependency visible in `go.mod`. See [ADR-006](https://github.com/trippwill/midwatch.work/blob/main/docs/decisions/006-structural-capability-separation.md) for rationale.

## API Pacing

The client automatically enforces IBKR's pacing limits:

- Global ceiling of 10 requests/second
- Per-endpoint rate limits for sensitive paths
- Concurrency semaphores where documented

Pacing can be disabled for testing with `WithRateLimit(nil)` or observed via `WithPacingObserver`.

## Error Handling

All domain errors are inspectable via standard Go error patterns:

```go
import "errors"

// Check sentinel errors
if errors.Is(err, gbkr.ErrAPIRequest) {
    // non-2xx API response
}

// Extract structured context
var apiErr *gbkr.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Status)
}
```

## CLI

A test CLI is included for exercising the library:

```bash
go run ./cmd/gbkr --insecure
```

## Field Name Mapping

Model structs use friendly Go names where the IBKR API uses abbreviations or
inconsistent casing. The table below lists every rename; JSON serialization
uses the original API keys so wire compatibility is preserved.

### PnLEntry (`GET /iserver/account/pnl/partitioned`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `DailyPnL` | `dpl` | Daily profit/loss |
| `NetLiquidation` | `nl` | Net liquidity |
| `UnrealizedPnL` | `upl` | Unrealized profit/loss |
| `RealizedPnL` | `rpl` | Realized profit/loss |
| `ExcessLiquidity` | `el` | Excess liquidity |
| `MarginValue` | `mv` | Margin value |

### AccountList (`GET /iserver/accounts`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `SelectedAcct` | `selectedAccount` | Currently selected account |

### Position (`GET /portfolio/{accountId}/positions/{pageId}`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `Qty` | `position` | Total position size |

### LedgerEntry (`GET /portfolio/{accountId}/ledger`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `FuturesPnL` | `futuresonlypnl` | Futures position PnL |
| `FutureOptionValue` | `futureoptionmarketvalue` | Futures options market value |
| `NetLiquidation` | `netliquidationvalue` | Net liquidation value |

### HistoryResponse / HistoryBar (`GET /iserver/marketdata/history`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `Bars` | `data` | Array of OHLCV bars |
| `Open` | `o` | Bar open price |
| `High` | `h` | Bar high price |
| `Low` | `l` | Bar low price |
| `Close` | `c` | Bar close price |
| `Volume` | `v` | Bar volume |
| `Time` | `t` | Epoch timestamp |

### Snapshot (`GET /iserver/marketdata/snapshot`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `ServerID` | `server_id` | Internal server identifier |
| `UpdateTime` | `_updated` | Last update timestamp |

### SessionStatus (`POST /iserver/auth/ssodh/init`, `POST /iserver/auth/status`)

| Go Field | API Key | Description |
|----------|---------|-------------|
| `HardwareInfo` | `hardware_info` | Hardware info string |

## Flex Web Service (`gbkr/flex`)

The `flex` subpackage is a separate Go module for retrieving IBKR Activity Statement reports
via the Flex Web Service — a standalone IBKR API independent of the Client Portal Gateway.

```bash
go get github.com/trippwill/gbkr/flex
```

### What it does

- Fetches Activity Statement XML reports using a long-lived API token and pre-configured query ID
- Parses Trades, Cash Transactions, Option Events, and Commission Details sections
- Handles the two-step retrieval protocol (SendRequest → poll GetStatement) with configurable retry/backoff
- Optionally saves raw response bodies to disk for debugging
- Detects common misconfigurations (CSV output, expired tokens, query errors)

### Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/trippwill/gbkr/flex"
)

func main() {
    c := flex.NewClient(
        flex.WithReportDir("/tmp/flex-reports"), // optional: save raw XML
    )

    resp, err := c.FetchReport(context.Background(), "YOUR_TOKEN", "YOUR_QUERY_ID",
        flex.WithMaxRetries(10),
        flex.WithInitialDelay(5*time.Second),
    )
    if err != nil {
        log.Fatal(err)
    }

    for _, stmt := range resp.Statements {
        fmt.Printf("Account %s: %d trades\n", stmt.AccountID, len(stmt.Trades))
    }
}
```

### Error Types

```go
import "errors"

// Check for specific conditions
if errors.Is(err, flex.ErrWrongFormat) {
    // Query Format is not set to XML in IBKR Flex query template
}
if errors.Is(err, flex.ErrTokenExpired) {
    // API token needs to be renewed in IBKR Client Portal
}
```

## Development

```bash
mise run precommit   # fmt → build → test with race detection
mise run ci          # full CI pipeline
mise run vet         # golangci-lint
```

See [AGENTS.md](AGENTS.md) for full development guidelines.

## License

See [LICENSE](LICENSE) for details.
