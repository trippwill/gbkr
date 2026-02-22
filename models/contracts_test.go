package models

import (
	"encoding/json"
	"testing"
)

func TestContractDetails_UnmarshalJSON(t *testing.T) {
	data := `{
		"con_id": 265598,
		"symbol": "AAPL",
		"instrument_type": "STK",
		"exchange": "NASDAQ",
		"company_name": "Apple Inc",
		"currency": "USD",
		"multiplier": 1.0,
		"strike": 0,
		"expiry": "",
		"right": "",
		"und_conid": 0
	}`
	var cd ContractDetails
	if err := json.Unmarshal([]byte(data), &cd); err != nil {
		t.Fatal(err)
	}
	if cd.ConID != 265598 {
		t.Errorf("ConID = %d", cd.ConID)
	}
	if cd.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", cd.Symbol)
	}
	if cd.SecType != "STK" {
		t.Errorf("SecType = %q", cd.SecType)
	}
	if cd.Exchange != "NASDAQ" {
		t.Errorf("Exchange = %q", cd.Exchange)
	}
	if cd.CompanyName != "Apple Inc" {
		t.Errorf("CompanyName = %q", cd.CompanyName)
	}
	if cd.Currency != "USD" {
		t.Errorf("Currency = %q", cd.Currency)
	}
}

func TestContractDetails_Option(t *testing.T) {
	data := `{
		"con_id": 999,
		"symbol": "AAPL",
		"instrument_type": "OPT",
		"exchange": "CBOE",
		"company_name": "Apple Inc",
		"currency": "USD",
		"multiplier": 100.0,
		"strike": 190.0,
		"expiry": "20240119",
		"right": "C",
		"und_conid": 265598
	}`
	var cd ContractDetails
	if err := json.Unmarshal([]byte(data), &cd); err != nil {
		t.Fatal(err)
	}
	if cd.Strike != 190.0 {
		t.Errorf("Strike = %f", cd.Strike)
	}
	if cd.Expiry != "20240119" {
		t.Errorf("Expiry = %q", cd.Expiry)
	}
	if cd.PutOrCall != "C" {
		t.Errorf("PutOrCall = %q", cd.PutOrCall)
	}
	if cd.UndConID != 265598 {
		t.Errorf("UndConID = %d", cd.UndConID)
	}
	if cd.Multiplier != 100.0 {
		t.Errorf("Multiplier = %f", cd.Multiplier)
	}
}

func TestContractSearchResult_UnmarshalJSON(t *testing.T) {
	data := `{
		"conid": 265598,
		"companyName": "Apple Inc",
		"symbol": "AAPL",
		"secType": "STK"
	}`
	var r ContractSearchResult
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		t.Fatal(err)
	}
	if r.ConID != 265598 {
		t.Errorf("ConID = %d", r.ConID)
	}
	if r.CompanyName != "Apple Inc" {
		t.Errorf("CompanyName = %q", r.CompanyName)
	}
	if r.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", r.Symbol)
	}
	if r.SecType != "STK" {
		t.Errorf("SecType = %q", r.SecType)
	}
}

func TestContractSearchResult_Partial(t *testing.T) {
	data := `{"conid": 123}`
	var r ContractSearchResult
	if err := json.Unmarshal([]byte(data), &r); err != nil {
		t.Fatal(err)
	}
	if r.ConID != 123 {
		t.Errorf("ConID = %d", r.ConID)
	}
	if r.Symbol != "" {
		t.Errorf("Symbol = %q, want empty", r.Symbol)
	}
}
