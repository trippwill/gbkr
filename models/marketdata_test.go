package models

import (
	"encoding/json"
	"testing"
)

func TestFieldValue_Absent(t *testing.T) {
	var fv FieldValue
	if fv.Present() {
		t.Error("zero FieldValue should not be present")
	}
	if fv.Float64() != 0 {
		t.Errorf("Float64() = %f, want 0", fv.Float64())
	}
	if fv.Int64() != 0 {
		t.Errorf("Int64() = %d, want 0", fv.Int64())
	}
	if fv.Bool() {
		t.Error("Bool() should be false")
	}
	if fv.String() != "xxx" {
		t.Errorf("String() = %q, want %q", fv.String(), "xxx")
	}
	if fv.Raw() != nil {
		t.Error("Raw() should be nil")
	}
}

func TestFieldValue_Numeric(t *testing.T) {
	fv := FieldValue{raw: json.RawMessage(`175.50`)}
	if !fv.Present() {
		t.Error("should be present")
	}
	if fv.Float64() != 175.50 {
		t.Errorf("Float64() = %f, want 175.50", fv.Float64())
	}
	if fv.String() != "175.5" {
		t.Errorf("String() = %q, want %q", fv.String(), "175.5")
	}
}

func TestFieldValue_StringEncoded(t *testing.T) {
	// IBKR sometimes returns numbers as strings.
	fv := FieldValue{raw: json.RawMessage(`"123.45"`)}
	if fv.Float64() != 123.45 {
		t.Errorf("Float64() = %f, want 123.45", fv.Float64())
	}
	if fv.String() != "123.45" {
		t.Errorf("String() = %q, want %q", fv.String(), "123.45")
	}
}

func TestFieldValue_Int(t *testing.T) {
	fv := FieldValue{raw: json.RawMessage(`42`)}
	if fv.Int64() != 42 {
		t.Errorf("Int64() = %d, want 42", fv.Int64())
	}
}

func TestFieldValue_Bool(t *testing.T) {
	fv := FieldValue{raw: json.RawMessage(`true`)}
	if !fv.Bool() {
		t.Error("Bool() should be true")
	}
}

func TestFieldValue_RawJSON(t *testing.T) {
	fv := FieldValue{raw: json.RawMessage(`[1,2,3]`)}
	if fv.String() != "[1,2,3]" {
		t.Errorf("String() = %q, want %q", fv.String(), "[1,2,3]")
	}
}

func TestSnapshot_UnmarshalJSON(t *testing.T) {
	data := `{
		"conid": 265598,
		"server_id": "srv1",
		"_updated": 1700000000,
		"31": 175.50,
		"84": "170.25",
		"55": "AAPL"
	}`
	var s Snapshot
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		t.Fatal(err)
	}
	if s.ConID != 265598 {
		t.Errorf("ConID = %d", s.ConID)
	}
	if s.ServerID != "srv1" {
		t.Errorf("ServerID = %q", s.ServerID)
	}
	if s.UpdateTime != 1700000000 {
		t.Errorf("UpdateTime = %d", s.UpdateTime)
	}

	last := s.Get(FieldLast).Float64()
	if last != 175.50 {
		t.Errorf("FieldLast = %f", last)
	}
	bid := s.Get(FieldBid).Float64()
	if bid != 170.25 {
		t.Errorf("FieldBid = %f", bid)
	}
	sym := s.Get(FieldSymbol).String()
	if sym != "AAPL" {
		t.Errorf("FieldSymbol = %q", sym)
	}
}

func TestSnapshot_Get_Absent(t *testing.T) {
	var s Snapshot
	fv := s.Get(FieldLast)
	if fv.Present() {
		t.Error("absent field should not be present")
	}
}

func TestSnapshot_hasAll(t *testing.T) {
	data := `{"conid": 1, "31": 100}`
	var s Snapshot
	json.Unmarshal([]byte(data), &s) //nolint:errcheck

	if !s.hasAll([]SnapshotField{FieldLast}) {
		t.Error("should have FieldLast")
	}
	if s.hasAll([]SnapshotField{FieldLast, FieldBid}) {
		t.Error("should not have FieldBid")
	}
}

func TestSnapshot_AsQuote(t *testing.T) {
	data := `{
		"conid": 1,
		"55": "AAPL", "7051": "Apple Inc.",
		"31": 175.5, "84": 175.0, "86": 176.0,
		"70": 177.0, "71": 174.0, "7295": 174.5, "7296": 175.0, "7741": 173.0,
		"87": 1000000, "82": 2.5, "83": 1.45
	}`
	var s Snapshot
	json.Unmarshal([]byte(data), &s) //nolint:errcheck

	q, ok := s.AsQuote()
	if !ok {
		t.Error("expected ok=true")
	}
	if q.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", q.Symbol)
	}
	if q.Last != "175.5" {
		t.Errorf("Last = %v", q.Last)
	}
	if q.ChangePct != 1.45 {
		t.Errorf("ChangePct = %f", q.ChangePct)
	}
}

func TestSnapshot_AsQuote_Incomplete(t *testing.T) {
	data := `{"conid": 1, "31": 100}`
	var s Snapshot
	json.Unmarshal([]byte(data), &s) //nolint:errcheck

	_, ok := s.AsQuote()
	if ok {
		t.Error("expected ok=false for incomplete data")
	}
}

func TestSnapshot_AsGreeks(t *testing.T) {
	data := `{
		"conid": 1,
		"7308": 0.65, "7309": 0.03, "7310": -0.05,
		"7311": 0.15, "7633": 25.5, "7283": 22.0
	}`
	var s Snapshot
	json.Unmarshal([]byte(data), &s) //nolint:errcheck

	g, ok := s.AsGreeks()
	if !ok {
		t.Error("expected ok=true")
	}
	if g.Delta != 0.65 {
		t.Errorf("Delta = %f", g.Delta)
	}
	if g.Theta != -0.05 {
		t.Errorf("Theta = %f", g.Theta)
	}
}

func TestSnapshot_AsPnL(t *testing.T) {
	data := `{
		"conid": 1,
		"73": 50000, "74": 170.5, "75": 500, "80": 2.5,
		"79": 200, "78": 150, "7292": 17050
	}`
	var s Snapshot
	json.Unmarshal([]byte(data), &s) //nolint:errcheck

	p, ok := s.AsPnL()
	if !ok {
		t.Error("expected ok=true")
	}
	if p.MarketValue != 50000 {
		t.Errorf("MarketValue = %f", p.MarketValue)
	}
}

func TestSnapshot_AsBond(t *testing.T) {
	data := `{
		"conid": 1,
		"7698": 4.5, "7699": 4.4, "7720": 4.6,
		"7706": "AAA", "7708": "Corporate", "7705": "Senior",
		"7715": "2020-01-01", "7714": "2030-12-31"
	}`
	var s Snapshot
	json.Unmarshal([]byte(data), &s) //nolint:errcheck

	b, ok := s.AsBond()
	if !ok {
		t.Error("expected ok=true")
	}
	if b.LastYield != 4.5 {
		t.Errorf("LastYield = %f", b.LastYield)
	}
	if b.Ratings != "AAA" {
		t.Errorf("Ratings = %q", b.Ratings)
	}
}

func TestHistoryResponse_UnmarshalJSON(t *testing.T) {
	data := `{
		"symbol": "AAPL",
		"text": "Apple Inc.",
		"timePeriod": "1d",
		"data": [
			{"o": 170.0, "h": 176.0, "l": 169.5, "c": 175.5, "v": 1000000, "t": 1700000000}
		]
	}`
	var hr HistoryResponse
	if err := json.Unmarshal([]byte(data), &hr); err != nil {
		t.Fatal(err)
	}
	if hr.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", hr.Symbol)
	}
	if hr.TimePeriod != "1d" {
		t.Errorf("TimePeriod = %q", hr.TimePeriod)
	}
	if len(hr.Bars) != 1 {
		t.Fatalf("Bars = %d", len(hr.Bars))
	}
	bar := hr.Bars[0]
	if bar.Open != 170.0 || bar.Close != 175.5 {
		t.Errorf("bar = %+v", bar)
	}
	if bar.Volume != 1000000 {
		t.Errorf("Volume = %f", bar.Volume)
	}
	if bar.Time != 1700000000 {
		t.Errorf("Time = %d", bar.Time)
	}
}

func TestHistoryResponse_Empty(t *testing.T) {
	data := `{}`
	var hr HistoryResponse
	if err := json.Unmarshal([]byte(data), &hr); err != nil {
		t.Fatal(err)
	}
	if hr.Symbol != "" {
		t.Errorf("Symbol = %q, want empty", hr.Symbol)
	}
	if len(hr.Bars) != 0 {
		t.Errorf("Bars = %d, want 0", len(hr.Bars))
	}
}
