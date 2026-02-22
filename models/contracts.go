package models

import "encoding/json"

// ContractDetails represents contract information returned by
// GET /iserver/contract/{conid}/info.
type ContractDetails struct {
	// ConID is the contract identifier.
	ConID ConID
	// Symbol is the trading symbol.
	Symbol string
	// SecType is the security type (e.g., "STK", "OPT").
	SecType AssetClass
	// Exchange is the primary listing exchange.
	Exchange Exchange
	// CompanyName is the company or instrument name.
	CompanyName string
	// Currency is the trading currency.
	Currency Currency
	// Multiplier is the contract multiplier.
	Multiplier float64
	// Strike is the option strike price.
	Strike float64
	// Expiry is the contract expiry date (e.g., "20240119").
	Expiry string
	// PutOrCall is "P" for put, "C" for call, or empty for non-options.
	PutOrCall string
	// UndConID is the underlying contract ID (for derivatives).
	UndConID ConID
}

func (c *ContractDetails) UnmarshalJSON(data []byte) error {
	var raw struct {
		ConID       *int     `json:"con_id,omitempty"`
		Symbol      *string  `json:"symbol,omitempty"`
		SecType     *string  `json:"instrument_type,omitempty"`
		Exchange    *string  `json:"exchange,omitempty"`
		CompanyName *string  `json:"company_name,omitempty"`
		Currency    *string  `json:"currency,omitempty"`
		Multiplier  *float64 `json:"multiplier,omitempty"`
		Strike      *float64 `json:"strike,omitempty"`
		Expiry      *string  `json:"expiry,omitempty"`
		PutOrCall   *string  `json:"right,omitempty"`
		UndConID    *int     `json:"und_conid,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	c.ConID = ConID(deref(raw.ConID))
	c.Symbol = deref(raw.Symbol)
	c.SecType = AssetClass(deref(raw.SecType))
	c.Exchange = Exchange(deref(raw.Exchange))
	c.CompanyName = deref(raw.CompanyName)
	c.Currency = Currency(deref(raw.Currency))
	c.Multiplier = deref(raw.Multiplier)
	c.Strike = deref(raw.Strike)
	c.Expiry = deref(raw.Expiry)
	c.PutOrCall = deref(raw.PutOrCall)
	c.UndConID = ConID(deref(raw.UndConID))
	return nil
}

// ContractSearchResult represents a single result from
// GET /iserver/secdef/search.
type ContractSearchResult struct {
	// ConID is the contract identifier.
	ConID ConID
	// CompanyName is the company or instrument name.
	CompanyName string
	// Symbol is the trading symbol.
	Symbol string
	// SecType is the security type.
	SecType AssetClass
}

func (r *ContractSearchResult) UnmarshalJSON(data []byte) error {
	var raw struct {
		ConID       *int    `json:"conid,omitempty"`
		CompanyName *string `json:"companyName,omitempty"`
		Symbol      *string `json:"symbol,omitempty"`
		SecType     *string `json:"secType,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.ConID = ConID(deref(raw.ConID))
	r.CompanyName = deref(raw.CompanyName)
	r.Symbol = deref(raw.Symbol)
	r.SecType = AssetClass(deref(raw.SecType))
	return nil
}
