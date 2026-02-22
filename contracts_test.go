package gbkr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContracts_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithPermissions(PermissionSet{}))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}

	_, err = bc.Contracts()
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestContracts_Info(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/contract/265598/info" {
			t.Errorf("path = %q, want /iserver/contract/265598/info", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"con_id":          265598,
			"symbol":          "AAPL",
			"instrument_type": "STK",
			"company_name":    "Apple Inc",
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}

	cr, err := bc.Contracts()
	if err != nil {
		t.Fatal(err)
	}

	info, err := cr.Info(context.Background(), 265598)
	if err != nil {
		t.Fatal(err)
	}
	if info.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", info.Symbol)
	}
}

func TestContracts_Search(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/secdef/search" {
			t.Errorf("path = %q, want /iserver/secdef/search", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("symbol = %q, want AAPL", r.URL.Query().Get("symbol"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{"conid": 265598, "symbol": "AAPL", "companyName": "Apple Inc", "secType": "STK"},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}

	cr, err := bc.Contracts()
	if err != nil {
		t.Fatal(err)
	}

	results, err := cr.Search(context.Background(), "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", results[0].Symbol)
	}
}
