package flex

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseActivityStatement(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "activity_statement.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := ParseActivityStatement(f)
	if err != nil {
		t.Fatalf("ParseActivityStatement: %v", err)
	}

	if resp.QueryName != "MidwatchActivityQuery" {
		t.Errorf("QueryName = %q, want %q", resp.QueryName, "MidwatchActivityQuery")
	}
	if resp.Type != "AF" {
		t.Errorf("Type = %q, want %q", resp.Type, "AF")
	}
	if len(resp.Statements) != 1 {
		t.Fatalf("Statements count = %d, want 1", len(resp.Statements))
	}

	stmt := resp.Statements[0]
	if stmt.AccountID != "U1234567" {
		t.Errorf("AccountID = %q, want %q", stmt.AccountID, "U1234567")
	}
	if stmt.FromDate != "2026-01-01" {
		t.Errorf("FromDate = %q, want %q", stmt.FromDate, "2026-01-01")
	}

	// Trades
	if len(stmt.Trades) != 3 {
		t.Fatalf("Trades count = %d, want 3", len(stmt.Trades))
	}
	tr := stmt.Trades[0]
	if tr.TransactionID != "TXN001" {
		t.Errorf("Trade[0].TransactionID = %q, want %q", tr.TransactionID, "TXN001")
	}
	if tr.ConID != 265598 {
		t.Errorf("Trade[0].ConID = %d, want %d", tr.ConID, 265598)
	}
	if tr.Symbol != "AAPL" {
		t.Errorf("Trade[0].Symbol = %q, want %q", tr.Symbol, "AAPL")
	}
	if tr.TradePrice != 175.50 {
		t.Errorf("Trade[0].TradePrice = %f, want %f", tr.TradePrice, 175.50)
	}
	if tr.IBCommission != -1.00 {
		t.Errorf("Trade[0].IBCommission = %f, want %f", tr.IBCommission, -1.00)
	}

	// Option trade
	otr := stmt.Trades[1]
	if otr.AssetCategory != "OPT" {
		t.Errorf("Trade[1].AssetCategory = %q, want %q", otr.AssetCategory, "OPT")
	}
	if otr.Strike != 180.0 {
		t.Errorf("Trade[1].Strike = %f, want %f", otr.Strike, 180.0)
	}
	if otr.PutCall != "C" {
		t.Errorf("Trade[1].PutCall = %q, want %q", otr.PutCall, "C")
	}
	if otr.UnderlyingConID != 265598 {
		t.Errorf("Trade[1].UnderlyingConID = %d, want %d", otr.UnderlyingConID, 265598)
	}
	if otr.Multiplier != 100 {
		t.Errorf("Trade[1].Multiplier = %f, want %f", otr.Multiplier, 100.0)
	}

	// Closing trade with realized PnL
	ctr := stmt.Trades[2]
	if ctr.FIFOPnlRealized != 269.70 {
		t.Errorf("Trade[2].FIFOPnlRealized = %f, want %f", ctr.FIFOPnlRealized, 269.70)
	}
	if ctr.OpenCloseIndicator != "C" {
		t.Errorf("Trade[2].OpenCloseIndicator = %q, want %q", ctr.OpenCloseIndicator, "C")
	}

	// Cash Transactions
	if len(stmt.CashTransactions) != 3 {
		t.Fatalf("CashTransactions count = %d, want 3", len(stmt.CashTransactions))
	}
	ct := stmt.CashTransactions[0]
	if ct.TransactionID != "CTX001" {
		t.Errorf("CashTransaction[0].TransactionID = %q, want %q", ct.TransactionID, "CTX001")
	}
	if ct.Type != "Dividends" {
		t.Errorf("CashTransaction[0].Type = %q, want %q", ct.Type, "Dividends")
	}
	if ct.Amount != 25.00 {
		t.Errorf("CashTransaction[0].Amount = %f, want %f", ct.Amount, 25.00)
	}

	// Margin interest
	mi := stmt.CashTransactions[1]
	if mi.Type != "Broker Interest Paid" {
		t.Errorf("CashTransaction[1].Type = %q, want %q", mi.Type, "Broker Interest Paid")
	}
	if mi.Amount != -12.50 {
		t.Errorf("CashTransaction[1].Amount = %f, want %f", mi.Amount, -12.50)
	}

	// Option Events
	if len(stmt.OptionEvents) != 1 {
		t.Fatalf("OptionEvents count = %d, want 1", len(stmt.OptionEvents))
	}
	oe := stmt.OptionEvents[0]
	if oe.TransactionType != "Assignment" {
		t.Errorf("OptionEvent[0].TransactionType = %q, want %q", oe.TransactionType, "Assignment")
	}
	if oe.Strike != 180.0 {
		t.Errorf("OptionEvent[0].Strike = %f, want %f", oe.Strike, 180.0)
	}
	if oe.Proceeds != 18000.00 {
		t.Errorf("OptionEvent[0].Proceeds = %f, want %f", oe.Proceeds, 18000.00)
	}

	// Commission Details
	if len(stmt.CommissionDetails) != 1 {
		t.Fatalf("CommissionDetails count = %d, want 1", len(stmt.CommissionDetails))
	}
	cd := stmt.CommissionDetails[0]
	if cd.BrokerExecutionCharge != 0.50 {
		t.Errorf("CommissionDetail[0].BrokerExecutionCharge = %f, want %f", cd.BrokerExecutionCharge, 0.50)
	}
	if cd.RegFINRATradingActivityFee != 0.0119 {
		t.Errorf("CommissionDetail[0].RegFINRATradingActivityFee = %f, want %f", cd.RegFINRATradingActivityFee, 0.0119)
	}
}

func TestParseActivityStatement_Empty(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "activity_statement_empty.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := ParseActivityStatement(f)
	if err != nil {
		t.Fatalf("ParseActivityStatement: %v", err)
	}

	stmt := resp.Statements[0]
	if len(stmt.Trades) != 0 {
		t.Errorf("Trades count = %d, want 0", len(stmt.Trades))
	}
	if len(stmt.CashTransactions) != 0 {
		t.Errorf("CashTransactions count = %d, want 0", len(stmt.CashTransactions))
	}
	if len(stmt.OptionEvents) != 0 {
		t.Errorf("OptionEvents count = %d, want 0", len(stmt.OptionEvents))
	}
	if len(stmt.CommissionDetails) != 0 {
		t.Errorf("CommissionDetails count = %d, want 0", len(stmt.CommissionDetails))
	}
}

func TestParseActivityStatement_MissingAccountID(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="" fromDate="2026-01-01" toDate="2026-03-20" period="YTD" whenGenerated="2026-03-20;19:30:00">
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`

	_, err := ParseActivityStatement(strings.NewReader(xml))
	if err == nil {
		t.Fatal("expected error for missing accountId")
	}
}

func TestParseActivityStatement_NoStatements(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="0">
  </FlexStatements>
</FlexQueryResponse>`

	_, err := ParseActivityStatement(strings.NewReader(xml))
	if err == nil {
		t.Fatal("expected error for empty statements")
	}
}

func TestParseSendRequestResponse_Success(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "send_request_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := parseSendRequestResponse(f)
	if err != nil {
		t.Fatalf("parseSendRequestResponse: %v", err)
	}
	if resp.Status != "Success" {
		t.Errorf("Status = %q, want %q", resp.Status, "Success")
	}
	if resp.ReferenceCode != "1234567890" {
		t.Errorf("ReferenceCode = %q, want %q", resp.ReferenceCode, "1234567890")
	}
	if resp.ErrorCode != 0 {
		t.Errorf("ErrorCode = %d, want 0", resp.ErrorCode)
	}
}

func TestParseSendRequestResponse_Error(t *testing.T) {
	f, err := os.Open(filepath.Join("testdata", "send_request_token_expired.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	resp, err := parseSendRequestResponse(f)
	if err != nil {
		t.Fatalf("parseSendRequestResponse: %v", err)
	}
	if resp.Status != "Fail" {
		t.Errorf("Status = %q, want %q", resp.Status, "Fail")
	}
	if resp.ErrorCode != 1012 {
		t.Errorf("ErrorCode = %d, want 1012", resp.ErrorCode)
	}
	if resp.ErrorMessage != "Token has expired." {
		t.Errorf("ErrorMessage = %q, want %q", resp.ErrorMessage, "Token has expired.")
	}
}
