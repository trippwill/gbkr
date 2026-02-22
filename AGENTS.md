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
gbkr_err.go          # Error type, sentinel errors, APIError, PermissionDeniedError
options.go           # Functional options (WithBaseURL, WithHTTPClient, etc.)
permission.go        # Area/Resource/Action enums, PermissionSet, enforcement
permission_err.go    # Parse error sentinels and constructors
permcfg.go           # YAML config loading, interactive permission prompt
permcfg_err.go       # Config error sentinels and constructors
permcfg_grant_all.go # GrantAllPrompter (build tag: gbkr_grant_all)
auth.go              # SessionClient interface + Session(c) constructor
accounts.go          # AccountLister + AccountReader interfaces + constructors
positions.go         # PositionReader interface + Positions(c, id) constructor
marketdata.go        # MarketDataReader interface + MarketData(c) constructor
trades.go            # TradeReader interface + BrokerageClient.Trades() method
contracts.go         # ContractReader interface + Contracts(c) constructor
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

### Permission Model

Three-tier enums (`Area`, `Resource`, `Action`) using `iota+1`. Zero-value (0) acts as wildcard in `Permission.Allows()`.

Capability constructors (`Session`, `Accounts`, `Account`, `Positions`, `MarketData`) check permissions at runtime and return narrow interfaces. Consumers only see methods on the interface they obtained.

Default permissions are **empty** (fail-closed). Consumers must explicitly grant permissions via `WithPermissions()`, `WithPermissionsFromFile()`, or enable JIT prompting with `WithInteractivePrompt()`.

A `Prompter` interface supports JIT permission grants — when a capability constructor needs missing permissions, the prompter asks the user (or auto-grants in testing). Static permissions serve as a floor; JIT can only add, never remove.

### Dangerous Capabilities

Trading and banking capabilities will live in separate subpackages (`gbkr/trading`, `gbkr/banking`) so they require explicit imports. See `BACKLOG.md` for details.

## Conventions

### Code Style
- Format with `gofmt` (run `mise run fmt` before commits)
- Exported identifiers use `CamelCase`; unexported use `lowerCamel`
- File names are lowercase with underscores where needed
- Package documentation goes in `doc.go`
- Error definitions go in dedicated `*_err.go` files

### Error Handling
- Define `type Error string` as the base sentinel type in `gbkr_err.go`
- Use `const` sentinels for errors callers check with `errors.Is` (e.g., `ErrPermissionDenied`)
- Use struct types with `Unwrap()` for errors needing context (e.g., `ParseError`, `APIError`)
- Use `Err*()` constructor functions for ergonomic error creation
- Use `fmt.Errorf("context: %w", err)` only for internal wrapping not inspected by callers
- Callers check errors via `errors.Is(err, gbkr.ErrPermissionDenied)` or `errors.As(err, &apiErr)`

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
- `GrantAllPrompter` is gated by `//go:build gbkr_grant_all` — only include in test builds
