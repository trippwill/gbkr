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

### Package Roles

**`gbkr` (root package)** — Gateway-level client and shared primitives.
`Client` wraps an internal HTTP transport and provides gateway capabilities that do not require brokerage session elevation: session auth (status, logout, reauthenticate, tickle), portfolio (positions, summary, ledger, allocation), analysis (transaction history with TTL caching), portfolio account discovery, and trading schedule lookup.
Also defines: functional options (`WithBaseURL`, `WithHTTPClient`, `WithInsecureTLS`, `WithRateLimit`, `WithPacingObserver`, `WithLogger`), `Operation` constants for event sourcing, `PacingPolicy` with per-endpoint rate limits, `Cached[T]` for time-stamped cache results, and shared type aliases (`AccountID`, `ConID`, `Currency`, `Exchange`, etc.).

**`brokerage/`** — Elevated brokerage session capabilities.
`brokerage.Client` is obtained via `brokerage.NewSession(ctx, client, req)`, which performs an SSO/DH handshake. Provides: account listing and summary, market data (snapshots and history), contract info and search, security definitions, and trade history. DTOs are co-located here (no separate `models/` package). References shared primitives as `gbkr.AccountID`, `gbkr.ConID`, etc.

**`internal/transport/`** — HTTP plumbing shared by both tiers.
`Transport` struct holds base URL, HTTP client, logger, and request hook. Provides `Get`, `Post`, and `EmitOp` methods. Defines `APIError` and `ErrAPIRequest` (re-exported by the root package). Invisible to external consumers.

**`internal/jx/`** — `Deref[T]` generic helper for safe pointer dereference in JSON unmarshaling.

**`cmd/gbkr/`** — CLI test harness demonstrating two-tier session model.

**`gateway/`** — Container config for running the IBKR Client Portal Gateway locally (Containerfile, conf.yaml, helper scripts). Not a Go package.

### Capability Separation (ADR-008)

The package uses a two-tier client model mirroring the IBKR gateway session lifecycle:

1. **`gbkr.Client`** — created via `NewClient(opts...)`. Provides gateway capabilities:
   `SessionStatus`, `Logout`, `Reauthenticate`, `Tickle`, `Portfolio`, `Analysis`,
   `PortfolioAccounts`, and `TradingSchedule`.

2. **`brokerage.Client`** — obtained via `brokerage.NewSession(ctx, client, req)`, which
   performs an SSO/DH handshake to elevate to a full brokerage session. Provides brokerage
   capabilities: `Accounts`, `MarketData`, `Contracts`, `SecurityDefinitions`, and `Trades`.

DTOs are co-located with their capability. Shared
primitives (`AccountID`, `ConID`, `Currency`, etc.) live in `types.go` in the root package.

Dangerous capabilities (trading, banking) will live in **separate Go modules**
(e.g., `gbkr/brokerage/trade`) within the same repository. See ADR-006 and ADR-008.

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
- Don't repeat package name in exported identifiers

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
