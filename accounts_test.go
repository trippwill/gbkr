package gbkr

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippwill/gbkr/models"
)

func TestAccounts_List(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/accounts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"accounts":        []string{"U1234567", "U7654321"},
			"selectedAccount": "U1234567",
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}
	al := bc.Accounts()

	result, err := al.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Accounts) != 2 {
		t.Errorf("got %d accounts, want 2", len(result.Accounts))
	}
	if result.SelectedAcct != "U1234567" {
		t.Errorf("SelectedAcct = %q, want %q", result.SelectedAcct, "U1234567")
	}
}

func TestAccounts_PnL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/account/pnl/partitioned" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"acctPnl": map[string]any{
				"U1234567": map[string]any{"dpl": 100.5, "nl": 50000.0},
			},
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}
	al := bc.Accounts()

	result, err := al.PnL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := result.AcctPnL[models.AccountID("U1234567")]
	if !ok {
		t.Fatal("missing account in PnL")
	}
	if entry.DailyPnL != 100.5 {
		t.Errorf("DailyPnL = %f, want 100.5", entry.DailyPnL)
	}
}

func TestAccount_Summary(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/account/U1234567/summary" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"accountready": true,
			"accounttype":  "INDIVIDUAL",
			"accountId":    "U1234567",
			"currency":     "USD",
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}
	ar := bc.Account("U1234567")

	if ar.ID() != "U1234567" {
		t.Errorf("ID() = %q, want %q", ar.ID(), "U1234567")
	}

	result, err := ar.Summary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.AccountReady {
		t.Error("expected AccountReady=true")
	}
	if result.Currency != "USD" {
		t.Errorf("Currency = %q, want %q", result.Currency, "USD")
	}
}

func TestAccounts_List_EmitsEvent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"accounts":        []string{"U1234567"},
			"selectedAccount": "U1234567",
		})
	}))
	defer srv.Close()

	h := newCaptureHandler()
	logger := slog.New(h)
	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil), WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	bc := &BrokerageClient{Client: c}

	_, err = bc.Accounts().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	if h.count() != 1 {
		t.Fatalf("expected 1 event, got %d", h.count())
	}

	r := h.last()
	if r.Level != slog.LevelInfo {
		t.Errorf("level = %v, want Info", r.Level)
	}

	var gotOp string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "op" {
			gotOp = a.Value.String()
		}
		return true
	})
	if gotOp != string(OpListAccounts) {
		t.Errorf("op = %q, want %q", gotOp, OpListAccounts)
	}
}
