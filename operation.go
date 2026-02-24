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
	OpSessionStatus      Operation = "SessionStatus"
	OpBrokerageSession   Operation = "BrokerageSession"
	OpListAccounts       Operation = "ListAccounts"
	OpAccountPnL         Operation = "AccountPnL"
	OpAccountSummary     Operation = "AccountSummary"
	OpPortfolioPositions Operation = "PortfolioPositions"
	OpPortfolioPosition  Operation = "PortfolioPosition"
	OpPortfolioSummary   Operation = "PortfolioSummary"
	OpPortfolioLedger    Operation = "PortfolioLedger"
	OpContractInfo       Operation = "ContractInfo"
	OpMarketDataSnapshot Operation = "MarketDataSnapshot"
	OpMarketDataHistory  Operation = "MarketDataHistory"
	OpSecuritySearch     Operation = "SecuritySearch"
	OpRecentTrades       Operation = "RecentTrades"
	OpTransactions       Operation = "Transactions"
)

// emitOp logs a structured operation event via the client's slog logger.
// Successful operations emit at LevelInfo; failures emit at LevelWarn.
func (c *Client) emitOp(ctx context.Context, op Operation, err error, dur time.Duration, attrs ...slog.Attr) {
	level := slog.LevelInfo
	if err != nil {
		level = slog.LevelWarn
	}
	allAttrs := make([]slog.Attr, 0, 3+len(attrs))
	allAttrs = append(allAttrs, slog.String("op", string(op)))
	allAttrs = append(allAttrs, slog.Duration("duration", dur))
	if err != nil {
		allAttrs = append(allAttrs, slog.String("error", err.Error()))
	}
	allAttrs = append(allAttrs, attrs...)
	c.logger.LogAttrs(ctx, level, "operation", allAttrs...)
}
