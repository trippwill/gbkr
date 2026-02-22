package gbkr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTransactionHistory_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithPermissions(PermissionSet{}))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.TransactionHistory()
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestTransactionHistory_CallsPAEndpoint(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()))
	if err != nil {
		t.Fatal(err)
	}

	tr, err := c.TransactionHistory()
	if err != nil {
		t.Fatal(err)
	}

	result, err := tr.TransactionHistory(context.Background(), "U1234567", 265598, 30)
	if err != nil {
		t.Fatal(err)
	}
	if result.Currency != "USD" {
		t.Errorf("Currency = %q, want %q", result.Currency, "USD")
	}
}

func TestTrades_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithPermissions(PermissionSet{}))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}

	_, err = bc.Trades()
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()))
	if err != nil {
		t.Fatal(err)
	}

	bc := &BrokerageClient{Client: c}
	tr, err := bc.Trades()
	if err != nil {
		t.Fatal(err)
	}

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

func TestTrades_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithPermissions(PermissionSet{}))
	if err != nil {
		t.Fatal(err)
	}

	bc := &BrokerageClient{Client: c}
	_, err = bc.Trades()
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}
