package models

import (
	"encoding/json"
	"testing"
)

func TestTradeExecution_UnmarshalJSON(t *testing.T) {
	data := `{
		"execution_id": "exec123",
		"symbol": "AAPL",
		"side": "B",
		"order_description": "Buy 100 AAPL",
		"trade_time": "20231211-18:00:49",
		"trade_time_r": 1702317649000,
		"size": 100,
		"price": "175.50",
		"order_ref": "ref1",
		"exchange": "SMART",
		"commission": "1.00",
		"net_amount": -17550.0,
		"account": "U1234567",
		"company_name": "Apple Inc",
		"contract_description_1": "AAPL",
		"sec_type": "STK",
		"listing_exchange": "NASDAQ",
		"conid": 265598
	}`
	var te TradeExecution
	if err := json.Unmarshal([]byte(data), &te); err != nil {
		t.Fatal(err)
	}
	if te.ExecutionID != "exec123" {
		t.Errorf("ExecutionID = %q", te.ExecutionID)
	}
	if te.Symbol != "AAPL" {
		t.Errorf("Symbol = %q", te.Symbol)
	}
	if te.Side != "B" {
		t.Errorf("Side = %q", te.Side)
	}
	if te.TradeTimeEpoch != 1702317649000 {
		t.Errorf("TradeTimeEpoch = %d", te.TradeTimeEpoch)
	}
	if te.Size != 100 {
		t.Errorf("Size = %f", te.Size)
	}
	if te.NetAmount != -17550.0 {
		t.Errorf("NetAmount = %f", te.NetAmount)
	}
	if te.Account != "U1234567" {
		t.Errorf("Account = %q", te.Account)
	}
	if te.ConID != 265598 {
		t.Errorf("ConID = %d", te.ConID)
	}
	if te.SecType != "STK" {
		t.Errorf("SecType = %q", te.SecType)
	}
}

func TestTradeExecution_Partial(t *testing.T) {
	data := `{"execution_id": "e1", "symbol": "MSFT"}`
	var te TradeExecution
	if err := json.Unmarshal([]byte(data), &te); err != nil {
		t.Fatal(err)
	}
	if te.ExecutionID != "e1" {
		t.Errorf("ExecutionID = %q", te.ExecutionID)
	}
	if te.Symbol != "MSFT" {
		t.Errorf("Symbol = %q", te.Symbol)
	}
	if te.ConID != 0 {
		t.Errorf("ConID = %d, want 0", te.ConID)
	}
}

func TestTransactionHistoryResponse_UnmarshalJSON(t *testing.T) {
	data := `{
		"currency": "USD",
		"from": 1700000000,
		"to": 1700100000,
		"includesRealTime": true,
		"transactions": [
			{
				"date": "2023-12-01",
				"cur": "USD",
				"fxRate": 1.0,
				"pr": 175.50,
				"qty": -50,
				"acctid": "U1234567",
				"amt": 8775.0,
				"conid": 265598,
				"type": "Sell",
				"desc": "Apple Inc"
			}
		]
	}`
	var resp TransactionHistoryResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Currency != "USD" {
		t.Errorf("Currency = %q", resp.Currency)
	}
	if resp.From != 1700000000 {
		t.Errorf("From = %d", resp.From)
	}
	if resp.To != 1700100000 {
		t.Errorf("To = %d", resp.To)
	}
	if !resp.IncludesRealTime {
		t.Error("IncludesRealTime should be true")
	}
	if len(resp.Transactions) != 1 {
		t.Fatalf("Transactions = %d, want 1", len(resp.Transactions))
	}
	tx := resp.Transactions[0]
	if tx.Date != "2023-12-01" {
		t.Errorf("Date = %q", tx.Date)
	}
	if tx.Qty != -50 {
		t.Errorf("Qty = %d", tx.Qty)
	}
	if tx.Amount != 8775.0 {
		t.Errorf("Amount = %f", tx.Amount)
	}
	if tx.ConID != 265598 {
		t.Errorf("ConID = %d", tx.ConID)
	}
	if tx.Type != "Sell" {
		t.Errorf("Type = %q", tx.Type)
	}
}

func TestTransactionHistoryResponse_NilTransactions(t *testing.T) {
	data := `{"currency": "EUR", "from": 0, "to": 0}`
	var resp TransactionHistoryResponse
	if err := json.Unmarshal([]byte(data), &resp); err != nil {
		t.Fatal(err)
	}
	if resp.Currency != "EUR" {
		t.Errorf("Currency = %q", resp.Currency)
	}
	if resp.Transactions != nil {
		t.Errorf("Transactions should be nil, got %v", resp.Transactions)
	}
}

func TestTransaction_UnmarshalJSON(t *testing.T) {
	data := `{
		"date": "2023-11-15",
		"cur": "GBP",
		"fxRate": 0.79,
		"pr": 200.0,
		"qty": 25,
		"acctid": "U9999999",
		"amt": 5000.0,
		"conid": 12345,
		"type": "Buy",
		"desc": "Test Corp"
	}`
	var tx Transaction
	if err := json.Unmarshal([]byte(data), &tx); err != nil {
		t.Fatal(err)
	}
	if tx.Date != "2023-11-15" {
		t.Errorf("Date = %q", tx.Date)
	}
	if tx.Currency != "GBP" {
		t.Errorf("Currency = %q", tx.Currency)
	}
	if tx.FxRate != 0.79 {
		t.Errorf("FxRate = %f", tx.FxRate)
	}
	if tx.Price != 200.0 {
		t.Errorf("Price = %f", tx.Price)
	}
	if tx.Qty != 25 {
		t.Errorf("Qty = %d", tx.Qty)
	}
	if tx.Account != "U9999999" {
		t.Errorf("Account = %q", tx.Account)
	}
	if tx.Desc != "Test Corp" {
		t.Errorf("Desc = %q", tx.Desc)
	}
}
