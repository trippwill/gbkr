package brokerage

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/num"
)

func TestMarketData_Snapshot(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/marketdata/snapshot" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("conids"); got != "265598" {
			t.Errorf("conids = %q, want %q", got, "265598")
		}
		if got := r.URL.Query().Get("fields"); got != "31" {
			t.Errorf("fields = %q, want %q", got, "31")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{
				"conid":     265598,
				"server_id": "srv1",
				"_updated":  1700000000,
				"31":        175.50,
			},
		})
	})
	defer srv.Close()

	params := SnapshotParams{
		ConIDs: []gbkr.ConID{265598},
		Fields: []SnapshotField{FieldLast},
	}
	result, err := bc.MarketData().Snapshot(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d snapshots, want 1", len(result))
	}
	if result[0].ConID != 265598 {
		t.Errorf("ConID = %d", result[0].ConID)
	}
	if result[0].ServerID != "srv1" {
		t.Errorf("ServerID = %q", result[0].ServerID)
	}
	last := result[0].Get(FieldLast).Num()
	if !last.Equal(num.FromFloat64(175.50)) {
		t.Errorf("FieldLast = %s, want 175.50", last)
	}
}

func TestMarketData_History(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/marketdata/history" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if got := r.URL.Query().Get("conid"); got != "265598" {
			t.Errorf("conid = %q", got)
		}
		if got := r.URL.Query().Get("period"); got != "1d" {
			t.Errorf("period = %q", got)
		}
		if got := r.URL.Query().Get("bar"); got != "5min" {
			t.Errorf("bar = %q", got)
		}
		if got := r.URL.Query().Get("exchange"); got != "NASDAQ" {
			t.Errorf("exchange = %q", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"symbol":     "AAPL",
			"timePeriod": "1d",
			"data": []map[string]any{
				{"o": 170.0, "h": 176.0, "l": 169.5, "c": 175.5, "v": 1000000.0, "t": 1700000000},
			},
		})
	})
	defer srv.Close()

	params := HistoryParams{
		ConID:    265598,
		Period:   TimePeriod{Count: 1, Unit: PeriodDays},
		Bar:      BarSize{Count: 5, Unit: BarMinutes},
		Exchange: "NASDAQ",
	}
	result, err := bc.MarketData().History(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if result.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", result.Symbol)
	}
	if len(result.Bars) != 1 {
		t.Fatalf("got %d bars, want 1", len(result.Bars))
	}
	bar := result.Bars[0]
	if !bar.Open.Equal(num.FromFloat64(170.0)) {
		t.Errorf("Open = %s", bar.Open)
	}
	if !bar.Close.Equal(num.FromFloat64(175.5)) {
		t.Errorf("Close = %s", bar.Close)
	}
}

func TestFieldValue_Absent(t *testing.T) {
	var fv FieldValue
	if fv.Present() {
		t.Error("zero FieldValue should not be present")
	}
	if !fv.Num().IsZero() {
		t.Errorf("Num() = %s, want zero", fv.Num())
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
	if !fv.Num().Equal(num.FromString("175.50")) {
		t.Errorf("Num() = %s, want 175.50", fv.Num())
	}
	if fv.String() != "175.5" {
		t.Errorf("String() = %q, want %q", fv.String(), "175.5")
	}
}

func TestFieldValue_StringEncoded(t *testing.T) {
	fv := FieldValue{raw: json.RawMessage(`"123.45"`)}
	if !fv.Num().Equal(num.FromString("123.45")) {
		t.Errorf("Num() = %s, want 123.45", fv.Num())
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
	data := []byte(`{
		"conid": 265598,
		"server_id": "srv1",
		"_updated": 1700000000,
		"31": 175.50,
		"84": "170.25",
		"55": "AAPL"
	}`)
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		t.Fatal(err)
	}
	if s.ConID != 265598 {
		t.Errorf("ConID = %d", s.ConID)
	}
	if s.ServerID != "srv1" {
		t.Errorf("ServerID = %q", s.ServerID)
	}
	if s.UpdateTime.IsZero() {
		t.Error("UpdateTime should not be zero")
	}

	last := s.Get(FieldLast).Num()
	if !last.Equal(num.FromFloat64(175.50)) {
		t.Errorf("FieldLast = %s", last)
	}
	bid := s.Get(FieldBid).Num()
	if !bid.Equal(num.FromFloat64(170.25)) {
		t.Errorf("FieldBid = %s", bid)
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
	data := []byte(`{"conid": 1, "31": 100}`)
	var s Snapshot
	json.Unmarshal(data, &s) //nolint:errcheck

	if !s.hasAll([]SnapshotField{FieldLast}) {
		t.Error("should have FieldLast")
	}
	if s.hasAll([]SnapshotField{FieldLast, FieldBid}) {
		t.Error("should not have FieldBid")
	}
}

func TestSnapshot_AsQuote(t *testing.T) {
	data := []byte(`{
		"conid": 1,
		"55": "AAPL", "7051": "Apple Inc.",
		"31": 175.5, "84": 175.0, "86": 176.0,
		"70": 177.0, "71": 174.0, "7295": 174.5, "7296": 175.0, "7741": 173.0,
		"87": 1000000, "82": 2.5, "83": 1.45
	}`)
	var s Snapshot
	json.Unmarshal(data, &s) //nolint:errcheck

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
	if !q.ChangePct.Equal(num.FromFloat64(1.45)) {
		t.Errorf("ChangePct = %s", q.ChangePct)
	}
}

func TestSnapshot_AsQuote_Incomplete(t *testing.T) {
	data := []byte(`{"conid": 1, "31": 100}`)
	var s Snapshot
	json.Unmarshal(data, &s) //nolint:errcheck

	_, ok := s.AsQuote()
	if ok {
		t.Error("expected ok=false for incomplete data")
	}
}

func TestSnapshot_AsGreeks(t *testing.T) {
	data := []byte(`{
		"conid": 1,
		"7308": 0.65, "7309": 0.03, "7310": -0.05,
		"7311": 0.15, "7633": 25.5, "7283": 22.0
	}`)
	var s Snapshot
	json.Unmarshal(data, &s) //nolint:errcheck

	g, ok := s.AsGreeks()
	if !ok {
		t.Error("expected ok=true")
	}
	if !g.Delta.Equal(num.FromFloat64(0.65)) {
		t.Errorf("Delta = %s", g.Delta)
	}
	if !g.Theta.Equal(num.FromFloat64(-0.05)) {
		t.Errorf("Theta = %s", g.Theta)
	}
}

func TestSnapshot_AsPnL(t *testing.T) {
	data := []byte(`{
		"conid": 1,
		"73": 50000, "74": 170.5, "75": 500, "80": 2.5,
		"79": 200, "78": 150, "7292": 17050
	}`)
	var s Snapshot
	json.Unmarshal(data, &s) //nolint:errcheck

	p, ok := s.AsPnL()
	if !ok {
		t.Error("expected ok=true")
	}
	if !p.MarketValue.Equal(num.FromFloat64(50000)) {
		t.Errorf("MarketValue = %s", p.MarketValue)
	}
}

func TestSnapshot_AsBond(t *testing.T) {
	data := []byte(`{
		"conid": 1,
		"7698": 4.5, "7699": 4.4, "7720": 4.6,
		"7706": "AAA", "7708": "Corporate", "7705": "Senior",
		"7715": "2020-01-01", "7714": "2030-12-31"
	}`)
	var s Snapshot
	json.Unmarshal(data, &s) //nolint:errcheck

	b, ok := s.AsBond()
	if !ok {
		t.Error("expected ok=true")
	}
	if !b.LastYield.Equal(num.FromFloat64(4.5)) {
		t.Errorf("LastYield = %s", b.LastYield)
	}
	if b.Ratings != "AAA" {
		t.Errorf("Ratings = %q", b.Ratings)
	}
}

func TestHistoryResponse_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"symbol": "AAPL",
		"text": "Apple Inc.",
		"timePeriod": "1d",
		"data": [
			{"o": 170.0, "h": 176.0, "l": 169.5, "c": 175.5, "v": 1000000, "t": 1700000000}
		]
	}`)
	var hr HistoryResponse
	if err := json.Unmarshal(data, &hr); err != nil {
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
	if !bar.Open.Equal(num.FromFloat64(170.0)) || !bar.Close.Equal(num.FromFloat64(175.5)) {
		t.Errorf("bar = Open:%s Close:%s", bar.Open, bar.Close)
	}
	if !bar.Volume.Equal(num.FromFloat64(1000000)) {
		t.Errorf("Volume = %s", bar.Volume)
	}
	if bar.Time.IsZero() {
		t.Error("Time should not be zero")
	}
}

func TestHistoryResponse_Empty(t *testing.T) {
	data := []byte(`{}`)
	var hr HistoryResponse
	if err := json.Unmarshal(data, &hr); err != nil {
		t.Fatal(err)
	}
	if hr.Symbol != "" {
		t.Errorf("Symbol = %q, want empty", hr.Symbol)
	}
	if len(hr.Bars) != 0 {
		t.Errorf("Bars = %d, want 0", len(hr.Bars))
	}
}

func TestFieldValue_InvalidTypes(t *testing.T) {
	// Int64: non-numeric JSON triggers fallback to zero.
	intFV := FieldValue{raw: json.RawMessage(`"not_a_number"`)}
	if intFV.Int64() != 0 {
		t.Errorf("Int64() = %d, want 0 for non-numeric string", intFV.Int64())
	}

	// Bool: non-boolean JSON triggers fallback to false.
	boolFV := FieldValue{raw: json.RawMessage(`"yes"`)}
	if boolFV.Bool() {
		t.Error("Bool() should be false for non-boolean string")
	}

	// Num: string that can't be parsed as number → Num with error.
	numFV := FieldValue{raw: json.RawMessage(`"not_a_float"`)}
	if numFV.Num().Err == nil {
		t.Error("Num() should have error for unparseable string")
	}

	// Num: non-string, non-number JSON (e.g. array) → Num with error.
	arrayFV := FieldValue{raw: json.RawMessage(`[1,2,3]`)}
	if arrayFV.Num().Err == nil {
		t.Error("Num() should have error for array value")
	}
}

func TestSnapshot_hasAll_NilFields(t *testing.T) {
	// A zero-value snapshot has nil fields.
	var s Snapshot
	if s.hasAll([]SnapshotField{FieldLast}) {
		t.Error("hasAll should return false for nil fields with non-empty check list")
	}
	if !s.hasAll(nil) {
		t.Error("hasAll should return true for nil fields with empty check list")
	}
}

func TestMarketData_Snapshot_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.MarketData().Snapshot(context.Background(), SnapshotParams{
		ConIDs: []gbkr.ConID{265598},
		Fields: []SnapshotField{FieldLast},
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestMarketData_History_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.MarketData().History(context.Background(), HistoryParams{
		ConID: 265598,
	})
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
