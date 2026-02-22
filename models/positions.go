package models

import (
	"encoding/json"
	"fmt"
	"strconv"
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
	Qty float64
	// MktPrice is the current market price per share.
	MktPrice float64
	// MktValue is the total market value of the position.
	MktValue float64
	// AvgCost is the average cost per share multiplied by the contract multiplier.
	AvgCost float64
	// AvgPrice is the average purchase price per share.
	AvgPrice float64
	// RealizedPnL is the realized profit/loss.
	RealizedPnL float64
	// UnrealizedPnL is the unrealized profit/loss.
	UnrealizedPnL float64
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
	// Strike is the option strike price. Zero for non-options.
	Strike float64
	// UndConID is the underlying contract ID for derivatives. Zero for stocks.
	UndConID ConID
	// Multiplier is the contract multiplier (e.g., 100 for equity options).
	Multiplier float64
}

func (p *Position) UnmarshalJSON(data []byte) error {
	var raw struct {
		AcctID        *string          `json:"acctId,omitempty"`
		ConID         *int             `json:"conid,omitempty"`
		ContractDesc  *string          `json:"contractDesc,omitempty"`
		Position      *float64         `json:"position,omitempty"`
		MktPrice      *float64         `json:"mktPrice,omitempty"`
		MktValue      *float64         `json:"mktValue,omitempty"`
		AvgCost       *float64         `json:"avgCost,omitempty"`
		AvgPrice      *float64         `json:"avgPrice,omitempty"`
		RealizedPnL   *float64         `json:"realizedPnl,omitempty"`
		UnrealizedPnL *float64         `json:"unrealizedPnl,omitempty"`
		Currency      *string          `json:"currency,omitempty"`
		AssetClass    *string          `json:"assetClass,omitempty"`
		Ticker        *string          `json:"ticker,omitempty"`
		Expiry        *string          `json:"expiry,omitempty"`
		PutOrCall     *string          `json:"putOrCall,omitempty"`
		Strike        *json.RawMessage `json:"strike,omitempty"`
		UndConID      *int             `json:"undConid,omitempty"`
		Multiplier    *json.RawMessage `json:"multiplier,omitempty"`
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
	p.Expiry = deref(raw.Expiry)
	p.PutOrCall = deref(raw.PutOrCall)
	strike, err := decodeFlexibleFloat(raw.Strike)
	if err != nil {
		return err
	}
	p.Strike = strike
	p.UndConID = ConID(deref(raw.UndConID))
	multiplier, err := decodeFlexibleFloat(raw.Multiplier)
	if err != nil {
		return err
	}
	p.Multiplier = multiplier
	return nil
}

func decodeFlexibleFloat(raw *json.RawMessage) (float64, error) {
	if raw == nil || len(*raw) == 0 {
		return 0, nil
	}
	var f float64
	if err := json.Unmarshal(*raw, &f); err == nil {
		return f, nil
	}
	var s string
	if err := json.Unmarshal(*raw, &s); err == nil {
		if s == "" {
			return 0, nil
		}
		parsed, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, err
		}
		return parsed, nil
	}
	return 0, fmt.Errorf("decode flexible float: unsupported JSON value %s", string(*raw))
}

// PortfolioSummary is the response for GET /portfolio/{accountId}/summary.
// It maps field names to their corresponding summary field objects.
type PortfolioSummary map[string]PortfolioSummaryField

// PortfolioSummaryField represents a single field within a portfolio summary.
type PortfolioSummaryField struct {
	// Amount is the numeric amount of the field.
	Amount float64
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
// It maps currency codes to their corresponding ledger entries.
type Ledger map[Currency]LedgerEntry

// LedgerEntry holds ledger data for a single currency.
type LedgerEntry struct {
	// CommodityMarketValue is the market value of commodity positions.
	CommodityMarketValue float64
	// FutureOptionValue is the market value of futures options positions.
	FutureOptionValue float64
	// FuturesPnL is the PnL of futures positions. (API: "futuresonlypnl")
	FuturesPnL float64
	// Interest is the receivable interest balance.
	Interest float64
	// NetLiquidation is the net liquidation value of positions. (API: "netliquidationvalue")
	NetLiquidation float64
	// RealizedPnL is the realized PnL.
	RealizedPnL float64
	// SettledCash is the settled cash balance.
	SettledCash float64
	// StockMarketValue is the market value of stock positions.
	StockMarketValue float64
	// TotalCashValue is the total cash value.
	TotalCashValue float64
	// UnrealizedPnL is the unrealized PnL.
	UnrealizedPnL float64
	// Currency is the three-letter currency code.
	Currency Currency
	// Key is always "LedgerList".
	Key string
}

func (e *LedgerEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		CommodityMarketValue *float64 `json:"commoditymarketvalue,omitempty"`
		FutureOptionValue    *float64 `json:"futureoptionmarketvalue,omitempty"`
		FuturesPnL           *float64 `json:"futuresonlypnl,omitempty"`
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
