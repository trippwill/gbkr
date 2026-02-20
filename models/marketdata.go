package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FieldValue wraps a raw JSON value from a snapshot response.
// Zero-value is safe: an absent field returns zero/empty from all accessors.
type FieldValue struct {
	raw json.RawMessage
}

// Present reports whether the field was included in the response.
func (v FieldValue) Present() bool { return len(v.raw) > 0 }

// Float64 returns the value as float64, or 0 if absent or not numeric.
func (v FieldValue) Float64() float64 {
	if !v.Present() {
		return 0
	}
	var f float64
	if json.Unmarshal(v.raw, &f) == nil {
		return f
	}
	// IBKR sometimes returns numeric fields as strings.
	var s string
	if json.Unmarshal(v.raw, &s) == nil {
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return f
		}
	}
	return 0
}

// Int64 returns the value as int64, or 0 if absent or not numeric.
func (v FieldValue) Int64() int64 {
	if !v.Present() {
		return 0
	}
	var n int64
	if json.Unmarshal(v.raw, &n) == nil {
		return n
	}
	return 0
}

// Bool returns the value as bool, or false if absent.
func (v FieldValue) Bool() bool {
	if !v.Present() {
		return false
	}
	var b bool
	if json.Unmarshal(v.raw, &b) == nil {
		return b
	}
	return false
}

// Raw returns the underlying JSON for custom decoding.
func (v FieldValue) Raw() json.RawMessage { return v.raw }

// String returns a print-friendly representation of the value.
// Satisfies fmt.Stringer so FieldValue works directly in format verbs.
func (v FieldValue) String() string {
	if !v.Present() {
		return "xxx" // Placeholder for absent fields
	}

	// Try string first (most common for text fields).
	var s string
	if json.Unmarshal(v.raw, &s) == nil {
		return s
	}

	// Try number — format without trailing zeros.
	var f float64
	if json.Unmarshal(v.raw, &f) == nil {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}

	// Fallback: raw JSON text.
	return strings.TrimSpace(string(v.raw))
}

var _ fmt.Stringer = FieldValue{}

// SnapshotParams are query parameters for GET /iserver/marketdata/snapshot.
type SnapshotParams struct {
	ConIDs []ConID         // Contract IDs to query
	Fields []SnapshotField // Field codes to return
}

// Snapshot is one element of the response for GET /iserver/marketdata/snapshot.
// Metadata fields (ConID, ServerID, UpdateTime) are always populated.
// Market data fields are accessed via Get.
type Snapshot struct {
	ConID      ConID
	ServerID   string
	UpdateTime int64
	fields     map[SnapshotField]json.RawMessage
}

// Get returns the FieldValue for the given snapshot field.
// The returned value is safe to use even if the field was not in the response.
func (s Snapshot) Get(f SnapshotField) FieldValue {
	if s.fields == nil {
		return FieldValue{}
	}
	return FieldValue{raw: s.fields[f]}
}

// hasAll reports whether the snapshot contains all of the given fields.
func (s Snapshot) hasAll(fields []SnapshotField) bool {
	if s.fields == nil {
		return len(fields) == 0
	}
	for _, f := range fields {
		if _, ok := s.fields[f]; !ok {
			return false
		}
	}
	return true
}

func (s *Snapshot) UnmarshalJSON(data []byte) error {
	// Decode all keys into a generic map.
	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}

	// Extract metadata fields.
	if raw, ok := all["conid"]; ok {
		var id int
		if json.Unmarshal(raw, &id) == nil {
			s.ConID = ConID(id)
		}
		delete(all, "conid")
	}
	if raw, ok := all["server_id"]; ok {
		var v string
		if json.Unmarshal(raw, &v) == nil {
			s.ServerID = v
		}
		delete(all, "server_id")
	}
	if raw, ok := all["_updated"]; ok {
		var v int64
		if json.Unmarshal(raw, &v) == nil {
			s.UpdateTime = v
		}
		delete(all, "_updated")
	}

	// Everything remaining is a market data field.
	if len(all) > 0 {
		s.fields = make(map[SnapshotField]json.RawMessage, len(all))
		for k, v := range all {
			s.fields[SnapshotField(k)] = v
		}
	}
	return nil
}

// Quote is a strongly-typed projection of a Snapshot using FieldsQuote.
type Quote struct {
	Symbol      string
	CompanyName string
	Last        string
	Bid         float64
	Ask         float64
	High        float64
	Low         float64
	Open        float64
	Close       float64
	PriorClose  float64
	Volume      float64
	Change      float64
	ChangePct   float64
}

// AsQuote projects the snapshot into a Quote struct.
// Returns ok=true if all FieldsQuote fields are present.
func (s Snapshot) AsQuote() (Quote, bool) {
	return Quote{
		Symbol:      s.Get(FieldSymbol).String(),
		CompanyName: s.Get(FieldCompanyName).String(),
		Last:        s.Get(FieldLast).String(),
		Bid:         s.Get(FieldBid).Float64(),
		Ask:         s.Get(FieldAsk).Float64(),
		High:        s.Get(FieldHigh).Float64(),
		Low:         s.Get(FieldLow).Float64(),
		Open:        s.Get(FieldOpen).Float64(),
		Close:       s.Get(FieldClose).Float64(),
		PriorClose:  s.Get(FieldPriorClose).Float64(),
		Volume:      s.Get(FieldVolume).Float64(),
		Change:      s.Get(FieldChange).Float64(),
		ChangePct:   s.Get(FieldChangePct).Float64(),
	}, s.hasAll(FieldsQuote)
}

// Greeks is a strongly-typed projection of a Snapshot using FieldsGreeks.
type Greeks struct {
	Delta      float64
	Gamma      float64
	Theta      float64
	Vega       float64
	ImpliedVol float64
	OptIV      float64 // Implied vol from underlying (30-day)
}

// AsGreeks projects the snapshot into a Greeks struct.
// Returns ok=true if all FieldsGreeks fields are present.
func (s Snapshot) AsGreeks() (Greeks, bool) {
	return Greeks{
		Delta:      s.Get(FieldDelta).Float64(),
		Gamma:      s.Get(FieldGamma).Float64(),
		Theta:      s.Get(FieldTheta).Float64(),
		Vega:       s.Get(FieldVega).Float64(),
		ImpliedVol: s.Get(FieldImpliedVol).Float64(),
		OptIV:      s.Get(FieldOptImpliedVol).Float64(),
	}, s.hasAll(FieldsGreeks)
}

// PnLSnapshot is a strongly-typed projection of a Snapshot using FieldsPnL.
type PnLSnapshot struct {
	MarketValue      float64
	AvgPrice         float64
	UnrealizedPnL    float64
	UnrealizedPnLPct float64
	RealizedPnL      float64
	DailyPnL         float64
	CostBasis        float64
}

// AsPnL projects the snapshot into a PnLSnapshot struct.
// Returns ok=true if all FieldsPnL fields are present.
func (s Snapshot) AsPnL() (PnLSnapshot, bool) {
	return PnLSnapshot{
		MarketValue:      s.Get(FieldMarketValue).Float64(),
		AvgPrice:         s.Get(FieldAvgPrice).Float64(),
		UnrealizedPnL:    s.Get(FieldUnrealizedPnL).Float64(),
		UnrealizedPnLPct: s.Get(FieldUnrealizedPnLPct).Float64(),
		RealizedPnL:      s.Get(FieldRealizedPnL).Float64(),
		DailyPnL:         s.Get(FieldDailyPnL).Float64(),
		CostBasis:        s.Get(FieldCostBasis).Float64(),
	}, s.hasAll(FieldsPnL)
}

// BondSnapshot is a strongly-typed projection of a Snapshot using FieldsBond.
type BondSnapshot struct {
	LastYield       float64
	BidYield        float64
	AskYield        float64
	Ratings         string
	BondType        string
	DebtClass       string
	IssueDate       string
	LastTradingDate string
}

// AsBond projects the snapshot into a BondSnapshot struct.
// Returns ok=true if all FieldsBond fields are present.
func (s Snapshot) AsBond() (BondSnapshot, bool) {
	return BondSnapshot{
		LastYield:       s.Get(FieldLastYield).Float64(),
		BidYield:        s.Get(FieldBidYield).Float64(),
		AskYield:        s.Get(FieldAskYield).Float64(),
		Ratings:         s.Get(FieldRatings).String(),
		BondType:        s.Get(FieldBondType).String(),
		DebtClass:       s.Get(FieldDebtClass).String(),
		IssueDate:       s.Get(FieldIssueDate).String(),
		LastTradingDate: s.Get(FieldLastTradingDate).String(),
	}, s.hasAll(FieldsBond)
}

// HistoryParams are query parameters for GET /iserver/marketdata/history.
type HistoryParams struct {
	ConID    ConID      // Contract ID
	Period   TimePeriod // Time period (e.g., Period(1, PeriodDays))
	Bar      BarSize    // Bar size (e.g., Bar(5, BarMinutes))
	Exchange Exchange   // Optional exchange
}

// HistoryResponse is the response for GET /iserver/marketdata/history.
type HistoryResponse struct {
	Symbol     string
	Text       string
	TimePeriod string
	Bars       []HistoryBar
}

func (h *HistoryResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Symbol     *string      `json:"symbol,omitempty"`
		Text       *string      `json:"text,omitempty"`
		TimePeriod *string      `json:"timePeriod,omitempty"`
		Bars       []HistoryBar `json:"data,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	h.Symbol = deref(raw.Symbol)
	h.Text = deref(raw.Text)
	h.TimePeriod = deref(raw.TimePeriod)
	h.Bars = raw.Bars
	return nil
}

// HistoryBar represents a single OHLC bar.
type HistoryBar struct {
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	Time   int64
}

func (b *HistoryBar) UnmarshalJSON(data []byte) error {
	var raw struct {
		Open   *float64 `json:"o,omitempty"`
		High   *float64 `json:"h,omitempty"`
		Low    *float64 `json:"l,omitempty"`
		Close  *float64 `json:"c,omitempty"`
		Volume *float64 `json:"v,omitempty"`
		Time   *int64   `json:"t,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	b.Open = deref(raw.Open)
	b.High = deref(raw.High)
	b.Low = deref(raw.Low)
	b.Close = deref(raw.Close)
	b.Volume = deref(raw.Volume)
	b.Time = deref(raw.Time)
	return nil
}
