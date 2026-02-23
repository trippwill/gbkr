# gbkr

A Go client library for the [Interactive Brokers](https://www.interactivebrokers.com/) Client Portal Gateway REST API.

## Features

- **Two-phase session model** — `Client` for gateway access, `BrokerageClient` for brokerage session capabilities
- **Narrow capability interfaces** — consumers get purpose-built types (`AccountLister`, `AccountReader`, `PositionReader`, `MarketDataReader`, `ContractReader`, `TradeReader`)
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
    "github.com/trippwill/gbkr/models"
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
    bc, err := client.BrokerageSession(context.Background(), &models.SSOInitRequest{})
    if err != nil {
        log.Fatal(err)
    }

    // Use narrow capability interfaces.
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
| Gateway | `*Client` | `NewClient(opts...)` | `SessionStatus`, `Portfolio`, `Analysis` |
| Brokerage | `*BrokerageClient` | `client.BrokerageSession(ctx, req)` | `Accounts`, `Account`, `MarketData`, `Contracts`, `SecurityDefinitions`, `Trades` + all gateway capabilities |

### Capability-to-Path Mapping

| Capability | Access Point | IBKR Path Prefix |
|------------|-------------|-----------------|
| Portfolio | `Client.Portfolio()` | `/portfolio/{accountId}/*` |
| Analysis | `Client.Analysis()` | `/pa/*` |
| Accounts | `BrokerageClient.Accounts()` | `/iserver/accounts` |
| Account | `BrokerageClient.Account()` | `/iserver/account/{id}/*` |
| MarketData | `BrokerageClient.MarketData()` | `/iserver/marketdata/*` |
| Contracts | `BrokerageClient.Contracts()` | `/iserver/contract/{conid}/*` |
| SecurityDefinitions | `BrokerageClient.SecurityDefinitions()` | `/iserver/secdef/*` |
| Trades | `BrokerageClient.Trades()` | `/iserver/account/trades` |

## Capability Separation

gbkr currently provides read-only access to the IBKR API. Dangerous capabilities
(trading, banking) will live in **separate Go modules** (`gbkr/trading`, `gbkr/banking`)
so that consuming applications must explicitly import them — making the dependency
visible in `go.mod`. See [ADR-006](https://github.com/trippwill/midwatch.work/blob/main/docs/decisions/006-structural-capability-separation.md) for rationale.

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

## Development

This project uses [mise](https://mise.jdx.dev/) for task automation:

```bash
mise run precommit   # fmt → build → test with race detection
mise run ci          # full CI pipeline
mise run vet         # golangci-lint
```

See [AGENTS.md](AGENTS.md) for full development guidelines.

## License

See [LICENSE](LICENSE) for details.
