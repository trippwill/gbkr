package gbkr

import (
	"context"
	"fmt"
	"time"

	"github.com/trippwill/gbkr/models"
)

// AnalysisReader provides read-only access to Portfolio Analyst data.
// IBKR path prefix: /pa/*
//
// Gateway access — no permission check required.
type AnalysisReader interface {
	// Transactions returns transaction history for an account and contract
	// (POST /pa/transactions).
	Transactions(ctx context.Context, accountID models.AccountID, conID models.ConID, days int) (*Cached[models.TransactionHistoryResponse], error)
}

// Analysis returns an [AnalysisReader] for querying Portfolio Analyst data.
// Gateway access — no permission check required.
//
// The returned reader caches results for 15 minutes (matching the IBKR pacing
// limit). Callers should retain and reuse the reader rather than calling
// Analysis() repeatedly.
func (c *Client) Analysis() AnalysisReader {
	var obs PacingObserver
	if c.pacing != nil {
		obs = c.pacing.observer
	}
	return &analysisReader{
		c: c,
		txCache: &ttlCache[models.TransactionHistoryResponse]{
			ttl:      15 * time.Minute,
			observer: obs,
			path:     "/pa/transactions",
		},
	}
}

type analysisReader struct {
	c       *Client
	txCache *ttlCache[models.TransactionHistoryResponse]
}

func (a *analysisReader) Transactions(ctx context.Context, accountID models.AccountID, conID models.ConID, days int) (*Cached[models.TransactionHistoryResponse], error) {
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
