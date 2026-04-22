package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
	"github.com/trippwill/gbkr/num"
)

// Position represents a single portfolio position returned by
// GET /portfolio/{accountId}/positions/{pageId}.
type Position struct {
	// AcctID is the account identifier.
	AcctID AccountID
	// ConID is the contract identifier.
	ConID ConID
	// ContractDesc is the local symbol / contract description.
	ContractDesc string
	// Qty is the total size of the position. (API: "position")
	Qty num.Num
	// MktPrice is the current market price per share.
	MktPrice num.Num
	// MktValue is the total market value of the position.
	MktValue num.Num
	// AvgCost is the average cost per share multiplied by the contract multiplier.
	AvgCost num.Num
	// AvgPrice is the average purchase price per share.
	AvgPrice num.Num
	// RealizedPnL is the realized profit/loss.
	RealizedPnL num.Num
	// UnrealizedPnL is the unrealized profit/loss.
	UnrealizedPnL num.Num
	// Currency is the traded currency.
	Currency Currency
	// AssetClass is the security type (STK, OPT, etc.).
	AssetClass AssetClass
	// Ticker is the ticker symbol.
	Ticker string
	// Expiry is the contract expiry date (e.g., "20240119"). Empty for non-expiry instruments.
	Expiry string
	// PutOrCall indicates "P" (put) or "C" (call) for options. Empty for non-options.
	PutOrCall string
	// Strike is the option strike price. Absent for non-options.
	Strike num.NullNum
	// UndConID is the underlying contract ID for derivatives. Zero for stocks.
	UndConID ConID
	// Multiplier is the contract multiplier (e.g., 100 for equity options).
	Multiplier num.Num
}

func (p *Position) UnmarshalJSON(data []byte) error {
	raw := struct {
		AcctID        *string     `json:"acctId,omitempty"`
		ConID         *int        `json:"conid,omitempty"`
		ContractDesc  *string     `json:"contractDesc,omitempty"`
		Position      num.Num     `json:"position"`
		MktPrice      num.Num     `json:"mktPrice"`
		MktValue      num.Num     `json:"mktValue"`
		AvgCost       num.Num     `json:"avgCost"`
		AvgPrice      num.Num     `json:"avgPrice"`
		RealizedPnL   num.Num     `json:"realizedPnl"`
		UnrealizedPnL num.Num     `json:"unrealizedPnl"`
		Currency      *string     `json:"currency,omitempty"`
		AssetClass    *string     `json:"assetClass,omitempty"`
		Ticker        *string     `json:"ticker,omitempty"`
		Expiry        *string     `json:"expiry,omitempty"`
		PutOrCall     *string     `json:"putOrCall,omitempty"`
		Strike        num.NullNum `json:"strike"`
		UndConID      *int        `json:"undConid,omitempty"`
		Multiplier    num.Num     `json:"multiplier"`
	}{
		Position:      num.Zero(),
		MktPrice:      num.Zero(),
		MktValue:      num.Zero(),
		AvgCost:       num.Zero(),
		AvgPrice:      num.Zero(),
		RealizedPnL:   num.Zero(),
		UnrealizedPnL: num.Zero(),
		Multiplier:    num.Zero(),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.AcctID = AccountID(jx.Deref(raw.AcctID))
	p.ConID = ConID(jx.Deref(raw.ConID))
	p.ContractDesc = jx.Deref(raw.ContractDesc)
	p.Qty = raw.Position
	p.MktPrice = raw.MktPrice
	p.MktValue = raw.MktValue
	p.AvgCost = raw.AvgCost
	p.AvgPrice = raw.AvgPrice
	p.RealizedPnL = raw.RealizedPnL
	p.UnrealizedPnL = raw.UnrealizedPnL
	p.Currency = Currency(jx.Deref(raw.Currency))
	p.AssetClass = AssetClass(jx.Deref(raw.AssetClass))
	p.Ticker = jx.Deref(raw.Ticker)
	p.Expiry = jx.Deref(raw.Expiry)
	p.PutOrCall = jx.Deref(raw.PutOrCall)
	p.Strike = raw.Strike
	p.UndConID = ConID(jx.Deref(raw.UndConID))
	p.Multiplier = raw.Multiplier
	return nil
}

// PortfolioSummary is the response for GET /portfolio/{accountId}/summary.
// It maps field names to their corresponding summary field objects.
type PortfolioSummary map[string]PortfolioSummaryField

// PortfolioSummaryField represents a single field within a portfolio summary.
type PortfolioSummaryField struct {
	// Amount is the numeric amount of the field.
	Amount num.Num
	// Currency is the currency code for the field value.
	Currency Currency
	// IsNull indicates whether the field value is null.
	IsNull bool
	// Severity is the severity level of the field.
	Severity int
	// Timestamp is the epoch time when the field was last updated.
	Timestamp int64
	// Value is the string representation of the field.
	Value string
}

func (f *PortfolioSummaryField) UnmarshalJSON(data []byte) error {
	raw := struct {
		Amount    num.Num `json:"amount"`
		Currency  *string `json:"currency,omitempty"`
		IsNull    *bool   `json:"isNull,omitempty"`
		Severity  *int    `json:"severity,omitempty"`
		Timestamp *int64  `json:"timestamp,omitempty"`
		Value     *string `json:"value,omitempty"`
	}{
		Amount: num.Zero(),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Amount = raw.Amount
	f.Currency = Currency(jx.Deref(raw.Currency))
	f.IsNull = jx.Deref(raw.IsNull)
	f.Severity = jx.Deref(raw.Severity)
	f.Timestamp = jx.Deref(raw.Timestamp)
	f.Value = jx.Deref(raw.Value)
	return nil
}

// Ledger is the response for GET /portfolio/{accountId}/ledger.
// It maps currency codes to their corresponding ledger entries.
type Ledger map[Currency]LedgerEntry

// LedgerEntry holds ledger data for a single currency.
type LedgerEntry struct {
	// CommodityMarketValue is the market value of commodity positions.
	CommodityMarketValue num.Num
	// FutureOptionValue is the market value of futures options positions.
	FutureOptionValue num.Num
	// FuturesPnL is the PnL of futures positions. (API: "futuresonlypnl")
	FuturesPnL num.Num
	// Interest is the receivable interest balance.
	Interest num.Num
	// NetLiquidation is the net liquidation value of positions. (API: "netliquidationvalue")
	NetLiquidation num.Num
	// RealizedPnL is the realized PnL.
	RealizedPnL num.Num
	// SettledCash is the settled cash balance.
	SettledCash num.Num
	// StockMarketValue is the market value of stock positions.
	StockMarketValue num.Num
	// TotalCashValue is the total cash value.
	TotalCashValue num.Num
	// UnrealizedPnL is the unrealized PnL.
	UnrealizedPnL num.Num
	// Currency is the three-letter currency code.
	Currency Currency
	// Key is always "LedgerList".
	Key string
}

func (e *LedgerEntry) UnmarshalJSON(data []byte) error {
	raw := struct {
		CommodityMarketValue num.Num `json:"commoditymarketvalue"`
		FutureOptionValue    num.Num `json:"futureoptionmarketvalue"`
		FuturesPnL           num.Num `json:"futuresonlypnl"`
		Interest             num.Num `json:"interest"`
		NetLiquidation       num.Num `json:"netliquidationvalue"`
		RealizedPnL          num.Num `json:"realizedpnl"`
		SettledCash          num.Num `json:"settledcash"`
		StockMarketValue     num.Num `json:"stockmarketvalue"`
		TotalCashValue       num.Num `json:"totalcashvalue"`
		UnrealizedPnL        num.Num `json:"unrealizedpnl"`
		Currency             *string `json:"currency,omitempty"`
		Key                  *string `json:"key,omitempty"`
	}{
		CommodityMarketValue: num.Zero(),
		FutureOptionValue:    num.Zero(),
		FuturesPnL:           num.Zero(),
		Interest:             num.Zero(),
		NetLiquidation:       num.Zero(),
		RealizedPnL:          num.Zero(),
		SettledCash:          num.Zero(),
		StockMarketValue:     num.Zero(),
		TotalCashValue:       num.Zero(),
		UnrealizedPnL:        num.Zero(),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.CommodityMarketValue = raw.CommodityMarketValue
	e.FutureOptionValue = raw.FutureOptionValue
	e.FuturesPnL = raw.FuturesPnL
	e.Interest = raw.Interest
	e.NetLiquidation = raw.NetLiquidation
	e.RealizedPnL = raw.RealizedPnL
	e.SettledCash = raw.SettledCash
	e.StockMarketValue = raw.StockMarketValue
	e.TotalCashValue = raw.TotalCashValue
	e.UnrealizedPnL = raw.UnrealizedPnL
	e.Currency = Currency(jx.Deref(raw.Currency))
	e.Key = jx.Deref(raw.Key)
	return nil
}

// Portfolio provides read access to portfolio positions for a specific account.
// IBKR path prefix: /portfolio/{accountId}/*
type Portfolio struct {
	c         *Client
	accountID AccountID
}

// Portfolio returns a [*Portfolio] handle scoped to the given account ID.
func (c *Client) Portfolio(accountID AccountID) *Portfolio {
	return &Portfolio{c: c, accountID: accountID}
}

// ID returns the account this handle is scoped to.
func (p *Portfolio) ID() AccountID {
	return p.accountID
}

// Positions returns positions for a page
// (GET /portfolio/{accountId}/positions/{pageId}).
func (p *Portfolio) Positions(ctx context.Context, page int) ([]Position, error) {
	start := time.Now()
	var result []Position
	path := fmt.Sprintf("/portfolio/%s/positions/%d", url.PathEscape(string(p.accountID)), page)
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioPositions, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Position returns a single position
// (GET /portfolio/{accountId}/position/{conid}).
func (p *Portfolio) Position(ctx context.Context, conID ConID) (*Position, error) {
	start := time.Now()
	var result Position
	path := fmt.Sprintf("/portfolio/%s/position/%d", url.PathEscape(string(p.accountID)), int(conID))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioPosition, err, time.Since(start),
		slog.String("account_id", string(p.accountID)),
		slog.Int64("conid", int64(conID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Summary returns a portfolio summary
// (GET /portfolio/{accountId}/summary).
func (p *Portfolio) Summary(ctx context.Context) (*PortfolioSummary, error) {
	start := time.Now()
	var result PortfolioSummary
	path := fmt.Sprintf("/portfolio/%s/summary", url.PathEscape(string(p.accountID)))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioSummary, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Ledger returns the account ledger
// (GET /portfolio/{accountId}/ledger).
func (p *Portfolio) Ledger(ctx context.Context) (*Ledger, error) {
	start := time.Now()
	var result Ledger
	path := fmt.Sprintf("/portfolio/%s/ledger", url.PathEscape(string(p.accountID)))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioLedger, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Allocation represents the asset allocation for a portfolio account.
type Allocation struct {
	AssetClass map[AssetClass]AllocationEntry
	Sector     map[string]AllocationEntry
	Group      map[string]AllocationEntry
}

// AllocationEntry holds the long and short values for an allocation bucket.
type AllocationEntry struct {
	Long  num.Num
	Short num.Num
}

// UnmarshalJSON implements [json.Unmarshaler].
func (a *Allocation) UnmarshalJSON(data []byte) error {
	var raw struct {
		AssetClass map[string]map[string]num.Num `json:"assetClass"`
		Sector     map[string]map[string]num.Num `json:"sector"`
		Group      map[string]map[string]num.Num `json:"group"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	a.AssetClass = decodeAllocationMap[AssetClass](raw.AssetClass)
	a.Sector = decodeAllocationMap[string](raw.Sector)
	a.Group = decodeAllocationMap[string](raw.Group)
	return nil
}

func decodeAllocationMap[K ~string](m map[string]map[string]num.Num) map[K]AllocationEntry {
	if len(m) == 0 {
		return nil
	}
	result := make(map[K]AllocationEntry, len(m))
	for k, v := range m {
		entry := AllocationEntry{Long: num.Zero(), Short: num.Zero()}
		if n, ok := v["long"]; ok {
			entry.Long = n
		}
		if n, ok := v["short"]; ok {
			entry.Short = n
		}
		result[K(k)] = entry
	}
	return result
}

// Allocation returns the portfolio allocation breakdown
// (GET /portfolio/{accountId}/allocation).
func (p *Portfolio) Allocation(ctx context.Context) (*Allocation, error) {
	start := time.Now()
	var result Allocation
	path := fmt.Sprintf("/portfolio/%s/allocation", url.PathEscape(string(p.accountID)))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioAllocation, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// InvalidatePositions clears the gateway's cached position data
// (POST /portfolio/{accountId}/positions/invalidate).
func (p *Portfolio) InvalidatePositions(ctx context.Context) error {
	start := time.Now()
	path := fmt.Sprintf("/portfolio/%s/positions/invalidate", url.PathEscape(string(p.accountID)))
	err := p.c.doPost(ctx, path, nil, nil)
	p.c.emitOp(ctx, OpPortfolioInvalidatePositions, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	return err
}
