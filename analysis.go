package gbkr

import (
	"context"
	"fmt"
	"time"

	"github.com/trippwill/gbkr/models"
)

// Analysis provides read-only access to Portfolio Analyst data.
// IBKR path prefix: /pa/*
type Analysis struct {
	c       *Client
	txCache *ttlCache[models.TransactionHistoryResponse]
}

// Analysis returns an [*Analysis] handle for querying Portfolio Analyst data.
//
// The returned handle caches results for 15 minutes (matching the IBKR pacing
// limit). Callers should retain and reuse the handle rather than calling
// Analysis() repeatedly.
func (c *Client) Analysis() *Analysis {
	var obs PacingObserver
	if c.pacing != nil {
		obs = c.pacing.observer
	}
	return &Analysis{
		c: c,
		txCache: &ttlCache[models.TransactionHistoryResponse]{
			ttl:      15 * time.Minute,
			observer: obs,
			path:     "/pa/transactions",
		},
	}
}

// Transactions returns transaction history for an account and contract
// (POST /pa/transactions).
func (a *Analysis) Transactions(ctx context.Context, accountID models.AccountID, conID models.ConID, days int) (*Cached[models.TransactionHistoryResponse], error) {
	key := fmt.Sprintf("%s:%d:%d", accountID, conID, days)
	if cached := a.txCache.get(key); cached != nil {
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
	if err := a.c.doPost(ctx, "/pa/transactions", req, &result); err != nil {
		a.txCache.invalidate()
		return nil, err
	}
	return a.txCache.set(key, result), nil
}
