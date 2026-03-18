package brokerage

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/trippwill/gbkr"
)

func TestAccounts_List(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/accounts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"accounts":        []string{"U1234567", "U7654321"},
			"selectedAccount": "U1234567",
		})
	})
	defer srv.Close()

	result, err := bc.Accounts().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Accounts) != 2 {
		t.Errorf("got %d accounts, want 2", len(result.Accounts))
	}
	if result.SelectedAcct != "U1234567" {
		t.Errorf("SelectedAcct = %q", result.SelectedAcct)
	}
}

func TestAccounts_PnL(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/account/pnl/partitioned" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"acctPnl": map[string]any{
				"U1234567": map[string]any{"dpl": 100.5, "nl": 50000.0},
			},
		})
	})
	defer srv.Close()

	result, err := bc.Accounts().PnL(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := result.AcctPnL[gbkr.AccountID("U1234567")]
	if !ok {
		t.Fatal("missing account in PnL")
	}
	if entry.DailyPnL != 100.5 {
		t.Errorf("DailyPnL = %f, want 100.5", entry.DailyPnL)
	}
}

func TestAccount_Summary(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
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
	})
	defer srv.Close()

	ar := bc.Account("U1234567")
	if ar.ID() != "U1234567" {
		t.Errorf("ID() = %q", ar.ID())
	}

	result, err := ar.Summary(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.AccountReady {
		t.Error("expected AccountReady=true")
	}
	if result.Currency != "USD" {
		t.Errorf("Currency = %q", result.Currency)
	}
}

func TestAccountList_UnmarshalJSON_Error(t *testing.T) {
	var al AccountList
	if err := json.Unmarshal([]byte(`"not_an_object"`), &al); err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}

func TestAccountSummary_UnmarshalJSON_Error(t *testing.T) {
	var as AccountSummary
	if err := json.Unmarshal([]byte(`"not_an_object"`), &as); err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}

func TestSummaryField_UnmarshalJSON_Error(t *testing.T) {
	var sf SummaryField
	if err := json.Unmarshal([]byte(`"not_an_object"`), &sf); err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}

func TestPnLEntry_UnmarshalJSON_Error(t *testing.T) {
	var e PnLEntry
	if err := json.Unmarshal([]byte(`"not_an_object"`), &e); err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}

func TestPnLPartitioned_UnmarshalJSON_Error(t *testing.T) {
	var p PnLPartitioned
	if err := json.Unmarshal([]byte(`"not_an_object"`), &p); err == nil {
		t.Fatal("expected unmarshal error for invalid JSON")
	}
}

func TestAccountList_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"accounts": ["U1234567", "U7654321"],
		"selectedAccount": "U1234567",
		"allowFeatures": {
			"allowCrypto": true,
			"allowedAssetTypes": "STK,OPT"
		}
	}`)
	var al AccountList
	if err := json.Unmarshal(data, &al); err != nil {
		t.Fatal(err)
	}
	if len(al.Accounts) != 2 {
		t.Fatalf("Accounts = %d", len(al.Accounts))
	}
	if al.Accounts[0] != "U1234567" {
		t.Errorf("Accounts[0] = %q", al.Accounts[0])
	}
	if al.SelectedAcct != "U1234567" {
		t.Errorf("SelectedAcct = %q", al.SelectedAcct)
	}
	if !al.AllowFeatures.AllowCrypto {
		t.Error("AllowCrypto should be true")
	}
	if al.AllowFeatures.AllowedAssetTypes != "STK,OPT" {
		t.Errorf("AllowedAssetTypes = %q", al.AllowFeatures.AllowedAssetTypes)
	}
}

func TestAccountSummary_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"accountready": true,
		"accounttype": "INDIVIDUAL",
		"accountId": "U1234567",
		"currency": "USD"
	}`)
	var as AccountSummary
	if err := json.Unmarshal(data, &as); err != nil {
		t.Fatal(err)
	}
	if !as.AccountReady {
		t.Error("AccountReady should be true")
	}
	if as.AccountType != "INDIVIDUAL" {
		t.Errorf("AccountType = %q", as.AccountType)
	}
	if as.AccountID != "U1234567" {
		t.Errorf("AccountID = %q", as.AccountID)
	}
	if as.Currency != "USD" {
		t.Errorf("Currency = %q", as.Currency)
	}
}

func TestSummaryField_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
		"amount": 50000.0,
		"currency": "USD",
		"isNull": false,
		"severity": 0,
		"timestamp": 1700000000,
		"value": "50000"
	}`)
	var sf SummaryField
	if err := json.Unmarshal(data, &sf); err != nil {
		t.Fatal(err)
	}
	if sf.Amount != 50000.0 {
		t.Errorf("Amount = %f", sf.Amount)
	}
	if sf.Currency != "USD" {
		t.Errorf("Currency = %q", sf.Currency)
	}
	if sf.Value != "50000" {
		t.Errorf("Value = %q", sf.Value)
	}
}

func TestPnLEntry_UnmarshalJSON(t *testing.T) {
	data := []byte(`{"dpl": 100.5, "nl": 50000.0, "upl": 250.0, "rpl": 50.0, "el": 0, "mv": 75000}`)
	var e PnLEntry
	if err := json.Unmarshal(data, &e); err != nil {
		t.Fatal(err)
	}
	if e.DailyPnL != 100.5 {
		t.Errorf("DailyPnL = %f", e.DailyPnL)
	}
	if e.NetLiquidation != 50000.0 {
		t.Errorf("NetLiquidation = %f", e.NetLiquidation)
	}
	if e.MarginValue != 75000 {
		t.Errorf("MarginValue = %f", e.MarginValue)
	}
}

func TestAccounts_List_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.Accounts().List(context.Background())
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestAccounts_PnL_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.Accounts().PnL(context.Background())
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

func TestAccount_Summary_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.Account("U1234567").Summary(context.Background())
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
