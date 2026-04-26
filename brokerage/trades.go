package brokerage

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/internal/jx"
	"github.com/trippwill/gbkr/num"
)

// Trades provides read access to recent trade executions (brokerage session required).
// IBKR path prefix: /iserver/account/trades
type Trades struct {
	c *Client
}

// Trades returns a [*Trades] handle for querying recent trade executions.
func (c *Client) Trades() *Trades {
	return &Trades{c: c}
}

// Recent returns trade executions for up to the last 7 days
// (GET /iserver/account/trades).
func (t *Trades) Recent(ctx context.Context, days int) ([]TradeExecution, error) {
	start := time.Now()
	query := url.Values{}
	if days > 0 {
		query.Set("days", fmt.Sprintf("%d", days))
	}
	var result []TradeExecution
	err := t.c.doGet(ctx, "/iserver/account/trades", query, &result)
	t.c.emitOp(ctx, gbkr.OpRecentTrades, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// TradeExecution represents a single trade execution returned by
// GET /iserver/account/trades.
type TradeExecution struct {
	// ExecutionID is the unique execution identifier.
	ExecutionID string
	// Symbol is the trading symbol.
	Symbol string
	// Side is the order side (e.g., "B" for buy, "S" for sell).
	Side string
	// OrderDescription is a human-readable order summary.
	OrderDescription string
	// TradeTime is the UTC trade time (format: "20231211-18:00:49").
	TradeTime string
	// TradeTimeEpoch is the epoch millisecond timestamp of the trade.
	TradeTimeEpoch int64
	// Quantity is the number of units traded.
	Quantity num.Num
	// Price is the execution price.
	Price num.Num
	// OrderRef is the user-defined order reference.
	OrderRef string
	// Exchange is the execution exchange.
	Exchange gbkr.Exchange
	// Commission is the trade commission.
	Commission num.Num
	// NetAmount is the total net cost of the trade.
	NetAmount num.Num
	// Account is the account identifier.
	Account gbkr.AccountID
	// CompanyName is the long company name.
	CompanyName string
	// ContractDesc is the local symbol / contract description.
	ContractDesc string
	// SecType is the security type (e.g., "STK", "OPT").
	SecType gbkr.AssetClass
	// ListingExchange is the primary listing exchange.
	ListingExchange gbkr.Exchange
	// ConID is the contract identifier.
	ConID gbkr.ConID
}

func (t *TradeExecution) UnmarshalJSON(data []byte) error {
	raw := struct {
		ExecutionID      *string `json:"execution_id,omitempty"`
		Symbol           *string `json:"symbol,omitempty"`
		Side             *string `json:"side,omitempty"`
		OrderDescription *string `json:"order_description,omitempty"`
		TradeTime        *string `json:"trade_time,omitempty"`
		TradeTimeEpoch   *int64  `json:"trade_time_r,omitempty"`
		Quantity         num.Num `json:"size"`
		Price            num.Num `json:"price"`
		OrderRef         *string `json:"order_ref,omitempty"`
		Exchange         *string `json:"exchange,omitempty"`
		Commission       num.Num `json:"commission"`
		NetAmount        num.Num `json:"net_amount"`
		Account          *string `json:"account,omitempty"`
		CompanyName      *string `json:"company_name,omitempty"`
		ContractDesc     *string `json:"contract_description_1,omitempty"`
		SecType          *string `json:"sec_type,omitempty"`
		ListingExchange  *string `json:"listing_exchange,omitempty"`
		ConID            *int    `json:"conid,omitempty"`
	}{
		Quantity:   num.Zero(),
		Price:      num.Zero(),
		Commission: num.Zero(),
		NetAmount:  num.Zero(),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t.ExecutionID = jx.Deref(raw.ExecutionID)
	t.Symbol = jx.Deref(raw.Symbol)
	t.Side = jx.Deref(raw.Side)
	t.OrderDescription = jx.Deref(raw.OrderDescription)
	t.TradeTime = jx.Deref(raw.TradeTime)
	t.TradeTimeEpoch = jx.Deref(raw.TradeTimeEpoch)
	t.Quantity = raw.Quantity
	t.Price = raw.Price
	t.OrderRef = jx.Deref(raw.OrderRef)
	t.Exchange = gbkr.Exchange(jx.Deref(raw.Exchange))
	t.Commission = raw.Commission
	t.NetAmount = raw.NetAmount
	t.Account = gbkr.AccountID(jx.Deref(raw.Account))
	t.CompanyName = jx.Deref(raw.CompanyName)
	t.ContractDesc = jx.Deref(raw.ContractDesc)
	t.SecType = gbkr.AssetClass(jx.Deref(raw.SecType))
	t.ListingExchange = gbkr.Exchange(jx.Deref(raw.ListingExchange))
	t.ConID = gbkr.ConID(jx.Deref(raw.ConID))
	return nil
}
