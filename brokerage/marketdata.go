package brokerage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/internal/jx"
	"github.com/trippwill/gbkr/num"
	"github.com/trippwill/gbkr/when"
)

// MarketData provides read access to IBKR market data.
// IBKR path prefix: /iserver/marketdata/*
type MarketData struct {
	c *Client
}

// MarketData returns a [*MarketData] handle.
func (c *Client) MarketData() *MarketData {
	return &MarketData{c: c}
}

// Snapshot returns a live market data snapshot
// (GET /iserver/marketdata/snapshot).
func (m *MarketData) Snapshot(ctx context.Context, params SnapshotParams) ([]Snapshot, error) {
	start := time.Now()
	q := url.Values{}
	if len(params.ConIDs) > 0 {
		ids := make([]string, len(params.ConIDs))
		for i, id := range params.ConIDs {
			ids[i] = fmt.Sprintf("%d", int(id))
		}
		q.Set("conids", strings.Join(ids, ","))
	}
	if len(params.Fields) > 0 {
		fs := make([]string, len(params.Fields))
		for i, f := range params.Fields {
			fs[i] = f.String()
		}
		q.Set("fields", strings.Join(fs, ","))
	}

	var result []Snapshot
	err := m.c.doGet(ctx, "/iserver/marketdata/snapshot", q, &result)
	m.c.emitOp(ctx, gbkr.OpMarketDataSnapshot, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// History returns historical OHLC bar data
// (GET /iserver/marketdata/history).
func (m *MarketData) History(ctx context.Context, params HistoryParams) (*HistoryResponse, error) {
	start := time.Now()
	q := url.Values{}
	q.Set("conid", fmt.Sprintf("%d", int(params.ConID)))
	q.Set("period", params.Period.String())
	q.Set("bar", params.Bar.String())
	if params.Exchange != "" {
		q.Set("exchange", params.Exchange.String())
	}

	var result HistoryResponse
	err := m.c.doGet(ctx, "/iserver/marketdata/history", q, &result)
	m.c.emitOp(ctx, gbkr.OpMarketDataHistory, err, time.Since(start),
		slog.Int64("conid", int64(params.ConID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// FieldValue wraps a raw JSON value from a snapshot response.
// Zero-value is safe: an absent field returns zero/empty from all accessors.
type FieldValue struct {
	raw json.RawMessage
}

// Present reports whether the field was included in the response.
func (v FieldValue) Present() bool { return len(v.raw) > 0 }

// Num returns the value as [num.Num], handling both bare numeric and
// quoted string formats from the IBKR API. Returns [num.Zero] if absent.
func (v FieldValue) Num() num.Num {
	if !v.Present() {
		return num.Zero()
	}
	n := num.Zero()
	_ = n.UnmarshalJSON(v.raw)
	return n
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
	ConIDs []gbkr.ConID    // Contract IDs to query
	Fields []SnapshotField // Field codes to return
}

// Snapshot is one element of the response array for GET /iserver/marketdata/snapshot.
// Metadata fields (ConID, ServerID, UpdateTime) are always populated.
// Dynamic market data fields are accessed via Get.
type Snapshot struct {
	ConID      gbkr.ConID    // Contract identifier.
	ServerID   string        // Internal server ID. (API: "server_id")
	UpdateTime when.DateTime // Last update timestamp. (API: "_updated")
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
			s.ConID = gbkr.ConID(id)
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
			s.UpdateTime = when.DateTimeFromEpoch(v)
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
	Symbol      string  // Ticker symbol.
	CompanyName string  // Company or instrument name.
	Last        string  // Last traded price (string; may contain formatting).
	Bid         num.Num // Current bid price.
	Ask         num.Num // Current ask price.
	High        num.Num // Day high.
	Low         num.Num // Day low.
	Open        num.Num // Day open.
	Close       num.Num // Day close.
	PriorClose  num.Num // Previous session close.
	Volume      num.Num // Day volume.
	Change      num.Num // Absolute price change.
	ChangePct   num.Num // Percentage price change.
}

// AsQuote projects the snapshot into a Quote struct.
// Returns ok=true if all FieldsQuote fields are present.
func (s Snapshot) AsQuote() (Quote, bool) {
	return Quote{
		Symbol:      s.Get(FieldSymbol).String(),
		CompanyName: s.Get(FieldCompanyName).String(),
		Last:        s.Get(FieldLast).String(),
		Bid:         s.Get(FieldBid).Num(),
		Ask:         s.Get(FieldAsk).Num(),
		High:        s.Get(FieldHigh).Num(),
		Low:         s.Get(FieldLow).Num(),
		Open:        s.Get(FieldOpen).Num(),
		Close:       s.Get(FieldClose).Num(),
		PriorClose:  s.Get(FieldPriorClose).Num(),
		Volume:      s.Get(FieldVolume).Num(),
		Change:      s.Get(FieldChange).Num(),
		ChangePct:   s.Get(FieldChangePct).Num(),
	}, s.hasAll(FieldsQuote)
}

// Greeks is a strongly-typed projection of a Snapshot using FieldsGreeks.
type Greeks struct {
	Delta      num.Num // Rate of change of option price vs underlying.
	Gamma      num.Num // Rate of change of delta.
	Theta      num.Num // Time decay per day.
	Vega       num.Num // Sensitivity to implied volatility.
	ImpliedVol num.Num // Option implied volatility.
	OptIV      num.Num // Implied vol from underlying (30-day).
}

// AsGreeks projects the snapshot into a Greeks struct.
// Returns ok=true if all FieldsGreeks fields are present.
func (s Snapshot) AsGreeks() (Greeks, bool) {
	return Greeks{
		Delta:      s.Get(FieldDelta).Num(),
		Gamma:      s.Get(FieldGamma).Num(),
		Theta:      s.Get(FieldTheta).Num(),
		Vega:       s.Get(FieldVega).Num(),
		ImpliedVol: s.Get(FieldImpliedVol).Num(),
		OptIV:      s.Get(FieldOptImpliedVol).Num(),
	}, s.hasAll(FieldsGreeks)
}

// PnLSnapshot is a strongly-typed projection of a Snapshot using FieldsPnL.
type PnLSnapshot struct {
	MarketValue      num.Num // Current market value of the position.
	AvgPrice         num.Num // Average entry price.
	UnrealizedPnL    num.Num // Unrealized profit/loss.
	UnrealizedPnLPct num.Num // Unrealized P&L as a percentage.
	RealizedPnL      num.Num // Realized profit/loss.
	DailyPnL         num.Num // Day profit/loss.
	CostBasis        num.Num // Total cost basis.
}

// AsPnL projects the snapshot into a PnLSnapshot struct.
// Returns ok=true if all FieldsPnL fields are present.
func (s Snapshot) AsPnL() (PnLSnapshot, bool) {
	return PnLSnapshot{
		MarketValue:      s.Get(FieldMarketValue).Num(),
		AvgPrice:         s.Get(FieldAvgPrice).Num(),
		UnrealizedPnL:    s.Get(FieldUnrealizedPnL).Num(),
		UnrealizedPnLPct: s.Get(FieldUnrealizedPnLPct).Num(),
		RealizedPnL:      s.Get(FieldRealizedPnL).Num(),
		DailyPnL:         s.Get(FieldDailyPnL).Num(),
		CostBasis:        s.Get(FieldCostBasis).Num(),
	}, s.hasAll(FieldsPnL)
}

// BondSnapshot is a strongly-typed projection of a Snapshot using FieldsBond.
type BondSnapshot struct {
	LastYield       num.Num       // Last yield value.
	BidYield        num.Num       // Bid yield value.
	AskYield        num.Num       // Ask yield value.
	Ratings         string        // Credit ratings.
	BondType        string        // Bond type classification.
	DebtClass       string        // Debt class (e.g. senior, subordinated).
	IssueDate       when.NullDate // Bond issue date.
	LastTradingDate when.NullDate // Last trading date for the bond.
}

// AsBond projects the snapshot into a BondSnapshot struct.
// Returns ok=true if all FieldsBond fields are present.
func (s Snapshot) AsBond() (BondSnapshot, bool) {
	return BondSnapshot{
		LastYield:       s.Get(FieldLastYield).Num(),
		BidYield:        s.Get(FieldBidYield).Num(),
		AskYield:        s.Get(FieldAskYield).Num(),
		Ratings:         s.Get(FieldRatings).String(),
		BondType:        s.Get(FieldBondType).String(),
		DebtClass:       s.Get(FieldDebtClass).String(),
		IssueDate:       parseNullDateFromJSON(s.Get(FieldIssueDate).String()),
		LastTradingDate: parseNullDateFromJSON(s.Get(FieldLastTradingDate).String()),
	}, s.hasAll(FieldsBond)
}

// HistoryParams are query parameters for GET /iserver/marketdata/history.
type HistoryParams struct {
	ConID    gbkr.ConID    // Contract ID.
	Period   TimePeriod    // Duration of the request (e.g., Period(1, PeriodDays)).
	Bar      BarSize       // Bar size (e.g., Bar(5, BarMinutes)).
	Exchange gbkr.Exchange // Optional exchange filter.
}

// HistoryResponse is the response for GET /iserver/marketdata/history.
type HistoryResponse struct {
	Symbol     string       // Ticker symbol.
	Text       string       // Long name (e.g. "APPLE INC").
	TimePeriod string       // Duration of the request.
	Bars       []HistoryBar // OHLCV bar data. (API: "data")
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
	h.Symbol = jx.Deref(raw.Symbol)
	h.Text = jx.Deref(raw.Text)
	h.TimePeriod = jx.Deref(raw.TimePeriod)
	h.Bars = raw.Bars
	return nil
}

// HistoryBar represents a single OHLCV bar in a history response.
type HistoryBar struct {
	Open   num.Num       // Open price. (API: "o")
	High   num.Num       // High price. (API: "h")
	Low    num.Num       // Low price. (API: "l")
	Close  num.Num       // Close price. (API: "c")
	Volume num.Num       // Volume. (API: "v")
	Time   when.DateTime // Bar timestamp. (API: "t")
}

func (b *HistoryBar) UnmarshalJSON(data []byte) error {
	raw := struct {
		Open   num.Num `json:"o"`
		High   num.Num `json:"h"`
		Low    num.Num `json:"l"`
		Close  num.Num `json:"c"`
		Volume num.Num `json:"v"`
		Time   *int64  `json:"t,omitempty"`
	}{
		Open:   num.Zero(),
		High:   num.Zero(),
		Low:    num.Zero(),
		Close:  num.Zero(),
		Volume: num.Zero(),
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	b.Open = raw.Open
	b.High = raw.High
	b.Low = raw.Low
	b.Close = raw.Close
	b.Volume = raw.Volume
	b.Time = when.DateTimeFromEpoch(jx.Deref(raw.Time))
	return nil
}
