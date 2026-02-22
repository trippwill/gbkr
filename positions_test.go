package gbkr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPositions_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithPermissions(PermissionSet{}), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.Positions("U1234567")
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestPositions_ListPositions(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr, _ := c.Positions("U1234567")

	if pr.AccountID() != "U1234567" {
		t.Errorf("AccountID() = %q", pr.AccountID())
	}

	positions, err := pr.ListPositions(context.Background(), 0)
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

func TestPositions_SinglePosition(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr, _ := c.Positions("U1234567")

	pos, err := pr.Position(context.Background(), 265598)
	if err != nil {
		t.Fatal(err)
	}
	if pos.Ticker != "AAPL" {
		t.Errorf("Ticker = %q", pos.Ticker)
	}
}

func TestPositions_PortfolioSummary(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr, _ := c.Positions("U1234567")

	summary, err := pr.PortfolioSummary(context.Background())
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

func TestPositions_Ledger(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	pr, _ := c.Positions("U1234567")

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

func TestAccountReader_Positions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{}) //nolint:errcheck
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}
	ar, err := bc.Account("U1234567")
	if err != nil {
		t.Fatal(err)
	}

	pr, err := ar.Positions()
	if err != nil {
		t.Fatal(err)
	}
	if pr.AccountID() != "U1234567" {
		t.Errorf("AccountID() = %q", pr.AccountID())
	}
}
