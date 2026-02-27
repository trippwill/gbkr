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
gbkr.go              # Client struct, NewClient, Transport() accessor
gbkr_err.go          # Error type, ErrBaseURLRequired, re-exported APIError/ErrAPIRequest
options.go           # Functional options (WithBaseURL, WithHTTPClient, WithLogger, etc.)
operation.go         # Operation type, Op* constants, emitOp helper (ADR-007)
auth.go              # Client.SessionStatus() + SessionStatus DTO
portfolio.go         # Portfolio handle + DTOs (Position, PortfolioSummary, Ledger)
analysis.go          # Analysis handle + DTOs (Transaction, TransactionHistory*)
cache.go             # ttlCache, Cached[T]
pacing.go            # PacingPolicy, PacingObserver, pacing rules
pacing_err.go        # ErrPacingWait
types.go             # Shared primitives: AccountID, ConID, Currency, Exchange, etc.
doc.go               # Package documentation
brokerage/
  doc.go             # Package documentation
  brokerage.go       # brokerage.Client, NewSession, SSOInitRequest
  accounts.go        # Accounts, Account + DTOs
  contracts.go       # Contracts + DTOs (ContractDetails, ContractSearchResult)
  marketdata.go      # MarketData + DTOs + SnapshotField constants
  secdef.go          # SecurityDefinitions
  trades.go          # Trades + TradeExecution DTO
  types.go           # BarSize, TimePeriod, SnapshotField + 100+ constants
  types_err.go       # ValidationError for BarSize/TimePeriod
internal/
  transport/
    transport.go     # Transport struct, Get, Post, doRequest, EmitOp
    transport_err.go # APIError, ErrAPIRequest
  jx/
    jx.go            # Deref[T] JSON helper
apispec/
  apispec.go         # Embedded OpenAPI spec (api-docs.json)
  api-docs.json      # IBKR Client Portal API specification
cmd/
  gbkr/
    main.go          # CLI test harness
gateway/             # Container config for local IBKR gateway (not a Go package)
```

### Capability Separation (ADR-008)

The package uses a two-tier client model mirroring the IBKR gateway session lifecycle:

1. **`gbkr.Client`** — created via `NewClient(opts...)`. Provides gateway capabilities:
   `SessionStatus`, `Portfolio`, and `Analysis`.

2. **`brokerage.Client`** — obtained via `brokerage.NewSession(ctx, client, req)`, which
   performs an SSO/DH handshake to elevate to a full brokerage session. Provides brokerage
   capabilities: `Accounts`, `MarketData`, `Contracts`, `SecurityDefinitions`, and `Trades`.

DTOs are co-located with their capability (no separate `models/` package). Shared
primitives (`AccountID`, `ConID`, `Currency`, etc.) live in `types.go` in the root package.

Dangerous capabilities (trading, banking) will live in **separate Go modules**
(`gbkr/trading`, `gbkr/banking`) within the same repository. See ADR-006 for the rationale.

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
