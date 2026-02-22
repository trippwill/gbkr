// Package gbkr provides a permission-gated client for the IBKR REST API.
//
// # Two-Phase Session Model
//
// The package uses a two-tier client model that mirrors the IBKR gateway's
// session lifecycle:
//
//  1. [Client] — created via [NewClient]. Provides read-only capabilities that
//     work with the gateway's default session: [Client.SessionStatus],
//     [Client.Positions], and [Client.TransactionHistory].
//
//  2. [BrokerageClient] — obtained by calling [Client.BrokerageSession], which
//     performs an SSO/DH handshake to elevate to a full brokerage session.
//     Provides brokerage capabilities: [BrokerageClient.Accounts],
//     [BrokerageClient.Account], [BrokerageClient.MarketData],
//     [BrokerageClient.Contracts], and [BrokerageClient.Trades].
//     Because [BrokerageClient] embeds [*Client], all read-only capabilities
//     remain available after elevation.
//
// # Permission Model
//
// A three-tier permission model (Area / Resource / Action) gates every
// capability at runtime. Consumers grant permissions via [WithPermissions],
// [WithPermissionsFromFile], or JIT prompting with [WithInteractivePrompt].
// Predefined sets [ReadOnly], [ReadOnlyAuth], and [AllPermissions] cover
// common scenarios.
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
// Pacing can be disabled for testing with [WithRateLimit](nil) or replaced
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
