package gbkr

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/trippwill/gbkr/models"
)

// TradeReader provides read access to recent trade executions (brokerage session required).
type TradeReader interface {
	// RecentTrades returns trade executions for up to the last 7 days
	// (GET /iserver/account/trades).
	RecentTrades(ctx context.Context, days int) ([]models.TradeExecution, error)
}

// TransactionReader provides read-only access to Portfolio Analyst
// transaction history (POST /pa/transactions).
type TransactionReader interface {
	TransactionHistory(ctx context.Context, accountID models.AccountID, conID models.ConID, days int) (*Cached[models.TransactionHistoryResponse], error)
}

var requiredTradePermissions = []Permission{
	{AreaTrading, ResourceTrades, ActionRead},
}

// Trades returns a [TradeReader] for querying recent trade executions.
// Requires: trading.trades.read.
func (bc *BrokerageClient) Trades() (TradeReader, error) {
	if err := checkPermissions(bc.Client, requiredTradePermissions...); err != nil {
		return nil, err
	}
	return &tradeReader{c: bc.Client}, nil
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

// TransactionHistory returns a [TransactionReader] for querying Portfolio Analyst
// transaction history. Requires: trading.trades.read.
func (c *Client) TransactionHistory() (TransactionReader, error) {
	if err := checkPermissions(c, Permission{AreaTrading, ResourceTrades, ActionRead}); err != nil {
		return nil, err
	}
	var obs PacingObserver
	if c.pacing != nil {
		obs = c.pacing.observer
	}
	return &transactionReader{
		c: c,
		txCache: &ttlCache[models.TransactionHistoryResponse]{
			ttl:      15 * time.Minute,
			observer: obs,
			path:     "/pa/transactions",
		},
	}, nil
}

type transactionReader struct {
	c       *Client
	txCache *ttlCache[models.TransactionHistoryResponse]
}

func (t *transactionReader) TransactionHistory(ctx context.Context, accountID models.AccountID, conID models.ConID, days int) (*Cached[models.TransactionHistoryResponse], error) {
	if cached := t.txCache.get(); cached != nil {
		return cached, nil
	}
	req := models.TransactionHistoryRequest{
		AcctIDs:  []string{string(accountID)},
		ConIDs:   []int{int(conID)},
		Currency: "USD",
	}
	if days > 0 {
		req.Days = fmt.Sprintf("%d", days)
	}
	var result models.TransactionHistoryResponse
	if err := t.c.doPost(ctx, "/pa/transactions", req, &result); err != nil {
		t.txCache.invalidate()
		return nil, err
	}
	return t.txCache.set(result), nil
}
