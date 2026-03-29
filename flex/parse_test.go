package flex

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/trippwill/gbkr/num"
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
	assertNum(t, "Trade[0].Price", tr.Price, "175.50")
	assertNum(t, "Trade[0].Commission", tr.Commission, "-1.00")
	if tr.OrderID != "ORD001" {
		t.Errorf("Trade[0].OrderID = %q, want %q", tr.OrderID, "ORD001")
	}
	if tr.ExecID != "EXEC001" {
		t.Errorf("Trade[0].ExecID = %q, want %q", tr.ExecID, "EXEC001")
	}
	if tr.Side != "BUY" {
		t.Errorf("Trade[0].Side = %q, want %q", tr.Side, "BUY")
	}
	if tr.AssetClass != "STK" {
		t.Errorf("Trade[0].AssetClass = %q, want %q", tr.AssetClass, "STK")
	}

	// NullNum fields: stock trade has empty strike/expiry
	if tr.Strike.Valid {
		t.Errorf("Trade[0].Strike should be invalid (empty), got Valid=true")
	}

	// CostBasis is "17551.00" — should be valid
	if !tr.CostBasis.Valid {
		t.Errorf("Trade[0].CostBasis should be valid")
	} else {
		assertNum(t, "Trade[0].CostBasis.Num", tr.CostBasis.Num, "17551.00")
	}

	// Option trade
	otr := stmt.Trades[1]
	if otr.AssetClass != "OPT" {
		t.Errorf("Trade[1].AssetClass = %q, want %q", otr.AssetClass, "OPT")
	}
	if !otr.Strike.Valid {
		t.Fatalf("Trade[1].Strike should be valid")
	}
	assertNum(t, "Trade[1].Strike.Num", otr.Strike.Num, "180")
	if otr.PutCall != "C" {
		t.Errorf("Trade[1].PutCall = %q, want %q", otr.PutCall, "C")
	}
	if otr.UnderlyingID != 265598 {
		t.Errorf("Trade[1].UnderlyingID = %d, want %d", otr.UnderlyingID, 265598)
	}
	assertNum(t, "Trade[1].Multiplier", otr.Multiplier, "100")

	// Closing trade with realized PnL
	ctr := stmt.Trades[2]
	if !ctr.RealizedPnL.Valid {
		t.Fatalf("Trade[2].RealizedPnL should be valid")
	}
	assertNum(t, "Trade[2].RealizedPnL.Num", ctr.RealizedPnL.Num, "269.70")
	if ctr.OpenClose != "C" {
		t.Errorf("Trade[2].OpenClose = %q, want %q", ctr.OpenClose, "C")
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
	assertNum(t, "CashTransaction[0].Amount", ct.Amount, "25.00")

	// Margin interest
	mi := stmt.CashTransactions[1]
	if mi.Type != "Broker Interest Paid" {
		t.Errorf("CashTransaction[1].Type = %q, want %q", mi.Type, "Broker Interest Paid")
	}
	assertNum(t, "CashTransaction[1].Amount", mi.Amount, "-12.50")

	// Option Events
	if len(stmt.OptionEvents) != 1 {
		t.Fatalf("OptionEvents count = %d, want 1", len(stmt.OptionEvents))
	}
	oe := stmt.OptionEvents[0]
	if oe.TransactionType != "Assignment" {
		t.Errorf("OptionEvent[0].TransactionType = %q, want %q", oe.TransactionType, "Assignment")
	}
	assertNum(t, "OptionEvent[0].Strike", oe.Strike, "180")
	assertNum(t, "OptionEvent[0].Proceeds", oe.Proceeds, "18000.00")
	if oe.Underlying != "AAPL" {
		t.Errorf("OptionEvent[0].Underlying = %q, want %q", oe.Underlying, "AAPL")
	}
	if oe.UnderlyingID != 265598 {
		t.Errorf("OptionEvent[0].UnderlyingID = %d, want %d", oe.UnderlyingID, 265598)
	}

	// Commission Details
	if len(stmt.CommissionDetails) != 1 {
		t.Fatalf("CommissionDetails count = %d, want 1", len(stmt.CommissionDetails))
	}
	cd := stmt.CommissionDetails[0]
	assertNum(t, "CommissionDetail[0].BrokerExecutionCharge", cd.BrokerExecutionCharge, "0.50")
	assertNum(t, "CommissionDetail[0].RegFINRATradingActivityFee", cd.RegFINRATradingActivityFee, "0.0119")
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

// assertNum checks that a num.Num equals the expected decimal string representation.
func assertNum(t *testing.T, name string, got num.Num, wantStr string) {
	t.Helper()
	if !got.Ok() {
		t.Errorf("%s: Num has error: %v", name, got.Err)
		return
	}
	want := num.FromString(wantStr)
	if !want.Ok() {
		t.Fatalf("%s: bad test expectation %q: %v", name, wantStr, want.Err)
	}
	if !got.Equal(want) {
		t.Errorf("%s = %s, want %s", name, got, want)
	}
}

func TestParseActivityStatement_SidePassthrough(t *testing.T) {
	// IBKR Flex uses "BUY" and "SELL" for the buySell attribute.
	// Verify these pass through to the Side field unchanged.
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="U1" fromDate="2026-01-01" toDate="2026-01-01" period="D" whenGenerated="2026-01-01;19:30:00">
      <Trades>
        <Trade transactionID="T1" tradeID="T1" ibOrderID="" ibExecID="" accountId="U1" conid="1" symbol="X" underlyingSymbol="" underlyingConid="0" assetCategory="STK" buySell="BUY" quantity="1" tradePrice="1" tradeMoney="1" proceeds="1" ibCommission="0" taxes="0" netCash="1" fifoPnlRealized="" costBasis="" strike="" expiry="" putCall="" openCloseIndicator="" orderReference="" tradeDate="2026-01-01" settleDate="" currency="USD" multiplier="1" />
        <Trade transactionID="T2" tradeID="T2" ibOrderID="" ibExecID="" accountId="U1" conid="1" symbol="X" underlyingSymbol="" underlyingConid="0" assetCategory="STK" buySell="SELL" quantity="-1" tradePrice="2" tradeMoney="2" proceeds="2" ibCommission="0" taxes="0" netCash="2" fifoPnlRealized="" costBasis="" strike="" expiry="" putCall="" openCloseIndicator="" orderReference="" tradeDate="2026-01-01" settleDate="" currency="USD" multiplier="1" />
      </Trades>
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`

	resp, err := ParseActivityStatement(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("ParseActivityStatement: %v", err)
	}

	if resp.Statements[0].Trades[0].Side != "BUY" {
		t.Errorf("Trade[0].Side = %q, want %q", resp.Statements[0].Trades[0].Side, "BUY")
	}
	if resp.Statements[0].Trades[1].Side != "SELL" {
		t.Errorf("Trade[1].Side = %q, want %q", resp.Statements[0].Trades[1].Side, "SELL")
	}
}

func TestParseActivityStatement_NullNumEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		strike    string
		wantValid bool
		wantZero  bool
	}{
		{"empty string → invalid", "", false, false},
		{"zero string → valid zero", "0", true, true},
		{"zero decimal → valid zero", "0.00", true, true},
		{"positive value → valid", "180.50", true, false},
		{"negative value → valid", "-10.25", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			xmlData := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="U1" fromDate="2026-01-01" toDate="2026-01-01" period="D" whenGenerated="2026-01-01;19:30:00">
      <Trades>
        <Trade transactionID="T1" tradeID="T1" ibOrderID="" ibExecID="" accountId="U1" conid="1" symbol="X" underlyingSymbol="" underlyingConid="0" assetCategory="STK" buySell="BUY" quantity="1" tradePrice="1" tradeMoney="1" proceeds="1" ibCommission="0" taxes="0" netCash="1" fifoPnlRealized="" costBasis="" strike="%s" expiry="" putCall="" openCloseIndicator="" orderReference="" tradeDate="2026-01-01" settleDate="" currency="USD" multiplier="1" />
      </Trades>
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`, tt.strike)

			resp, err := ParseActivityStatement(strings.NewReader(xmlData))
			if err != nil {
				t.Fatalf("ParseActivityStatement: %v", err)
			}

			strike := resp.Statements[0].Trades[0].Strike
			if strike.Valid != tt.wantValid {
				t.Errorf("Strike.Valid = %v, want %v", strike.Valid, tt.wantValid)
			}
			if tt.wantValid && tt.wantZero && !strike.Num.IsZero() {
				t.Errorf("Strike.Num should be zero, got %s", strike.Num)
			}
		})
	}
}

func TestParseActivityStatement_InvalidNumericField(t *testing.T) {
	// A non-numeric value in a required Num field should produce a FieldErrors.
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="U1" fromDate="2026-01-01" toDate="2026-01-01" period="D" whenGenerated="2026-01-01;19:30:00">
      <Trades>
        <Trade transactionID="T1" tradeID="T1" ibOrderID="" ibExecID="" accountId="U1" conid="1" symbol="X" underlyingSymbol="" underlyingConid="0" assetCategory="STK" buySell="BUY" quantity="not_a_number" tradePrice="1" tradeMoney="1" proceeds="1" ibCommission="0" taxes="0" netCash="1" fifoPnlRealized="" costBasis="" strike="" expiry="" putCall="" openCloseIndicator="" orderReference="" tradeDate="2026-01-01" settleDate="" currency="USD" multiplier="1" />
      </Trades>
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`

	_, err := ParseActivityStatement(strings.NewReader(xmlData))
	if err == nil {
		t.Fatal("expected error for invalid numeric field")
	}

	var fieldErrs FieldErrors
	if !errors.As(err, &fieldErrs) {
		t.Fatalf("expected FieldErrors, got %T: %v", err, err)
	}
	if len(fieldErrs) != 1 {
		t.Fatalf("expected 1 field error, got %d: %v", len(fieldErrs), fieldErrs)
	}
	if fieldErrs[0].Field != "Quantity" {
		t.Errorf("field error field = %q, want %q", fieldErrs[0].Field, "Quantity")
	}
	if !errors.Is(err, ErrFieldParse) {
		t.Errorf("error should wrap ErrFieldParse")
	}
}

func TestParseActivityStatement_LargeValues(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="U1" fromDate="2026-01-01" toDate="2026-01-01" period="D" whenGenerated="2026-01-01;19:30:00">
      <Trades>
        <Trade transactionID="T1" tradeID="T1" ibOrderID="" ibExecID="" accountId="U1" conid="999999999" symbol="X" underlyingSymbol="" underlyingConid="0" assetCategory="STK" buySell="BUY" quantity="1000000" tradePrice="99999.99" tradeMoney="99999990000.00" proceeds="-99999990000.00" ibCommission="-500.00" taxes="0" netCash="-99999990500.00" fifoPnlRealized="" costBasis="" strike="" expiry="" putCall="" openCloseIndicator="" orderReference="" tradeDate="2026-01-01" settleDate="" currency="USD" multiplier="1" />
      </Trades>
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`

	resp, err := ParseActivityStatement(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("ParseActivityStatement: %v", err)
	}

	tr := resp.Statements[0].Trades[0]
	assertNum(t, "Quantity", tr.Quantity, "1000000")
	assertNum(t, "Price", tr.Price, "99999.99")
	if tr.ConID != 999999999 {
		t.Errorf("ConID = %d, want %d", tr.ConID, 999999999)
	}
}

func TestParseActivityStatement_ZeroValueNumerics(t *testing.T) {
	xmlData := `<?xml version="1.0" encoding="UTF-8"?>
<FlexQueryResponse queryName="Q" type="AF">
  <FlexStatements count="1">
    <FlexStatement accountId="U1" fromDate="2026-01-01" toDate="2026-01-01" period="D" whenGenerated="2026-01-01;19:30:00">
      <Trades>
        <Trade transactionID="T1" tradeID="T1" ibOrderID="" ibExecID="" accountId="U1" conid="1" symbol="X" underlyingSymbol="" underlyingConid="0" assetCategory="STK" buySell="BUY" quantity="0" tradePrice="0.00" tradeMoney="0" proceeds="0" ibCommission="0" taxes="0" netCash="0" fifoPnlRealized="0" costBasis="0" strike="" expiry="" putCall="" openCloseIndicator="" orderReference="" tradeDate="2026-01-01" settleDate="" currency="USD" multiplier="1" />
      </Trades>
    </FlexStatement>
  </FlexStatements>
</FlexQueryResponse>`

	resp, err := ParseActivityStatement(strings.NewReader(xmlData))
	if err != nil {
		t.Fatalf("ParseActivityStatement: %v", err)
	}

	tr := resp.Statements[0].Trades[0]
	if !tr.Quantity.IsZero() {
		t.Errorf("Quantity should be zero, got %s", tr.Quantity)
	}
	if !tr.Price.IsZero() {
		t.Errorf("Price should be zero, got %s", tr.Price)
	}

	// RealizedPnL "0" → NullNum{Valid: true, Num: zero}
	if !tr.RealizedPnL.Valid {
		t.Errorf("RealizedPnL should be valid for '0'")
	} else if !tr.RealizedPnL.Num.IsZero() {
		t.Errorf("RealizedPnL should be zero, got %s", tr.RealizedPnL.Num)
	}

	// CostBasis "0" → NullNum{Valid: true, Num: zero}
	if !tr.CostBasis.Valid {
		t.Errorf("CostBasis should be valid for '0'")
	} else if !tr.CostBasis.Num.IsZero() {
		t.Errorf("CostBasis should be zero, got %s", tr.CostBasis.Num)
	}
}
