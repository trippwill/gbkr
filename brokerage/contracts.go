package brokerage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/internal/jx"
)

// Contracts provides read access to contract details.
// IBKR path prefix: /iserver/contract/{conid}/*
type Contracts struct {
	c *Client
}

// Contracts returns a [*Contracts] handle for querying contract information.
func (c *Client) Contracts() *Contracts {
	return &Contracts{c: c}
}

// Info returns full contract details
// (GET /iserver/contract/{conid}/info).
func (r *Contracts) Info(ctx context.Context, conID gbkr.ConID) (*ContractDetails, error) {
	start := time.Now()
	var result ContractDetails
	path := fmt.Sprintf("/iserver/contract/%d/info", int(conID))
	err := r.c.doGet(ctx, path, nil, &result)
	r.c.emitOp(ctx, gbkr.OpContractInfo, err, time.Since(start),
		slog.Int64("conid", int64(conID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// ContractDetails represents contract information returned by
// GET /iserver/contract/{conid}/info.
type ContractDetails struct {
	// ConID is the contract identifier.
	ConID gbkr.ConID
	// Symbol is the trading symbol.
	Symbol string
	// SecType is the security type (e.g., "STK", "OPT").
	SecType gbkr.AssetClass
	// Exchange is the primary listing exchange.
	Exchange gbkr.Exchange
	// CompanyName is the company or instrument name.
	CompanyName string
	// Currency is the trading currency.
	Currency gbkr.Currency
	// Multiplier is the contract multiplier.
	Multiplier float64
	// Strike is the option strike price.
	Strike float64
	// Expiry is the contract expiry date (e.g., "20240119").
	Expiry string
	// PutOrCall is "P" for put, "C" for call, or empty for non-options.
	PutOrCall string
	// UndConID is the underlying contract ID (for derivatives).
	UndConID gbkr.ConID
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
	c.ConID = gbkr.ConID(jx.Deref(raw.ConID))
	c.Symbol = jx.Deref(raw.Symbol)
	c.SecType = gbkr.AssetClass(jx.Deref(raw.SecType))
	c.Exchange = gbkr.Exchange(jx.Deref(raw.Exchange))
	c.CompanyName = jx.Deref(raw.CompanyName)
	c.Currency = gbkr.Currency(jx.Deref(raw.Currency))
	c.Multiplier = jx.Deref(raw.Multiplier)
	c.Strike = jx.Deref(raw.Strike)
	c.Expiry = jx.Deref(raw.Expiry)
	c.PutOrCall = jx.Deref(raw.PutOrCall)
	c.UndConID = gbkr.ConID(jx.Deref(raw.UndConID))
	return nil
}

// ContractSearchResult represents a single result from
// GET /iserver/secdef/search.
type ContractSearchResult struct {
	// ConID is the contract identifier.
	ConID gbkr.ConID
	// CompanyName is the company or instrument name.
	CompanyName string
	// Symbol is the trading symbol.
	Symbol string
	// SecType is the security type.
	SecType gbkr.AssetClass
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
	r.ConID = gbkr.ConID(jx.Deref(raw.ConID))
	r.CompanyName = jx.Deref(raw.CompanyName)
	r.Symbol = jx.Deref(raw.Symbol)
	r.SecType = gbkr.AssetClass(jx.Deref(raw.SecType))
	return nil
}
