package models

import (
	"encoding/json"
	"testing"
)

func TestSessionStatus_UnmarshalJSON(t *testing.T) {
	data := `{
		"authenticated": true,
		"connected": true,
		"competing": false,
		"established": true,
		"fail": "",
		"MAC": "AA:BB:CC",
		"hardware_info": "hw1",
		"serverName": "srv1",
		"serverVersion": "1.0"
	}`
	var s SessionStatus
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		t.Fatal(err)
	}
	if !s.Authenticated {
		t.Error("Authenticated should be true")
	}
	if !s.Connected {
		t.Error("Connected should be true")
	}
	if s.ServerName != "srv1" {
		t.Errorf("ServerName = %q", s.ServerName)
	}
	if s.MAC != "AA:BB:CC" {
		t.Errorf("MAC = %q", s.MAC)
	}
}

func TestSessionStatus_Partial(t *testing.T) {
	data := `{"authenticated": true}`
	var s SessionStatus
	if err := json.Unmarshal([]byte(data), &s); err != nil {
		t.Fatal(err)
	}
	if !s.Authenticated {
		t.Error("Authenticated should be true")
	}
	if s.ServerName != "" {
		t.Errorf("ServerName = %q, want empty", s.ServerName)
	}
}

func TestAccountList_UnmarshalJSON(t *testing.T) {
	data := `{
		"accounts": ["U1234567", "U7654321"],
		"selectedAccount": "U1234567",
		"allowFeatures": {
			"allowCrypto": true,
			"allowedAssetTypes": "STK,OPT"
		}
	}`
	var al AccountList
	if err := json.Unmarshal([]byte(data), &al); err != nil {
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
	data := `{
		"accountready": true,
		"accounttype": "INDIVIDUAL",
		"accountId": "U1234567",
		"currency": "USD"
	}`
	var as AccountSummary
	if err := json.Unmarshal([]byte(data), &as); err != nil {
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
	data := `{
		"amount": 50000.0,
		"currency": "USD",
		"isNull": false,
		"severity": 0,
		"timestamp": 1700000000,
		"value": "50000"
	}`
	var sf SummaryField
	if err := json.Unmarshal([]byte(data), &sf); err != nil {
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

func TestPortfolioSummaryField_UnmarshalJSON(t *testing.T) {
	data := `{
		"amount": 25000.0,
		"currency": "EUR",
		"severity": 1
	}`
	var psf PortfolioSummaryField
	if err := json.Unmarshal([]byte(data), &psf); err != nil {
		t.Fatal(err)
	}
	if psf.Amount != 25000.0 {
		t.Errorf("Amount = %f", psf.Amount)
	}
	if psf.Currency != "EUR" {
		t.Errorf("Currency = %q", psf.Currency)
	}
}

func TestPnLEntry_UnmarshalJSON(t *testing.T) {
	data := `{"dpl": 100.5, "nl": 50000.0, "upl": 250.0, "rpl": 50.0, "el": 0, "mv": 75000}`
	var e PnLEntry
	if err := json.Unmarshal([]byte(data), &e); err != nil {
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

func TestPosition_UnmarshalJSON(t *testing.T) {
	data := `{
		"acctId": "U1234567",
		"conid": 265598,
		"contractDesc": "AAPL",
		"position": 100.0,
		"mktPrice": 175.50,
		"mktValue": 17550.0,
		"avgCost": 150.0,
		"avgPrice": 150.0,
		"realizedPnl": 500.0,
		"unrealizedPnl": 2550.0,
		"currency": "USD",
		"assetClass": "STK",
		"ticker": "AAPL"
	}`
	var p Position
	if err := json.Unmarshal([]byte(data), &p); err != nil {
		t.Fatal(err)
	}
	if p.AcctID != "U1234567" {
		t.Errorf("AcctID = %q", p.AcctID)
	}
	if p.ConID != 265598 {
		t.Errorf("ConID = %d", p.ConID)
	}
	if p.Qty != 100.0 {
		t.Errorf("Qty = %f", p.Qty)
	}
	if p.MktPrice != 175.50 {
		t.Errorf("MktPrice = %f", p.MktPrice)
	}
	if p.Currency != "USD" {
		t.Errorf("Currency = %q", p.Currency)
	}
	if p.Ticker != "AAPL" {
		t.Errorf("Ticker = %q", p.Ticker)
	}
}

func TestLedgerEntry_UnmarshalJSON(t *testing.T) {
	data := `{
		"commoditymarketvalue": 0,
		"futureoptionvalue": 0,
		"futuresnlvalue": 0,
		"interest": 5.0,
		"netliquidationvalue": 50000.0,
		"realizedpnl": 200.0,
		"settledcash": 10000.0,
		"stockmarketvalue": 40000.0,
		"totalcashvalue": 10000.0,
		"unrealizedpnl": 500.0,
		"currency": "USD",
		"key": "LedgerList"
	}`
	var le LedgerEntry
	if err := json.Unmarshal([]byte(data), &le); err != nil {
		t.Fatal(err)
	}
	if le.NetLiquidation != 50000.0 {
		t.Errorf("NetLiquidation = %f", le.NetLiquidation)
	}
	if le.TotalCashValue != 10000.0 {
		t.Errorf("TotalCashValue = %f", le.TotalCashValue)
	}
	if le.Currency != "USD" {
		t.Errorf("Currency = %q", le.Currency)
	}
	if le.Key != "LedgerList" {
		t.Errorf("Key = %q", le.Key)
	}
	if le.Interest != 5.0 {
		t.Errorf("Interest = %f", le.Interest)
	}
}
