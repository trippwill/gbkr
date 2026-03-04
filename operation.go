package gbkr

import (
	"context"
	"log/slog"
	"time"
)

// Operation identifies an IBKR API operation for event sourcing.
// String values are wire-stable and must not be renamed.
type Operation string

const (
	OpSessionStatus                Operation = "SessionStatus"
	OpLogout                       Operation = "Logout"
	OpReauthenticate               Operation = "Reauthenticate"
	OpTickle                       Operation = "Tickle"
	OpBrokerageSession             Operation = "BrokerageSession"
	OpListAccounts                 Operation = "ListAccounts"
	OpAccountPnL                   Operation = "AccountPnL"
	OpAccountSummary               Operation = "AccountSummary"
	OpPortfolioPositions           Operation = "PortfolioPositions"
	OpPortfolioPosition            Operation = "PortfolioPosition"
	OpPortfolioSummary             Operation = "PortfolioSummary"
	OpPortfolioLedger              Operation = "PortfolioLedger"
	OpContractInfo                 Operation = "ContractInfo"
	OpMarketDataSnapshot           Operation = "MarketDataSnapshot"
	OpMarketDataHistory            Operation = "MarketDataHistory"
	OpSecuritySearch               Operation = "SecuritySearch"
	OpRecentTrades                 Operation = "RecentTrades"
	OpTransactions                 Operation = "Transactions"
	OpPortfolioAccounts            Operation = "PortfolioAccounts"
	OpPortfolioAllocation          Operation = "PortfolioAllocation"
	OpPortfolioInvalidatePositions Operation = "PortfolioInvalidatePositions"
	OpTradingSchedule              Operation = "TradingSchedule"
)

// emitOp logs a structured operation event via the client's slog logger.
// Successful operations emit at LevelInfo; failures emit at LevelWarn.
func (c *Client) emitOp(ctx context.Context, op Operation, err error, dur time.Duration, attrs ...slog.Attr) {
	c.t.EmitOp(ctx, string(op), err, dur, attrs...)
}
