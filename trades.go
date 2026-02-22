package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// TradeReader provides read access to recent trade executions (brokerage session required).
// IBKR path prefix: /iserver/account/trades
type TradeReader interface {
	// RecentTrades returns trade executions for up to the last 7 days
	// (GET /iserver/account/trades).
	RecentTrades(ctx context.Context, days int) ([]models.TradeExecution, error)
}

// Trades returns a [TradeReader] for querying recent trade executions.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) Trades() TradeReader {
	return &tradeReader{c: bc.Client}
}

type tradeReader struct {
	c *Client
}

func (t *tradeReader) RecentTrades(ctx context.Context, days int) ([]models.TradeExecution, error) {
	query := url.Values{}
	if days > 0 {
		query.Set("days", fmt.Sprintf("%d", days))
	}
	var result []models.TradeExecution
	if err := t.c.doGet(ctx, "/iserver/account/trades", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
