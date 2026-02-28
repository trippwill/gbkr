package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestPortfolioAccounts_List(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/portfolio/accounts" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{
				"accountId":    "U1234567",
				"accountTitle": "John Smith",
				"desc":         "Primary",
				"currency":     "USD",
				"type":         "Individual",
				"isProp":       false,
				"isMulti":      false,
			},
			{
				"accountId": "U7654321",
				"currency":  "EUR",
				"type":      "IRA",
			},
		})
	})
	defer srv.Close()

	result, err := c.PortfolioAccounts().List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Fatalf("got %d accounts, want 2", len(result))
	}
	if result[0].AccountID != "U1234567" {
		t.Errorf("AccountID = %q", result[0].AccountID)
	}
	if result[0].Title != "John Smith" {
		t.Errorf("Title = %q", result[0].Title)
	}
	if result[0].Currency != "USD" {
		t.Errorf("Currency = %q", result[0].Currency)
	}
	if result[0].Type != "Individual" {
		t.Errorf("Type = %q", result[0].Type)
	}
	if result[1].AccountID != "U7654321" {
		t.Errorf("[1] AccountID = %q", result[1].AccountID)
	}
}

func TestPortfolioAccount_UnmarshalJSON(t *testing.T) {
	data := `{
		"accountId": "U1234567",
		"accountTitle": "Test Account",
		"desc": "Cash Account",
		"currency": "USD",
		"type": "Individual",
		"isProp": false,
		"isMulti": true
	}`
	var a PortfolioAccount
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		t.Fatal(err)
	}
	if a.AccountID != "U1234567" {
		t.Errorf("AccountID = %q", a.AccountID)
	}
	if a.Title != "Test Account" {
		t.Errorf("Title = %q", a.Title)
	}
	if a.Desc != "Cash Account" {
		t.Errorf("Desc = %q", a.Desc)
	}
	if a.Currency != "USD" {
		t.Errorf("Currency = %q", a.Currency)
	}
	if a.Type != "Individual" {
		t.Errorf("Type = %q", a.Type)
	}
	if a.IsProp {
		t.Error("IsProp should be false")
	}
	if !a.IsMulti {
		t.Error("IsMulti should be true")
	}
}

func TestPortfolioAccount_Partial(t *testing.T) {
	data := `{"accountId": "U999"}`
	var a PortfolioAccount
	if err := json.Unmarshal([]byte(data), &a); err != nil {
		t.Fatal(err)
	}
	if a.AccountID != "U999" {
		t.Errorf("AccountID = %q", a.AccountID)
	}
	if a.Title != "" {
		t.Errorf("Title = %q, want empty", a.Title)
	}
	if a.Currency != "" {
		t.Errorf("Currency = %q, want empty", a.Currency)
	}
}
