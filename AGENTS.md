# Repository Guidelines

## Build, Test, and Development Commands

Prefer `mise` tasks over raw Go commands for consistent environment and task chaining.

### Key mise tasks
- `mise run ci`: full CI pipeline (fmt-check → vet → build → coverage)
- `mise run precommit`: fmt → build → test with race detection
- `mise run build`: compile (depends on vet)
- `mise run vet`: golangci-lint
- `mise run test`: run all tests without cache
- `mise run test-race`: tests with race detector

### Linting with golangci-lint
Linting uses `golangci-lint` v2 (managed by mise):
- Config: `.golangci.yml` at repo root
- Enabled linters: govet, staticcheck, gosec, errcheck, errname, errorlint, gocritic, revive, and more
- Generated files excluded via `exclusions.generated: strict`
- Run via `mise run vet` or directly: `golangci-lint run ./...`

### Raw Go commands (when needed)
```bash
go build ./...                    # build all packages
go test ./... -count=1            # run all tests
go test ./... -count=1 -race      # run tests with race detection
go vet ./...                      # basic vet
go fmt ./...                      # format code
```

## Architecture

### Package Layout

```
doc.go               # Package documentation
gbkr.go              # Core Client struct, HTTP transport (doGet, doPost)
gbkr_err.go          # Error type, sentinel errors, APIError
options.go           # Functional options (WithBaseURL, WithHTTPClient, WithLogger, etc.)
operation.go         # Operation type, Op* constants, emitOp helper (ADR-007)
auth.go              # SessionClient interface + Session(c) constructor
accounts.go          # AccountLister + AccountReader interfaces + constructors
positions.go         # PositionReader interface + Positions(c, id) constructor
marketdata.go        # MarketDataReader interface + MarketData(c) constructor
trades.go            # TradeReader interface + BrokerageClient.Trades() method
contracts.go         # ContractReader interface + BrokerageClient.Contracts() method
example_test.go      # Go doc examples for capability constructors (package gbkr_test)
models/
  doc.go             # Package documentation
  json.go            # Generic deref[T] helper
  types.go           # Strongly-typed aliases: AccountID, ConID, Currency, Exchange, OrderID, AlertID, BarSize, TimePeriod
  types_err.go       # Validation error sentinels and constructors for BarSize/TimePeriod
  auth.go            # SessionStatus, SSOInitRequest
  accounts.go        # AccountList, AccountSummary, PnLPartitioned
  positions.go       # Position (incl. option fields), PortfolioSummary, Ledger
  marketdata.go      # SnapshotParams, Snapshot, HistoryParams, HistoryResponse
  trades.go          # TradeExecution, Transaction, TransactionHistoryRequest/Response
  contracts.go       # ContractDetails, ContractSearchResult
apispec/
  apispec.go         # Embedded OpenAPI spec (api-docs.json)
  api-docs.json      # IBKR Client Portal API specification
cmd/
  gbkr/
    main.go          # CLI test harness
gateway/             # Container config for local IBKR gateway (not a Go package)
```

### Capability Separation

There is no runtime permission system. Dangerous capabilities (trading, banking) will
live in separate Go modules (`gbkr/trading`, `gbkr/banking`) within the same repository,
each with their own `go.mod`. This makes the dependency visible in any consumer's `go.mod`.
See ADR-006 for the rationale.

### Operation Event Sourcing

Every IBKR API call emits a structured `slog` record via the client's logger, namespaced
under the `"gbkr"` group (see ADR-007). Key points:

- **Logger**: `Client` holds a `*slog.Logger` (default: `slog.Default().WithGroup("gbkr")`).
  Consumers can override via `WithLogger(*slog.Logger)`.
- **Operation constants**: `type Operation string` constants in `operation.go` are the
  stable event vocabulary. Their string values are part of the wire contract — do not rename after a stable module version is declared.
- **Emission helper**: `emitOp()` on `*Client` handles level selection (Info/Warn), timing,
  and attribute construction. Call it in every public method that makes an IBKR API call.
- **Attributes**: Emit only relevant context fields per operation (`account_id`, `symbol`,
  `conid`). Do not emit zero-value attributes. Never emit request/response body content.
- **When adding new API methods**: Define a new `Op*` constant and call `emitOp()` following
  the same pattern as existing methods.

## Conventions

### Code Style
- Format with `gofmt` (run `mise run fmt` before commits)
- Exported identifiers use `CamelCase`; unexported use `lowerCamel`
- File names are lowercase with underscores where needed
- Package documentation goes in `doc.go`
- Error definitions go in dedicated `*_err.go` files

### Error Handling
- Define `type Error string` as the base sentinel type in `gbkr_err.go`
- Use `const` sentinels for errors callers check with `errors.Is` (e.g., `ErrAPIRequest`)
- Use struct types with `Unwrap()` for errors needing context (e.g., `APIError`)
- Use `Err*()` constructor functions for ergonomic error creation
- Use `fmt.Errorf("context: %w", err)` only for internal wrapping not inspected by callers
- Callers check errors via `errors.Is(err, gbkr.ErrAPIRequest)` or `errors.As(err, &apiErr)`

### Naming Conventions
- Avoid "Get" prefix for getters — use noun-like names (e.g., `Summary()` not `GetSummary()`)
- Initialisms should be consistently cased: `ID`, `URL`, `HTTP` (exported) or `id`, `url`, `http` (unexported)
- Don't repeat package name in function names (e.g., `models.Bar()` not `models.NewBarSize()`)

### Commits & Pull Requests
- Run `mise run precommit` before committing
- Use Conventional Commits with scope when helpful (e.g., `feat(positions): ...`)
- Reference issues/PRs in commit messages when applicable

### Testing
- Standard Go `testing` package; tests are `*_test.go` next to code
- Use `-count=1` to disable test caching
- Use `-race` for concurrent safety (pre-commit)

## Agent Notes
- **Green-field**: No module version has been declared. Prefer improving architecture and consistency over backward compatibility.
- Prefer `mise` tasks for consistent environment (cache paths are set in `.mise/config.toml`)
- Run `mise run precommit` before commits
- Run `mise run ci` to validate changes match what GitHub Actions will run
- Linting uses `golangci-lint` v2 (managed by mise); config in `.golangci.yml`
