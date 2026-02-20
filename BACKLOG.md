# Backlog

## Gateway

- [ ] Add `gateway.WaitReady(ctx, baseURL)` health/readiness probe helper (polls gateway until accepting connections)
- [ ] Add `WithHealthCheck()` client option that probes the gateway on `NewClient`
- [ ] Add Go helpers for generating gateway `conf.yaml` from code (`gateway.Config` struct → YAML)
- [ ] Improve `gateway/README.md` with clearer end-user setup guide (prerequisites, auth flow, troubleshooting common issues)

## Architecture

- [ ] Move dangerous capabilities (orders, banking) into separate subpackages (`gbkr/trading`, `gbkr/banking`) so they require explicit imports and can be audited with `go list -deps` / `go version -m`
- [ ] Consider build tags as additional hardening on top of package separation (e.g., `//go:build gbkr_trading`)
- [ ] Explore auto-enforcement of IBKR request rate limits (per-endpoint throttling, backoff, quota tracking in the Client HTTP layer)

## API Coverage

- [ ] Add Orders area (PlaceOrder, ModifyOrder, CancelOrder) — in `gbkr/trading` subpackage
- [ ] Add Alerts area (create, modify, delete alerts)
- [ ] Add Contracts area (search, details, trading rules)
- [ ] Add Scanner area (market scanner parameters and results)
- [ ] Add Watchlists area (create, modify, delete watchlists)
- [ ] Add Banking/Transfers area — in `gbkr/banking` subpackage

## Infrastructure

- [ ] Tag `v0.1.0`
- [ ] Add `go install` section to README
- [ ] Register on pkg.go.dev
- [x] Add Go doc examples for capability constructors
