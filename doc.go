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
package gbkr
