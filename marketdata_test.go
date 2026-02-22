package gbkr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippwill/gbkr/models"
)

func TestMarketData_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithPermissions(PermissionSet{}), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	bc := &BrokerageClient{Client: c}
	_, err = bc.MarketData()
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestMarketData_Snapshot(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}
	md, _ := bc.MarketData()

	params := models.SnapshotParams{
		ConIDs: []models.ConID{265598},
		Fields: []models.SnapshotField{models.FieldLast},
	}
	result, err := md.Snapshot(context.Background(), params)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 {
		t.Fatalf("got %d snapshots, want 1", len(result))
	}
	if result[0].ConID != 265598 {
		t.Errorf("ConID = %d, want 265598", result[0].ConID)
	}
	if result[0].ServerID != "srv1" {
		t.Errorf("ServerID = %q", result[0].ServerID)
	}
	last := result[0].Get(models.FieldLast).Float64()
	if last != 175.50 {
		t.Errorf("FieldLast = %f, want 175.50", last)
	}
}

func TestMarketData_History(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}
	md, _ := bc.MarketData()

	params := models.HistoryParams{
		ConID:    265598,
		Period:   models.TimePeriod{Count: 1, Unit: models.PeriodDays},
		Bar:      models.BarSize{Count: 5, Unit: models.BarMinutes},
		Exchange: "NASDAQ",
	}
	result, err := md.History(context.Background(), params)
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
	if bar.Open != 170.0 {
		t.Errorf("Open = %f", bar.Open)
	}
	if bar.Close != 175.5 {
		t.Errorf("Close = %f", bar.Close)
	}
}
