package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
	"github.com/trippwill/gbkr/num"
	"github.com/trippwill/gbkr/when"
)

// TransactionHistoryRequest is the request body for POST /pa/transactions.
type TransactionHistoryRequest struct {
	AcctIDs  []string `json:"acctIds"`
	ConIDs   []int    `json:"conids"`
	Currency string   `json:"currency"`
	Days     string   `json:"days,omitempty"`
}

// TransactionHistoryResponse is the response from POST /pa/transactions.
type TransactionHistoryResponse struct {
	// Currency is the reporting currency.
	Currency Currency
	// From is the start time of the range.
	From when.DateTime
	// To is the end time of the range.
	To when.DateTime
	// IncludesRealTime indicates if trades are up to date.
	IncludesRealTime bool
	// Transactions is the list of transactions.
	Transactions []Transaction
}

func (r *TransactionHistoryResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Currency         *string          `json:"currency,omitempty"`
		From             *int64           `json:"from,omitempty"`
		To               *int64           `json:"to,omitempty"`
		IncludesRealTime *bool            `json:"includesRealTime,omitempty"`
		Transactions     *json.RawMessage `json:"transactions,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Currency = Currency(jx.Deref(raw.Currency))
	if raw.From != nil {
		r.From = when.DateTimeFromEpoch(*raw.From)
	}
	if raw.To != nil {
		r.To = when.DateTimeFromEpoch(*raw.To)
	}
	r.IncludesRealTime = jx.Deref(raw.IncludesRealTime)
	if raw.Transactions != nil {
		if err := json.Unmarshal(*raw.Transactions, &r.Transactions); err != nil {
			return err
		}
	}
	return nil
}

// Transaction represents a single transaction from POST /pa/transactions.
type Transaction struct {
	// Date is the datetime of the transaction.
	Date when.DateTime
	// Currency is the traded instrument currency.
	Currency Currency
	// FxRate is the forex conversion rate.
	FxRate num.Num
	// Price is the price per share.
	Price num.Num
	// Qty is the quantity traded (negative for sells).
	Qty int
	// Account is the account identifier.
	Account AccountID
	// Amount is the total trade value.
	Amount num.Num
	// ConID is the contract identifier.
	ConID ConID
	// Type is the order side (e.g., "Sell", "Buy").
	Type string
	// Desc is the long company name.
	Desc string
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	raw := struct {
		Date     *string `json:"date,omitempty"`
		Currency *string `json:"cur,omitempty"`
		FxRate   num.Num `json:"fxRate"`
		Price    num.Num `json:"pr"`
		Qty      *int    `json:"qty,omitempty"`
		Account  *string `json:"acctid,omitempty"`
		Amount   num.Num `json:"amt"`
		ConID    *int    `json:"conid,omitempty"`
		Type     *string `json:"type,omitempty"`
		Desc     *string `json:"desc,omitempty"`
	}{
		FxRate: num.Zero(),
		Price:  num.Zero(),
		Amount: num.Zero(),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	dateStr := jx.Deref(raw.Date)
	if dateStr != "" {
		dt, err := when.ParseDateTime(dateStr)
		if err != nil {
			return fmt.Errorf("parse transaction date %q: %w", dateStr, err)
		}
		t.Date = dt
	}
	t.Currency = Currency(jx.Deref(raw.Currency))
	t.FxRate = raw.FxRate
	t.Price = raw.Price
	t.Qty = jx.Deref(raw.Qty)
	t.Account = AccountID(jx.Deref(raw.Account))
	t.Amount = raw.Amount
	t.ConID = ConID(jx.Deref(raw.ConID))
	t.Type = jx.Deref(raw.Type)
	t.Desc = jx.Deref(raw.Desc)
	return nil
}

// Analysis provides read-only access to Portfolio Analyst data.
// IBKR path prefix: /pa/*
type Analysis struct {
	c       *Client
	txCache *ttlCache[TransactionHistoryResponse]
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
		txCache: &ttlCache[TransactionHistoryResponse]{
			ttl:      15 * time.Minute,
			observer: obs,
			path:     "/pa/transactions",
		},
	}
}

// Transactions returns transaction history for an account and contract
// (POST /pa/transactions).
func (a *Analysis) Transactions(ctx context.Context, accountID AccountID, conID ConID, days int) (*Cached[TransactionHistoryResponse], error) {
	key := fmt.Sprintf("%s:%d:%d", accountID, conID, days)
	if cached := a.txCache.get(key); cached != nil {
		return cached, nil
	}
	start := time.Now()
	req := TransactionHistoryRequest{
		AcctIDs:  []string{string(accountID)},
		ConIDs:   []int{int(conID)},
		Currency: "USD",
	}
	if days > 0 {
		req.Days = fmt.Sprintf("%d", days)
	}
	var result TransactionHistoryResponse
	err := a.c.doPost(ctx, "/pa/transactions", req, &result)
	a.c.emitOp(ctx, OpTransactions, err, time.Since(start),
		slog.String("account_id", string(accountID)),
		slog.Int64("conid", int64(conID)),
		slog.Int("count", len(result.Transactions)))
	if err != nil {
		a.txCache.invalidate()
		return nil, err
	}
	return a.txCache.set(key, result), nil
}
