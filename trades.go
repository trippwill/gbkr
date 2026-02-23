package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// Trades provides read access to recent trade executions (brokerage session required).
// IBKR path prefix: /iserver/account/trades
type Trades struct {
	c *Client
}

// Trades returns a [*Trades] handle for querying recent trade executions.
func (bc *BrokerageClient) Trades() *Trades {
	return &Trades{c: bc.Client}
}

// Recent returns trade executions for up to the last 7 days
// (GET /iserver/account/trades).
func (t *Trades) Recent(ctx context.Context, days int) ([]models.TradeExecution, error) {
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
