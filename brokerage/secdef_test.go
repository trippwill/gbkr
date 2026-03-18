package brokerage

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestSecurityDefinitions_Search(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/secdef/search" {
			t.Errorf("path = %q", r.URL.Path)
		}
		if r.URL.Query().Get("symbol") != "AAPL" {
			t.Errorf("symbol = %q", r.URL.Query().Get("symbol"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{"conid": 265598, "symbol": "AAPL", "companyName": "Apple Inc", "secType": "STK"},
		})
	})
	defer srv.Close()

	results, err := bc.SecurityDefinitions().Search(context.Background(), "AAPL")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("got %d results, want 1", len(results))
	}
	if results[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q", results[0].Symbol)
	}
}

func TestSecurityDefinitions_Search_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.SecurityDefinitions().Search(context.Background(), "AAPL")
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
