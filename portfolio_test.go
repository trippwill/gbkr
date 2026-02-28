package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPortfolio_Positions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/U1234567/positions/0" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{
				"acctId":       "U1234567",
				"conid":        265598,
				"contractDesc": "AAPL",
				"position":     100.0,
				"mktPrice":     175.50,
				"currency":     "USD",
				"ticker":       "AAPL",
			},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")

	if pr.ID() != "U1234567" {
		t.Errorf("ID() = %q", pr.ID())
	}

	positions, err := pr.Positions(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(positions) != 1 {
		t.Fatalf("got %d positions, want 1", len(positions))
	}
	if positions[0].Ticker != "AAPL" {
		t.Errorf("Ticker = %q, want %q", positions[0].Ticker, "AAPL")
	}
	if positions[0].Qty != 100 {
		t.Errorf("Qty = %f, want 100", positions[0].Qty)
	}
}

func TestPortfolio_SinglePosition(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/U1234567/position/265598" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"conid":    265598,
			"ticker":   "AAPL",
			"position": 50.0,
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")

	pos, err := pr.Position(context.Background(), 265598)
	if err != nil {
		t.Fatal(err)
	}
	if pos.Ticker != "AAPL" {
		t.Errorf("Ticker = %q", pos.Ticker)
	}
}

func TestPortfolio_Summary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/U1234567/summary" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"totalcashvalue": map[string]any{
				"amount":   10000.0,
				"currency": "USD",
			},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")

	summary, err := pr.Summary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if summary == nil {
		t.Fatal("nil summary")
	}
	field, ok := (*summary)["totalcashvalue"]
	if !ok {
		t.Fatal("missing totalcashvalue")
	}
	if field.Amount != 10000.0 {
		t.Errorf("Amount = %f, want 10000", field.Amount)
	}
}

func TestPortfolio_Ledger(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/U1234567/ledger" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"USD": map[string]any{
				"netliquidationvalue": 50000.0,
				"totalcashvalue":      10000.0,
				"currency":            "USD",
				"key":                 "LedgerList",
			},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")

	ledger, err := pr.Ledger(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if ledger == nil {
		t.Fatal("nil ledger")
	}
	entry, ok := (*ledger)["USD"]
	if !ok {
		t.Fatal("missing USD entry")
	}
	if entry.NetLiquidation != 50000.0 {
		t.Errorf("NetLiquidation = %f, want 50000", entry.NetLiquidation)
	}
}

func TestPortfolio_GatewayAccess(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")
	if pr == nil {
		t.Fatal("expected non-nil Portfolio")
	}
}

func TestPortfolio_Allocation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/U1234567/allocation" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"assetClass": map[string]any{
				"STK": map[string]any{"long": 85000.0, "short": 0.0},
				"OPT": map[string]any{"long": 5000.0, "short": -2000.0},
			},
			"sector": map[string]any{
				"Technology": map[string]any{"long": 50000.0, "short": 0.0},
			},
			"group": map[string]any{
				"US": map[string]any{"long": 90000.0, "short": -2000.0},
			},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")

	alloc, err := pr.Allocation(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	stk, ok := alloc.AssetClass[AssetStock]
	if !ok {
		t.Fatal("missing STK in AssetClass")
	}
	if stk.Long != 85000 {
		t.Errorf("STK long = %f, want 85000", stk.Long)
	}
	tech, ok := alloc.Sector["Technology"]
	if !ok {
		t.Fatal("missing Technology in Sector")
	}
	if tech.Long != 50000 {
		t.Errorf("Technology long = %f", tech.Long)
	}
	us, ok := alloc.Group["US"]
	if !ok {
		t.Fatal("missing US in Group")
	}
	if us.Short != -2000 {
		t.Errorf("US short = %f", us.Short)
	}
}

func TestAllocation_UnmarshalJSON_Empty(t *testing.T) {
	data := `{}`
	var a Allocation
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		t.Fatal(err)
	}
	if a.AssetClass != nil {
		t.Errorf("AssetClass should be nil, got %v", a.AssetClass)
	}
}

func TestPortfolio_InvalidatePositions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/U1234567/positions/invalidate" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr := c.Portfolio("U1234567")

	err = pr.InvalidatePositions(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}
