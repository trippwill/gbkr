package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTrades_RecentTrades(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/account/trades" {
			t.Errorf("path = %q, want /iserver/account/trades", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{
				"execution_id": "exec1",
				"symbol":       "AAPL",
				"side":         "B",
			},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(ReadOnly()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	bc := &BrokerageClient{Client: c}
	tr := bc.Trades()

	trades, err := tr.RecentTrades(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	if trades[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q, want %q", trades[0].Symbol, "AAPL")
	}
}
