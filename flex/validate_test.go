package flex

import (
	"sort"
	"testing"

	"github.com/trippwill/gbkr/num"
	"github.com/trippwill/gbkr/when"
)

func TestValidate_OK_AllSectionsPresent(t *testing.T) {
	stmt := Statement{
		Trades: []Trade{
			{TransactionID: "T1", ConID: 12345, Symbol: "AAPL", TradeDate: when.NewDate(2026, 1, 15)},
		},
		CashTransactions: []CashTransaction{
			{TransactionID: "CT1", Type: "Dividends", Amount: num.FromInt64(100), ReportDate: when.NewDate(2026, 1, 15)},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades":           {"TransactionID", "ConID", "Symbol"},
			"CashTransactions": {"TransactionID", "Type", "Amount"},
		},
	}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK, got MissingSections=%v, EmptyFields=%v", result.MissingSections, result.EmptyFields)
	}
}

func TestValidate_MissingSections(t *testing.T) {
	stmt := Statement{
		Trades: []Trade{
			{TransactionID: "T1", Symbol: "AAPL"},
		},
		// No CashTransactions, no OptionEvents
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades":           {"TransactionID"},
			"CashTransactions": {"TransactionID"},
			"OptionEvents":     {"Symbol"},
		},
	}

	result := stmt.Validate(required)
	if result.OK() {
		t.Fatal("expected not OK")
	}

	sort.Strings(result.MissingSections)
	if len(result.MissingSections) != 2 {
		t.Fatalf("MissingSections = %v, want 2 entries", result.MissingSections)
	}
	if result.MissingSections[0] != "CashTransactions" {
		t.Errorf("MissingSections[0] = %q, want CashTransactions", result.MissingSections[0])
	}
	if result.MissingSections[1] != "OptionEvents" {
		t.Errorf("MissingSections[1] = %q, want OptionEvents", result.MissingSections[1])
	}
}

func TestValidate_EmptyFields(t *testing.T) {
	stmt := Statement{
		Trades: []Trade{
			{TransactionID: "T1", ConID: 0, Symbol: ""}, // ConID zero, Symbol empty
			{TransactionID: "T2", ConID: 0, Symbol: ""}, // Both still zero
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades": {"TransactionID", "ConID", "Symbol"},
		},
	}

	result := stmt.Validate(required)
	if result.OK() {
		t.Fatal("expected not OK")
	}

	if len(result.MissingSections) != 0 {
		t.Errorf("MissingSections = %v, want empty", result.MissingSections)
	}

	fields := result.EmptyFields["Trades"]
	sort.Strings(fields)
	if len(fields) != 2 {
		t.Fatalf("EmptyFields[Trades] = %v, want [ConID, Symbol]", fields)
	}
	if fields[0] != "ConID" || fields[1] != "Symbol" {
		t.Errorf("EmptyFields[Trades] = %v, want [ConID, Symbol]", fields)
	}
}

func TestValidate_PartialData(t *testing.T) {
	// At least one record has ConID non-zero, so ConID should NOT appear in EmptyFields.
	stmt := Statement{
		Trades: []Trade{
			{TransactionID: "T1", ConID: 0, Symbol: "AAPL"},
			{TransactionID: "T2", ConID: 12345, Symbol: "AAPL"},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades": {"ConID", "Symbol"},
		},
	}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK (at least one record has non-zero ConID), got EmptyFields=%v", result.EmptyFields)
	}
}

func TestValidate_UnknownSection(t *testing.T) {
	stmt := Statement{
		Trades: []Trade{{TransactionID: "T1"}},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades":      {"TransactionID"},
			"NonExistent": {"Foo"},
		},
	}

	result := stmt.Validate(required)
	if result.OK() {
		t.Fatal("expected not OK (unknown section treated as missing)")
	}

	if len(result.MissingSections) != 1 || result.MissingSections[0] != "NonExistent" {
		t.Errorf("MissingSections = %v, want [NonExistent]", result.MissingSections)
	}
}

func TestValidate_UnknownField(t *testing.T) {
	stmt := Statement{
		Trades: []Trade{{TransactionID: "T1"}},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades": {"TransactionID", "BogusField"},
		},
	}

	result := stmt.Validate(required)
	if result.OK() {
		t.Fatal("expected not OK (unknown field treated as zero)")
	}

	fields := result.EmptyFields["Trades"]
	if len(fields) != 1 || fields[0] != "BogusField" {
		t.Errorf("EmptyFields[Trades] = %v, want [BogusField]", fields)
	}
}

func TestValidate_EmptyRequiredFields(t *testing.T) {
	stmt := Statement{}
	required := RequiredFields{}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK with no requirements, got MissingSections=%v, EmptyFields=%v",
			result.MissingSections, result.EmptyFields)
	}
}

func TestValidate_NullableFieldZero(t *testing.T) {
	// NullNum/NullDate fields: Valid == false means zero (absent).
	stmt := Statement{
		Trades: []Trade{
			{TransactionID: "T1", Strike: num.NullNum{}, Expiry: when.NullDate{}},
			{TransactionID: "T2", Strike: num.NullNum{}, Expiry: when.NullDate{}},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades": {"Strike", "Expiry"},
		},
	}

	result := stmt.Validate(required)
	if result.OK() {
		t.Fatal("expected not OK (nullable fields all invalid)")
	}

	fields := result.EmptyFields["Trades"]
	sort.Strings(fields)
	if len(fields) != 2 {
		t.Fatalf("EmptyFields[Trades] = %v, want [Expiry, Strike]", fields)
	}
}

func TestValidate_CashTransactions(t *testing.T) {
	stmt := Statement{
		CashTransactions: []CashTransaction{
			{TransactionID: "CT1", Type: "Dividends", Amount: num.FromInt64(50), Currency: "USD",
				ReportDate: when.NewDate(2026, 3, 1)},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"CashTransactions": {"TransactionID", "Type", "Amount", "Currency"},
		},
	}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK, got MissingSections=%v, EmptyFields=%v", result.MissingSections, result.EmptyFields)
	}
}

func TestValidate_OptionEvents(t *testing.T) {
	stmt := Statement{
		OptionEvents: []OptionEvent{
			{TransactionType: "Expiration", ConID: 99, Symbol: "SPY260320C500",
				Strike: num.FromInt64(500), TradeDate: when.NewDate(2026, 3, 20),
				Quantity: num.FromInt64(-1), Proceeds: num.FromInt64(0), RealizedPnL: num.FromInt64(0)},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"OptionEvents": {"TransactionType", "ConID", "Symbol", "Strike"},
		},
	}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK, got MissingSections=%v, EmptyFields=%v", result.MissingSections, result.EmptyFields)
	}
}

func TestValidate_CommissionDetails(t *testing.T) {
	stmt := Statement{
		CommissionDetails: []CommissionDetail{
			{AccountID: "U123", ConID: 42, Symbol: "AAPL", TradeID: "T1", ExecID: "E1",
				BrokerExecutionCharge: num.FromInt64(1), TradeDate: when.NewDate(2026, 1, 15)},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"CommissionDetails": {"AccountID", "ConID", "BrokerExecutionCharge"},
		},
	}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK, got MissingSections=%v, EmptyFields=%v", result.MissingSections, result.EmptyFields)
	}
}

// TestTradeFieldZero_AllFields exercises every switch branch in tradeFieldZero.
func TestTradeFieldZero_AllFields(t *testing.T) {
	// Build a Trade with all fields populated (non-zero).
	full := Trade{
		TransactionID: "T1",
		TradeID:       "TR1",
		OrderID:       "O1",
		ExecID:        "E1",
		AccountID:     "U123",
		ConID:         42,
		Symbol:        "AAPL",
		Underlying:    "AAPL",
		UnderlyingID:  42,
		AssetClass:    "STK",
		Side:          "BUY",
		Quantity:      num.FromInt64(100),
		Price:         num.FromInt64(150),
		TradeMoney:    num.FromInt64(15000),
		Proceeds:      num.FromInt64(-15000),
		Commission:    num.FromInt64(-1),
		Taxes:         num.FromInt64(0), // num.FromInt64(0) is still zero-value
		NetCash:       num.NullNum{Num: num.FromInt64(100), Valid: true},
		CostBasis:     num.NullNum{Num: num.FromInt64(100), Valid: true},
		RealizedPnL:   num.NullNum{Num: num.FromInt64(50), Valid: true},
		Strike:        num.NullNum{Num: num.FromInt64(150), Valid: true},
		Expiry:        when.NullDate{Date: when.NewDate(2026, 3, 20), Valid: true},
		PutCall:       "C",
		OpenClose:     "O",
		OrderRef:      "REF1",
		Currency:      "USD",
		Multiplier:    num.FromInt64(100),
		TradeDate:     when.NewDate(2026, 1, 15),
		SettleDate:    when.NullDate{Date: when.NewDate(2026, 1, 17), Valid: true},
	}

	fields := []string{
		"TransactionID", "TradeID", "OrderID", "ExecID", "AccountID",
		"ConID", "Symbol", "Underlying", "UnderlyingID", "AssetClass",
		"Side", "Quantity", "Price", "TradeMoney", "Proceeds",
		"Commission", "Taxes", "NetCash", "CostBasis", "RealizedPnL",
		"Strike", "Expiry", "PutCall", "OpenClose", "OrderRef",
		"Currency", "Multiplier", "TradeDate", "SettleDate",
	}

	// Num fields that use .IsZero() — an uninitialized Num{} is not Ok(),
	// so IsZero() returns false; skip these in the zero-struct check.
	numFields := map[string]bool{
		"Quantity": true, "Price": true, "TradeMoney": true,
		"Proceeds": true, "Commission": true, "Taxes": true, "Multiplier": true,
	}

	// Verify non-zero trade reports false for all known fields.
	for _, f := range fields {
		if f == "Taxes" {
			continue // FromInt64(0) is zero-valued for Num
		}
		if tradeFieldZero(&full, f) {
			t.Errorf("tradeFieldZero(&full, %q) = true, want false", f)
		}
	}

	// Verify zero-valued trade reports true for all known fields
	// (except Num fields whose zero-value struct is invalid, not zero).
	var zero Trade
	for _, f := range fields {
		if numFields[f] {
			continue
		}
		if !tradeFieldZero(&zero, f) {
			t.Errorf("tradeFieldZero(&zero, %q) = false, want true", f)
		}
	}

	// Num fields: verify that FromInt64(0) is zero.
	zeroNum := Trade{
		Quantity:   num.FromInt64(0),
		Price:      num.FromInt64(0),
		TradeMoney: num.FromInt64(0),
		Proceeds:   num.FromInt64(0),
		Commission: num.FromInt64(0),
		Taxes:      num.FromInt64(0),
		Multiplier: num.FromInt64(0),
	}
	for f := range numFields {
		if !tradeFieldZero(&zeroNum, f) {
			t.Errorf("tradeFieldZero(&zeroNum, %q) = false, want true", f)
		}
	}

	// Unknown field always returns true.
	if !tradeFieldZero(&full, "BogusField") {
		t.Error("tradeFieldZero for unknown field should return true")
	}
}

// TestCashTxFieldZero_AllFields exercises every switch branch in cashTxFieldZero.
func TestCashTxFieldZero_AllFields(t *testing.T) {
	full := CashTransaction{
		TransactionID: "CT1",
		AccountID:     "U123",
		ConID:         42,
		Symbol:        "AAPL",
		Type:          "Dividends",
		Amount:        num.FromInt64(100),
		Currency:      "USD",
		Description:   "Dividend payment",
		ReportDate:    when.NewDate(2026, 3, 1),
		SettleDate:    when.NewDate(2026, 3, 3),
	}

	fields := []string{
		"TransactionID", "AccountID", "ConID", "Symbol", "Type",
		"Amount", "Currency", "Description", "ReportDate", "SettleDate",
	}

	for _, f := range fields {
		if cashTxFieldZero(&full, f) {
			t.Errorf("cashTxFieldZero(&full, %q) = true, want false", f)
		}
	}

	// Amount is a Num field — uninitialized Num{} is not IsZero().
	var zero CashTransaction
	for _, f := range fields {
		if f == "Amount" {
			continue
		}
		if !cashTxFieldZero(&zero, f) {
			t.Errorf("cashTxFieldZero(&zero, %q) = false, want true", f)
		}
	}

	// Verify Amount with explicit zero Num.
	zeroAmt := CashTransaction{Amount: num.FromInt64(0)}
	if !cashTxFieldZero(&zeroAmt, "Amount") {
		t.Error("cashTxFieldZero(&zeroAmt, \"Amount\") = false, want true")
	}

	if !cashTxFieldZero(&full, "BogusField") {
		t.Error("cashTxFieldZero for unknown field should return true")
	}
}

// TestOptionEventFieldZero_AllFields exercises every switch branch in optionEventFieldZero.
func TestOptionEventFieldZero_AllFields(t *testing.T) {
	full := OptionEvent{
		TransactionType: "Expiration",
		AccountID:       "U123",
		ConID:           99,
		Symbol:          "SPY260320C500",
		Underlying:      "SPY",
		UnderlyingID:    78,
		Strike:          num.FromInt64(500),
		Expiry:          when.NullDate{Date: when.NewDate(2026, 3, 20), Valid: true},
		PutCall:         "C",
		Quantity:        num.FromInt64(-1),
		Proceeds:        num.FromInt64(100),
		RealizedPnL:     num.FromInt64(50),
		TradeDate:       when.NewDate(2026, 3, 20),
		Currency:        "USD",
		Multiplier:      num.FromInt64(100),
	}

	fields := []string{
		"TransactionType", "AccountID", "ConID", "Symbol", "Underlying",
		"UnderlyingID", "Strike", "Expiry", "PutCall", "Quantity",
		"Proceeds", "RealizedPnL", "TradeDate", "Currency", "Multiplier",
	}

	for _, f := range fields {
		if optionEventFieldZero(&full, f) {
			t.Errorf("optionEventFieldZero(&full, %q) = true, want false", f)
		}
	}

	// Num fields — uninitialized Num{} is not IsZero().
	oeNumFields := map[string]bool{
		"Strike": true, "Quantity": true, "Proceeds": true,
		"RealizedPnL": true, "Multiplier": true,
	}

	var zero OptionEvent
	for _, f := range fields {
		if oeNumFields[f] {
			continue
		}
		if !optionEventFieldZero(&zero, f) {
			t.Errorf("optionEventFieldZero(&zero, %q) = false, want true", f)
		}
	}

	// Verify Num fields with explicit zero values.
	zeroOE := OptionEvent{
		Strike:      num.FromInt64(0),
		Quantity:    num.FromInt64(0),
		Proceeds:    num.FromInt64(0),
		RealizedPnL: num.FromInt64(0),
		Multiplier:  num.FromInt64(0),
	}
	for f := range oeNumFields {
		if !optionEventFieldZero(&zeroOE, f) {
			t.Errorf("optionEventFieldZero(&zeroOE, %q) = false, want true", f)
		}
	}

	if !optionEventFieldZero(&full, "BogusField") {
		t.Error("optionEventFieldZero for unknown field should return true")
	}
}

// TestCommissionDetailFieldZero_AllFields exercises every switch branch in commissionDetailFieldZero.
func TestCommissionDetailFieldZero_AllFields(t *testing.T) {
	full := CommissionDetail{
		AccountID:                  "U123",
		ConID:                      42,
		Symbol:                     "AAPL",
		TradeID:                    "T1",
		ExecID:                     "E1",
		BrokerExecutionCharge:      num.FromInt64(1),
		BrokerClearingCharge:       num.FromInt64(2),
		ThirdPartyExecutionCharge:  num.FromInt64(3),
		RegFINRATradingActivityFee: num.FromInt64(4),
		RegSection31TransactionFee: num.FromInt64(5),
		Currency:                   "USD",
		TradeDate:                  when.NewDate(2026, 1, 15),
	}

	fields := []string{
		"AccountID", "ConID", "Symbol", "TradeID", "ExecID",
		"BrokerExecutionCharge", "BrokerClearingCharge",
		"ThirdPartyExecutionCharge", "RegFINRATradingActivityFee",
		"RegSection31TransactionFee", "Currency", "TradeDate",
	}

	for _, f := range fields {
		if commissionDetailFieldZero(&full, f) {
			t.Errorf("commissionDetailFieldZero(&full, %q) = true, want false", f)
		}
	}

	// Num fields — uninitialized Num{} is not IsZero().
	cdNumFields := map[string]bool{
		"BrokerExecutionCharge": true, "BrokerClearingCharge": true,
		"ThirdPartyExecutionCharge": true, "RegFINRATradingActivityFee": true,
		"RegSection31TransactionFee": true,
	}

	var zero CommissionDetail
	for _, f := range fields {
		if cdNumFields[f] {
			continue
		}
		if !commissionDetailFieldZero(&zero, f) {
			t.Errorf("commissionDetailFieldZero(&zero, %q) = false, want true", f)
		}
	}

	// Verify Num fields with explicit zero values.
	zeroCD := CommissionDetail{
		BrokerExecutionCharge:      num.FromInt64(0),
		BrokerClearingCharge:       num.FromInt64(0),
		ThirdPartyExecutionCharge:  num.FromInt64(0),
		RegFINRATradingActivityFee: num.FromInt64(0),
		RegSection31TransactionFee: num.FromInt64(0),
	}
	for f := range cdNumFields {
		if !commissionDetailFieldZero(&zeroCD, f) {
			t.Errorf("commissionDetailFieldZero(&zeroCD, %q) = false, want true", f)
		}
	}

	if !commissionDetailFieldZero(&full, "BogusField") {
		t.Error("commissionDetailFieldZero for unknown field should return true")
	}
}

// TestValidate_AllFieldsInAllSections validates all fields across all four
// section types through the Validate entry point, ensuring full integration
// coverage of the field-zero helpers.
func TestValidate_AllFieldsInAllSections(t *testing.T) {
	stmt := Statement{
		Trades: []Trade{
			{
				TransactionID: "T1", TradeID: "TR1", OrderID: "O1", ExecID: "E1",
				AccountID: "U123", ConID: 42, Symbol: "AAPL", Underlying: "AAPL",
				UnderlyingID: 42, AssetClass: "STK", Side: "BUY",
				Quantity: num.FromInt64(100), Price: num.FromInt64(150),
				TradeMoney: num.FromInt64(15000), Proceeds: num.FromInt64(-15000),
				Commission: num.FromInt64(-1), Taxes: num.FromInt64(1),
				NetCash:     num.NullNum{Num: num.FromInt64(100), Valid: true},
				CostBasis:   num.NullNum{Num: num.FromInt64(100), Valid: true},
				RealizedPnL: num.NullNum{Num: num.FromInt64(50), Valid: true},
				Strike:      num.NullNum{Num: num.FromInt64(150), Valid: true},
				Expiry:      when.NullDate{Date: when.NewDate(2026, 3, 20), Valid: true},
				PutCall:     "C", OpenClose: "O", OrderRef: "REF1", Currency: "USD",
				Multiplier: num.FromInt64(100), TradeDate: when.NewDate(2026, 1, 15),
				SettleDate: when.NullDate{Date: when.NewDate(2026, 1, 17), Valid: true},
			},
		},
		CashTransactions: []CashTransaction{
			{
				TransactionID: "CT1", AccountID: "U123", ConID: 42, Symbol: "AAPL",
				Type: "Dividends", Amount: num.FromInt64(100), Currency: "USD",
				Description: "Dividend", ReportDate: when.NewDate(2026, 3, 1),
				SettleDate: when.NewDate(2026, 3, 3),
			},
		},
		OptionEvents: []OptionEvent{
			{
				TransactionType: "Expiration", AccountID: "U123", ConID: 99,
				Symbol: "SPY260320C500", Underlying: "SPY", UnderlyingID: 78,
				Strike:  num.FromInt64(500),
				Expiry:  when.NullDate{Date: when.NewDate(2026, 3, 20), Valid: true},
				PutCall: "C", Quantity: num.FromInt64(-1), Proceeds: num.FromInt64(100),
				RealizedPnL: num.FromInt64(50), TradeDate: when.NewDate(2026, 3, 20),
				Currency: "USD", Multiplier: num.FromInt64(100),
			},
		},
		CommissionDetails: []CommissionDetail{
			{
				AccountID: "U123", ConID: 42, Symbol: "AAPL", TradeID: "T1",
				ExecID: "E1", BrokerExecutionCharge: num.FromInt64(1),
				BrokerClearingCharge:       num.FromInt64(2),
				ThirdPartyExecutionCharge:  num.FromInt64(3),
				RegFINRATradingActivityFee: num.FromInt64(4),
				RegSection31TransactionFee: num.FromInt64(5),
				Currency:                   "USD", TradeDate: when.NewDate(2026, 1, 15),
			},
		},
	}

	required := RequiredFields{
		Sections: map[string][]string{
			"Trades": {
				"TransactionID", "TradeID", "OrderID", "ExecID", "AccountID",
				"ConID", "Symbol", "Underlying", "UnderlyingID", "AssetClass",
				"Side", "Quantity", "Price", "TradeMoney", "Proceeds",
				"Commission", "Taxes", "NetCash", "CostBasis", "RealizedPnL",
				"Strike", "Expiry", "PutCall", "OpenClose", "OrderRef",
				"Currency", "Multiplier", "TradeDate", "SettleDate",
			},
			"CashTransactions": {
				"TransactionID", "AccountID", "ConID", "Symbol", "Type",
				"Amount", "Currency", "Description", "ReportDate", "SettleDate",
			},
			"OptionEvents": {
				"TransactionType", "AccountID", "ConID", "Symbol", "Underlying",
				"UnderlyingID", "Strike", "Expiry", "PutCall", "Quantity",
				"Proceeds", "RealizedPnL", "TradeDate", "Currency", "Multiplier",
			},
			"CommissionDetails": {
				"AccountID", "ConID", "Symbol", "TradeID", "ExecID",
				"BrokerExecutionCharge", "BrokerClearingCharge",
				"ThirdPartyExecutionCharge", "RegFINRATradingActivityFee",
				"RegSection31TransactionFee", "Currency", "TradeDate",
			},
		},
	}

	result := stmt.Validate(required)
	if !result.OK() {
		t.Errorf("expected OK with all fields populated, got MissingSections=%v, EmptyFields=%v",
			result.MissingSections, result.EmptyFields)
	}
}

func TestValidateResult_OK(t *testing.T) {
	tests := []struct {
		name   string
		result ValidationResult
		want   bool
	}{
		{"empty", ValidationResult{}, true},
		{"missing sections", ValidationResult{MissingSections: []string{"Trades"}}, false},
		{"empty fields", ValidationResult{EmptyFields: map[string][]string{"Trades": {"ConID"}}}, false},
		{"both", ValidationResult{
			MissingSections: []string{"OptionEvents"},
			EmptyFields:     map[string][]string{"Trades": {"ConID"}},
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.OK(); got != tt.want {
				t.Errorf("OK() = %v, want %v", got, tt.want)
			}
		})
	}
}
