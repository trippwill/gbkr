// Package gbkr provides a Go client for the IBKR REST API.
//
// # Two-Tier Session Model
//
// The package uses a two-tier client model that mirrors the IBKR gateway's
// session lifecycle:
//
//  1. [Client] — created via [NewClient]. Provides gateway capabilities:
//     [Client.SessionStatus], [Client.Portfolio], and [Client.Analysis].
//
//  2. [brokerage.Client] — obtained via [brokerage.NewSession], which performs
//     an SSO/DH handshake to elevate to a full brokerage session.
//     Provides brokerage capabilities: Accounts, MarketData, Contracts,
//     SecurityDefinitions, and Trades.
//     See the [github.com/trippwill/gbkr/brokerage] package for details.
//
// # Capability-to-Path Mapping
//
//	Capability                 Access Point                          IBKR Path Prefix
//	─────────────────────────  ─────────────────────────────────     ─────────────────────────
//	Portfolio                  Client.Portfolio()                      /portfolio/{accountId}/*
//	Analysis                   Client.Analysis()                       /pa/*
//	Accounts                   brokerage.Client.Accounts()             /iserver/accounts
//	MarketData                 brokerage.Client.MarketData()           /iserver/marketdata/*
//	Contracts                  brokerage.Client.Contracts()            /iserver/contract/{conid}/*
//	SecurityDefinitions        brokerage.Client.SecurityDefinitions()  /iserver/secdef/*
//	Trades                     brokerage.Client.Trades()               /iserver/account/trades
//
// # API Pacing
//
// By default, the client enforces the IBKR pacing limits automatically.
// Every outbound request passes through [PacingPolicy], which applies:
//
//   - A global ceiling of 10 requests/second
//   - Per-endpoint rate limits for sensitive paths (e.g., 1 req/5 s for
//     /iserver/account/orders)
//   - Concurrency semaphores where documented (e.g., 5 concurrent requests
//     for /iserver/marketdata/history)
//
// If a request would exceed the limit, the client blocks until a slot is
// available or the context is cancelled (returning [ErrPacingWait]).
//
// Pacing can be disabled for testing with WithRateLimit(nil) or replaced
// with a custom [PacingPolicy] via [WithRateLimit]. An optional
// [PacingObserver] (set via [WithPacingObserver]) receives notifications
// about pacing waits and cache events.
//
// # Data Freshness
//
// Endpoints with very long pacing intervals (e.g., /pa/transactions at
// 1 req/15 min) return [Cached] values. The [Cached.FetchedAt] timestamp
// lets consumers decide whether the data is fresh enough for their use case.
// On cache refresh failure, stale data is discarded (fail-closed).
package gbkr
