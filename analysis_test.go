package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAnalysis_Transactions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/pa/transactions" {
			t.Errorf("path = %q, want /pa/transactions", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"currency":         "USD",
			"from":             1700000000,
			"to":               1700100000,
			"includesRealTime": true,
			"transactions":     []map[string]any{},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	ar := c.Analysis()

	result, err := ar.Transactions(context.Background(), "U1234567", 265598, 30)
	if err != nil {
		t.Fatal(err)
	}
	if result.Value.Currency != "USD" {
		t.Errorf("Currency = %q, want %q", result.Value.Currency, "USD")
	}
}

func TestAnalysis_GatewayAccess(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	ar := c.Analysis()
	if ar == nil {
		t.Fatal("expected non-nil Analysis")
	}
}
