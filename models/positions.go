package models

import "encoding/json"

// Position represents a single portfolio position.
type Position struct {
	AcctID        AccountID
	ConID         ConID
	ContractDesc  string
	Qty           float64
	MktPrice      float64
	MktValue      float64
	AvgCost       float64
	AvgPrice      float64
	RealizedPnL   float64
	UnrealizedPnL float64
	Currency      Currency
	AssetClass    AssetClass
	Ticker        string
}

func (p *Position) UnmarshalJSON(data []byte) error {
	var raw struct {
		AcctID        *string  `json:"acctId,omitempty"`
		ConID         *int     `json:"conid,omitempty"`
		ContractDesc  *string  `json:"contractDesc,omitempty"`
		Position      *float64 `json:"position,omitempty"`
		MktPrice      *float64 `json:"mktPrice,omitempty"`
		MktValue      *float64 `json:"mktValue,omitempty"`
		AvgCost       *float64 `json:"avgCost,omitempty"`
		AvgPrice      *float64 `json:"avgPrice,omitempty"`
		RealizedPnL   *float64 `json:"realizedPnl,omitempty"`
		UnrealizedPnL *float64 `json:"unrealizedPnl,omitempty"`
		Currency      *string  `json:"currency,omitempty"`
		AssetClass    *string  `json:"assetClass,omitempty"`
		Ticker        *string  `json:"ticker,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	p.AcctID = AccountID(deref(raw.AcctID))
	p.ConID = ConID(deref(raw.ConID))
	p.ContractDesc = deref(raw.ContractDesc)
	p.Qty = deref(raw.Position)
	p.MktPrice = deref(raw.MktPrice)
	p.MktValue = deref(raw.MktValue)
	p.AvgCost = deref(raw.AvgCost)
	p.AvgPrice = deref(raw.AvgPrice)
	p.RealizedPnL = deref(raw.RealizedPnL)
	p.UnrealizedPnL = deref(raw.UnrealizedPnL)
	p.Currency = Currency(deref(raw.Currency))
	p.AssetClass = AssetClass(deref(raw.AssetClass))
	p.Ticker = deref(raw.Ticker)
	return nil
}

// PortfolioSummary is the response for GET /portfolio/{accountId}/summary.
type PortfolioSummary map[string]PortfolioSummaryField

// PortfolioSummaryField represents a single field within a portfolio summary.
type PortfolioSummaryField struct {
	Amount    float64
	Currency  Currency
	IsNull    bool
	Severity  int
	Timestamp int64
	Value     string
}

func (f *PortfolioSummaryField) UnmarshalJSON(data []byte) error {
	var raw struct {
		Amount    *float64 `json:"amount,omitempty"`
		Currency  *string  `json:"currency,omitempty"`
		IsNull    *bool    `json:"isNull,omitempty"`
		Severity  *int     `json:"severity,omitempty"`
		Timestamp *int64   `json:"timestamp,omitempty"`
		Value     *string  `json:"value,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Amount = deref(raw.Amount)
	f.Currency = Currency(deref(raw.Currency))
	f.IsNull = deref(raw.IsNull)
	f.Severity = deref(raw.Severity)
	f.Timestamp = deref(raw.Timestamp)
	f.Value = deref(raw.Value)
	return nil
}

// Ledger is the response for GET /portfolio/{accountId}/ledger.
type Ledger map[Currency]LedgerEntry

// LedgerEntry holds ledger data for a single currency.
type LedgerEntry struct {
	CommodityMarketValue float64
	FutureOptionValue    float64
	FuturesPnL           float64
	Interest             float64
	NetLiquidation       float64
	RealizedPnL          float64
	SettledCash          float64
	StockMarketValue     float64
	TotalCashValue       float64
	UnrealizedPnL        float64
	Currency             Currency
	Key                  string
}

func (e *LedgerEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		CommodityMarketValue *float64 `json:"commoditymarketvalue,omitempty"`
		FutureOptionValue    *float64 `json:"futureoptionvalue,omitempty"`
		FuturesPnL           *float64 `json:"futuresnlvalue,omitempty"`
		Interest             *float64 `json:"interest,omitempty"`
		NetLiquidation       *float64 `json:"netliquidationvalue,omitempty"`
		RealizedPnL          *float64 `json:"realizedpnl,omitempty"`
		SettledCash          *float64 `json:"settledcash,omitempty"`
		StockMarketValue     *float64 `json:"stockmarketvalue,omitempty"`
		TotalCashValue       *float64 `json:"totalcashvalue,omitempty"`
		UnrealizedPnL        *float64 `json:"unrealizedpnl,omitempty"`
		Currency             *string  `json:"currency,omitempty"`
		Key                  *string  `json:"key,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.CommodityMarketValue = deref(raw.CommodityMarketValue)
	e.FutureOptionValue = deref(raw.FutureOptionValue)
	e.FuturesPnL = deref(raw.FuturesPnL)
	e.Interest = deref(raw.Interest)
	e.NetLiquidation = deref(raw.NetLiquidation)
	e.RealizedPnL = deref(raw.RealizedPnL)
	e.SettledCash = deref(raw.SettledCash)
	e.StockMarketValue = deref(raw.StockMarketValue)
	e.TotalCashValue = deref(raw.TotalCashValue)
	e.UnrealizedPnL = deref(raw.UnrealizedPnL)
	e.Currency = Currency(deref(raw.Currency))
	e.Key = deref(raw.Key)
	return nil
}
