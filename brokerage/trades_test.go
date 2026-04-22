package brokerage

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/trippwill/gbkr/num"
)

func TestTrades_Recent(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/account/trades" {
			t.Errorf("path = %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{ //nolint:errcheck
			{
				"execution_id": "exec1",
				"symbol":       "AAPL",
				"side":         "B",
			},
		})
	})
	defer srv.Close()

	trades, err := bc.Trades().Recent(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(trades) != 1 {
		t.Fatalf("got %d trades, want 1", len(trades))
	}
	if trades[0].Symbol != "AAPL" {
		t.Errorf("Symbol = %q", trades[0].Symbol)
	}
}

func TestTradeExecution_UnmarshalJSON(t *testing.T) {
	data := []byte(`{
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
	}`)
	var te TradeExecution
	if err := json.Unmarshal(data, &te); err != nil {
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
	if !te.Size.Equal(num.FromFloat64(100)) {
		t.Errorf("Size = %s", te.Size)
	}
	if !te.NetAmount.Equal(num.FromFloat64(-17550.0)) {
		t.Errorf("NetAmount = %s", te.NetAmount)
	}
	if te.Account != "U1234567" {
		t.Errorf("Account = %q", te.Account)
	}
	if te.ConID != 265598 {
		t.Errorf("ConID = %d", te.ConID)
	}
}

func TestTradeExecution_Partial(t *testing.T) {
	data := []byte(`{"execution_id": "e1", "symbol": "MSFT"}`)
	var te TradeExecution
	if err := json.Unmarshal(data, &te); err != nil {
		t.Fatal(err)
	}
	if te.ExecutionID != "e1" {
		t.Errorf("ExecutionID = %q", te.ExecutionID)
	}
	if te.ConID != 0 {
		t.Errorf("ConID = %d, want 0", te.ConID)
	}
}

func TestTrades_Recent_Error(t *testing.T) {
	bc, srv := newTestBrokerageClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer srv.Close()

	_, err := bc.Trades().Recent(context.Background(), 7)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}
