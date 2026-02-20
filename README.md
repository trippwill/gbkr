# gbkr

A permission-gated Go client library for the [Interactive Brokers](https://www.interactivebrokers.com/) Client Portal Gateway REST API.

## Features

- **Narrow capability interfaces** — consumers only see the methods they're allowed to use (`SessionClient`, `AccountLister`, `AccountReader`, `PositionReader`, `MarketDataReader`)
- **Three-tier permission model** — `Area.Resource.Action` enums with wildcard support, enforced at both compile time (interface types) and runtime (constructor checks)
- **Fail-closed by default** — no permissions are granted unless explicitly configured
- **Strongly-typed domain aliases** — `AccountID`, `ConID`, `Currency`, `BarSize`, `TimePeriod` prevent parameter confusion
- **Flexible permission sources** — static sets, YAML config files, or interactive prompts
- **Structured error model** — const sentinel errors with `errors.Is`/`errors.As` support and `Err*()` constructors for context-rich errors

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
)

func main() {
    client, err := gbkr.NewClient(
        gbkr.WithBaseURL("https://localhost:5000/v1/api"),
        gbkr.WithInsecureTLS(),
        gbkr.WithPermissions(gbkr.ReadOnlyAuth),
    )
    if err != nil {
        log.Fatal(err)
    }

    lister, err := gbkr.Accounts(client)
    if err != nil {
        log.Fatal(err) // permission denied
    }

    accounts, err := lister.ListAccounts(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(accounts)
}
```

## Permissions

Permissions use three-tier enums (`Area`, `Resource`, `Action`). A zero value in any field acts as a wildcard.

```go
// Grant all read permissions + auth login
gbkr.WithPermissions(gbkr.ReadOnlyAuth)

// Grant specific permissions
gbkr.WithPermissions(gbkr.PermissionSet{
    {gbkr.AreaAuth, gbkr.ResourceSession, gbkr.ActionRead},
    {gbkr.AreaTrading, gbkr.ResourceAccounts, gbkr.ActionRead},
})

// Load from a YAML file
gbkr.WithPermissionsFromFile("permissions.yaml")

// JIT prompt — permissions are requested as needed, not upfront
gbkr.WithInteractivePrompt()
```

### JIT Permission Prompting

Instead of granting all permissions upfront, use `WithInteractivePrompt()` to prompt the user for each permission as it's needed:

```go
client, _ := gbkr.NewClient(
    gbkr.WithBaseURL("https://localhost:5000/v1/api"),
    gbkr.WithInteractivePrompt(),
)

// When Session(client) is called, the user sees:
//   Grant auth.session.read? [y/N] y
//   Grant auth.session.write? [y/N] y
sess, _ := gbkr.Session(client)
```

A permissions file can serve as a floor — JIT prompts only for anything missing:

```go
gbkr.WithPermissionsFromFile("perms.yaml"),  // pre-grant some
gbkr.WithInteractivePrompt(),                 // prompt for the rest
```

YAML format:

```yaml
permissions:
  - area: auth
    resource: session
    action: read
  - area: trading
    resource: "*"
    action: read
```

## Error Handling

All domain errors are inspectable via standard Go error patterns:

```go
import "errors"

// Check sentinel errors
if errors.Is(err, gbkr.ErrPermissionDenied) {
    // insufficient permissions
}

// Extract structured context
var apiErr *gbkr.APIError
if errors.As(err, &apiErr) {
    fmt.Println(apiErr.StatusCode, apiErr.Status)
}

var parseErr *gbkr.ParseError
if errors.As(err, &parseErr) {
    fmt.Println(parseErr.Kind, parseErr.Value) // e.g., "unknown area", "badvalue"
}
```

## CLI

A test CLI is included for exercising the library:

```bash
go run ./cmd/gbkr --base-url https://localhost:5000/v1/api --insecure
# Prompts for each permission as needed (JIT)

go run ./cmd/gbkr --permissions-file cmd/gbkr/examples/readonly.yaml --insecure
# Uses file as floor; prompts only for anything missing
```

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
