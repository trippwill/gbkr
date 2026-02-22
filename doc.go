// Package gbkr provides a permission-gated client for the IBKR REST API.
//
// # Two-Phase Session Model
//
// The package uses a two-tier client model that mirrors the IBKR gateway's
// session lifecycle:
//
//  1. [Client] — created via [NewClient]. Provides ungated gateway capabilities:
//     [Client.SessionStatus], [Client.Portfolio], and [Client.Analysis].
//
//  2. [BrokerageClient] — obtained by calling [Client.BrokerageSession], which
//     performs an SSO/DH handshake to elevate to a full brokerage session.
//     Provides brokerage capabilities: [BrokerageClient.Accounts],
//     [BrokerageClient.Account], [BrokerageClient.MarketData],
//     [BrokerageClient.Contracts], [BrokerageClient.SecurityDefinitions],
//     and [BrokerageClient.Trades].
//     Because [BrokerageClient] embeds [*Client], all gateway capabilities
//     remain available after elevation.
//
// # Interface-to-Path Mapping
//
//	Interface                  Access Point         IBKR Path Prefix
//	─────────────────────────  ──────────────────   ─────────────────────────
//	PortfolioReader            Client.Portfolio()   /portfolio/{accountId}/*
//	AnalysisReader             Client.Analysis()    /pa/*
//	AccountLister              BrokerageClient      /iserver/accounts
//	AccountReader              BrokerageClient      /iserver/account/{id}/*
//	MarketDataReader           BrokerageClient      /iserver/marketdata/*
//	ContractReader             BrokerageClient      /iserver/contract/{conid}/*
//	SecurityDefinitionReader   BrokerageClient      /iserver/secdef/*
//	TradeReader                BrokerageClient      /iserver/account/trades
//
// # Permission Model
//
// A two-tier permission model (Scope:Level) gates brokerage session elevation.
// Consumers grant permissions via [WithPermissions], [WithPermissionsFromFile],
// or JIT prompting with [WithInteractivePrompt].
// Predefined sets [ReadOnly] and [FullAccess] cover common scenarios.
//
// Gateway access ([Client.Portfolio], [Client.Analysis], [Client.SessionStatus])
// requires no permissions. Only [Client.BrokerageSession] checks permissions.
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
