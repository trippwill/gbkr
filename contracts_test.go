package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}

	cr := bc.Contracts()

	info, err := cr.Info(context.Background(), 265598)
	if err != nil {
		t.Fatal(err)
	}
	if info.Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want AAPL", info.Symbol)
	}
}
