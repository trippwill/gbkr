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
	// NullNum/NullDate fields: Ok() == false means zero.
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
