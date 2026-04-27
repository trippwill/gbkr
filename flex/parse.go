package flex

import (
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/trippwill/gbkr/num"
)

// FieldError describes a single field that failed to parse during wire → domain mapping.
type FieldError struct {
	Type  string // e.g. "Trade", "CashTransaction"
	Index int
	Field string
	Raw   string
	Err   error
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%s[%d].%s: failed to parse %q: %v", e.Type, e.Index, e.Field, e.Raw, e.Err)
}

func (e FieldError) Unwrap() error { return ErrFieldParse }

// FieldErrors collects multiple field parse failures.
type FieldErrors []FieldError

func (fe FieldErrors) Error() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d field parse error(s):", len(fe))
	for _, e := range fe {
		b.WriteString("\n  ")
		b.WriteString(e.Error())
	}
	return b.String()
}

func (fe FieldErrors) Unwrap() error { return ErrFieldParse }

// ParseActivityStatement decodes an Activity Flex Query XML response from r
// into a [QueryResponse]. Missing optional sections (Trades, CashTransactions,
// etc.) produce empty slices rather than errors.
//
// The response must contain at least one FlexStatement with a non-empty AccountID.
func ParseActivityStatement(r io.Reader) (*QueryResponse, error) {
	// Phase 1: XML → wire types
	var wire xmlQueryResponse
	if err := xml.NewDecoder(r).Decode(&wire); err != nil {
		return nil, fmt.Errorf("decode flex response: %w", err)
	}

	// Phase 2: wire → domain
	resp := &QueryResponse{
		QueryName: wire.QueryName,
		Type:      wire.Type,
	}

	for i, ws := range wire.Statements {
		stmt, err := mapStatement(ws)
		if err != nil {
			return nil, fmt.Errorf("statement %d: %w", i, err)
		}
		resp.Statements = append(resp.Statements, stmt)
	}

	if len(resp.Statements) == 0 {
		return nil, fmt.Errorf("flex response contains no statements")
	}

	for i, s := range resp.Statements {
		if s.AccountID == "" {
			return nil, fmt.Errorf("statement %d: missing accountId", i)
		}
	}

	// Phase 3: validate — check that all required Num fields parsed successfully
	if errs := validateResponse(resp); len(errs) > 0 {
		return nil, errs
	}

	return resp, nil
}

func mapStatement(ws xmlStatement) (Statement, error) {
	stmt := Statement{
		AccountID:     ws.AccountID,
		FromDate:      ws.FromDate,
		ToDate:        ws.ToDate,
		Period:        ws.Period,
		WhenGenerated: ws.WhenGenerated,
	}

	for i, wt := range ws.Trades {
		t, err := mapTrade(wt)
		if err != nil {
			return stmt, fmt.Errorf("trade %d: %w", i, err)
		}
		stmt.Trades = append(stmt.Trades, t)
	}

	for i, wc := range ws.CashTxns {
		ct, err := mapCashTransaction(wc)
		if err != nil {
			return stmt, fmt.Errorf("cash transaction %d: %w", i, err)
		}
		stmt.CashTransactions = append(stmt.CashTransactions, ct)
	}

	for i, wo := range ws.OptionEvents {
		oe, err := mapOptionEvent(wo)
		if err != nil {
			return stmt, fmt.Errorf("option event %d: %w", i, err)
		}
		stmt.OptionEvents = append(stmt.OptionEvents, oe)
	}

	for i, wcd := range ws.Commissions {
		cd, err := mapCommissionDetail(wcd)
		if err != nil {
			return stmt, fmt.Errorf("commission detail %d: %w", i, err)
		}
		stmt.CommissionDetails = append(stmt.CommissionDetails, cd)
	}

	return stmt, nil
}

func mapTrade(w xmlTrade) (Trade, error) {
	conID, err := parseInt64(w.ConID, "conid")
	if err != nil {
		return Trade{}, err
	}

	underlyingID, err := parseInt64(w.UnderlyingConID, "underlyingConid")
	if err != nil {
		return Trade{}, err
	}

	return Trade{
		TransactionID: w.TransactionID,
		TradeID:       w.TradeID,
		OrderID:       w.IBOrderID,
		ExecID:        w.IBExecID,
		AccountID:     w.AccountID,
		ConID:         conID,
		Symbol:        w.Symbol,
		Underlying:    w.UnderlyingSymbol,
		UnderlyingID:  underlyingID,
		AssetClass:    w.AssetCategory,
		Side:          w.BuySell,
		Quantity:      parseNum(w.Quantity),
		Price:         parseNum(w.TradePrice),
		TradeMoney:    parseNum(w.TradeMoney),
		Proceeds:      parseNum(w.Proceeds),
		Commission:    parseNum(w.IBCommission),
		Taxes:         parseNum(w.Taxes),
		NetCash:       parseNullNum(w.NetCash),
		CostBasis:     parseNullNum(w.CostBasis),
		RealizedPnL:   parseNullNum(w.FIFOPnlRealized),
		Strike:        parseNullNum(w.Strike),
		Expiry:        w.Expiry,
		PutCall:       w.PutCall,
		OpenClose:     w.OpenCloseIndicator,
		OrderRef:      w.OrderReference,
		Currency:      w.Currency,
		Multiplier:    parseNum(w.Multiplier),
		TradeDate:     w.TradeDate,
		SettleDate:    w.SettleDate,
	}, nil
}

func mapCashTransaction(w xmlCashTransaction) (CashTransaction, error) {
	conID, err := parseInt64(w.ConID, "conid")
	if err != nil {
		return CashTransaction{}, err
	}

	return CashTransaction{
		TransactionID: w.TransactionID,
		AccountID:     w.AccountID,
		ConID:         conID,
		Symbol:        w.Symbol,
		Type:          w.Type,
		Amount:        parseNum(w.Amount),
		Currency:      w.Currency,
		Description:   w.Description,
		ReportDate:    w.ReportDate,
		SettleDate:    w.SettleDate,
	}, nil
}

func mapOptionEvent(w xmlOptionEvent) (OptionEvent, error) {
	conID, err := parseInt64(w.ConID, "conid")
	if err != nil {
		return OptionEvent{}, err
	}

	underlyingID, err := parseInt64(w.UnderlyingConID, "underlyingConid")
	if err != nil {
		return OptionEvent{}, err
	}

	return OptionEvent{
		TransactionType: w.TransactionType,
		AccountID:       w.AccountID,
		ConID:           conID,
		Symbol:          w.Symbol,
		Underlying:      w.UnderlyingSymbol,
		UnderlyingID:    underlyingID,
		Strike:          parseNum(w.Strike),
		Expiry:          w.Expiry,
		PutCall:         w.PutCall,
		Quantity:        parseNum(w.Quantity),
		Proceeds:        parseNum(w.Proceeds),
		RealizedPnL:     parseNum(w.RealizedPnl),
		TradeDate:       w.TradeDate,
		Currency:        w.Currency,
		Multiplier:      parseNum(w.Multiplier),
	}, nil
}

func mapCommissionDetail(w xmlCommissionDetail) (CommissionDetail, error) {
	conID, err := parseInt64(w.ConID, "conid")
	if err != nil {
		return CommissionDetail{}, err
	}

	return CommissionDetail{
		AccountID:                  w.AccountID,
		ConID:                      conID,
		Symbol:                     w.Symbol,
		TradeID:                    w.TradeID,
		ExecID:                     w.ExecID,
		BrokerExecutionCharge:      parseNum(w.BrokerExecutionCharge),
		BrokerClearingCharge:       parseNum(w.BrokerClearingCharge),
		ThirdPartyExecutionCharge:  parseNum(w.ThirdPartyExecutionCharge),
		RegFINRATradingActivityFee: parseNum(w.RegFINRATradingActivityFee),
		RegSection31TransactionFee: parseNum(w.RegSection31TransactionFee),
		Currency:                   w.Currency,
		TradeDate:                  w.TradeDate,
	}, nil
}

// parseNum trims whitespace and parses s as a Num.
// XML attributes may carry incidental whitespace; trimming here keeps
// num.FromString strict while the flex layer is lenient.
func parseNum(s string) num.Num {
	return num.FromString(strings.TrimSpace(s))
}

// parseInt64 parses s as a base-10 int64. Empty string and "0" both return 0.
func parseInt64(s, field string) (int64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" {
		return 0, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse %s %q: %w", field, s, err)
	}
	return v, nil
}

// parseNullNum maps an XML string to a NullNum.
// Empty/whitespace-only → NullNum{Valid: false}; non-empty → NullNum{Num: parsed, Valid: true}.
func parseNullNum(s string) num.NullNum {
	s = strings.TrimSpace(s)
	if s == "" {
		return num.NullNum{}
	}
	return num.NullNum{Num: num.FromString(s), Valid: true}
}

// validateResponse checks all required Num fields for parse errors.
func validateResponse(resp *QueryResponse) FieldErrors {
	var errs FieldErrors
	for _, stmt := range resp.Statements {
		for i, t := range stmt.Trades {
			checkNum(&errs, "Trade", i, "Quantity", t.Quantity)
			checkNum(&errs, "Trade", i, "Price", t.Price)
			checkNum(&errs, "Trade", i, "TradeMoney", t.TradeMoney)
			checkNum(&errs, "Trade", i, "Proceeds", t.Proceeds)
			checkNum(&errs, "Trade", i, "Commission", t.Commission)
			checkNum(&errs, "Trade", i, "Taxes", t.Taxes)
			checkNullNum(&errs, "Trade", i, "NetCash", t.NetCash)
			checkNum(&errs, "Trade", i, "Multiplier", t.Multiplier)
			checkNullNum(&errs, "Trade", i, "CostBasis", t.CostBasis)
			checkNullNum(&errs, "Trade", i, "RealizedPnL", t.RealizedPnL)
			checkNullNum(&errs, "Trade", i, "Strike", t.Strike)
		}
		for i, ct := range stmt.CashTransactions {
			checkNum(&errs, "CashTransaction", i, "Amount", ct.Amount)
		}
		for i, oe := range stmt.OptionEvents {
			checkNum(&errs, "OptionEvent", i, "Strike", oe.Strike)
			checkNum(&errs, "OptionEvent", i, "Quantity", oe.Quantity)
			checkNum(&errs, "OptionEvent", i, "Proceeds", oe.Proceeds)
			checkNum(&errs, "OptionEvent", i, "RealizedPnL", oe.RealizedPnL)
			checkNum(&errs, "OptionEvent", i, "Multiplier", oe.Multiplier)
		}
		for i, cd := range stmt.CommissionDetails {
			checkNum(&errs, "CommissionDetail", i, "BrokerExecutionCharge", cd.BrokerExecutionCharge)
			checkNum(&errs, "CommissionDetail", i, "BrokerClearingCharge", cd.BrokerClearingCharge)
			checkNum(&errs, "CommissionDetail", i, "ThirdPartyExecutionCharge", cd.ThirdPartyExecutionCharge)
			checkNum(&errs, "CommissionDetail", i, "RegFINRATradingActivityFee", cd.RegFINRATradingActivityFee)
			checkNum(&errs, "CommissionDetail", i, "RegSection31TransactionFee", cd.RegSection31TransactionFee)
		}
	}
	return errs
}

func checkNum(errs *FieldErrors, typ string, idx int, field string, n num.Num) {
	if n.Err != nil {
		*errs = append(*errs, FieldError{
			Type:  typ,
			Index: idx,
			Field: field,
			Err:   n.Err,
		})
	}
}

func checkNullNum(errs *FieldErrors, typ string, idx int, field string, nn num.NullNum) {
	if nn.Valid && nn.Num.Err != nil {
		*errs = append(*errs, FieldError{
			Type:  typ,
			Index: idx,
			Field: field,
			Err:   nn.Num.Err,
		})
	}
}

// parseSendRequestResponse decodes a SendRequest endpoint XML response.
func parseSendRequestResponse(r io.Reader) (*sendRequestResponse, error) {
	var resp sendRequestResponse
	dec := xml.NewDecoder(r)
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("decode send-request response: %w", err)
	}
	return &resp, nil
}
