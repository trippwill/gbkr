package flex

import "github.com/trippwill/gbkr/num"

// RequiredFields specifies which Statement sections and fields a consumer
// expects. Keys are section names matching Statement struct field names
// (e.g., "Trades", "CashTransactions"). Values are field names within
// that section's element type that must be non-zero in at least one record.
type RequiredFields struct {
	Sections map[string][]string
}

// ValidationResult holds the outcome of [Statement.Validate].
type ValidationResult struct {
	// MissingSections lists section names that have zero records.
	MissingSections []string

	// EmptyFields maps section names to field names that are zero-valued
	// across every record in the section.
	EmptyFields map[string][]string
}

// OK returns true when no missing sections and no empty fields were found.
func (vr ValidationResult) OK() bool {
	return len(vr.MissingSections) == 0 && len(vr.EmptyFields) == 0
}

// sectionAccessor provides a way to inspect a Statement section without reflection.
type sectionAccessor struct {
	length    func(s *Statement) int
	fieldZero func(s *Statement, field string) bool // true if zero-valued across all records
}

var sectionRegistry = map[string]sectionAccessor{
	"Trades": {
		length: func(s *Statement) int { return len(s.Trades) },
		fieldZero: func(s *Statement, field string) bool {
			for i := range s.Trades {
				if !tradeFieldZero(&s.Trades[i], field) {
					return false
				}
			}
			return true
		},
	},
	"CashTransactions": {
		length: func(s *Statement) int { return len(s.CashTransactions) },
		fieldZero: func(s *Statement, field string) bool {
			for i := range s.CashTransactions {
				if !cashTxFieldZero(&s.CashTransactions[i], field) {
					return false
				}
			}
			return true
		},
	},
	"OptionEvents": {
		length: func(s *Statement) int { return len(s.OptionEvents) },
		fieldZero: func(s *Statement, field string) bool {
			for i := range s.OptionEvents {
				if !optionEventFieldZero(&s.OptionEvents[i], field) {
					return false
				}
			}
			return true
		},
	},
	"CommissionDetails": {
		length: func(s *Statement) int { return len(s.CommissionDetails) },
		fieldZero: func(s *Statement, field string) bool {
			for i := range s.CommissionDetails {
				if !commissionDetailFieldZero(&s.CommissionDetails[i], field) {
					return false
				}
			}
			return true
		},
	},
}

// Validate checks whether the Statement contains the sections and fields
// described by required. A section with zero records is reported as missing.
// A field that is zero-valued in every record of a present section is
// reported as empty.
func (s *Statement) Validate(required RequiredFields) ValidationResult {
	var result ValidationResult

	for section, fields := range required.Sections {
		acc, known := sectionRegistry[section]
		if !known || acc.length(s) == 0 {
			result.MissingSections = append(result.MissingSections, section)
			continue
		}

		for _, field := range fields {
			if acc.fieldZero(s, field) {
				if result.EmptyFields == nil {
					result.EmptyFields = make(map[string][]string)
				}
				result.EmptyFields[section] = append(result.EmptyFields[section], field)
			}
		}
	}

	return result
}

// Field zero-check helpers — explicit switch on field name, no reflection.
// For num.Num fields, a value is considered "zero" if it is uninitialized (!Ok())
// or its numeric value is zero. This ensures that missing/invalid parse results
// are treated as empty rather than silently passing validation.

func numZero(n num.Num) bool { return !n.Ok() || n.IsZero() }

func tradeFieldZero(t *Trade, field string) bool {
	switch field {
	case "TransactionID":
		return t.TransactionID == ""
	case "TradeID":
		return t.TradeID == ""
	case "OrderID":
		return t.OrderID == ""
	case "ExecID":
		return t.ExecID == ""
	case "AccountID":
		return t.AccountID == ""
	case "ConID":
		return t.ConID == 0
	case "Symbol":
		return t.Symbol == ""
	case "Underlying":
		return t.Underlying == ""
	case "UnderlyingID":
		return t.UnderlyingID == 0
	case "AssetClass":
		return t.AssetClass == ""
	case "Side":
		return t.Side == ""
	case "Quantity":
		return numZero(t.Quantity)
	case "Price":
		return numZero(t.Price)
	case "TradeMoney":
		return numZero(t.TradeMoney)
	case "Proceeds":
		return numZero(t.Proceeds)
	case "Commission":
		return numZero(t.Commission)
	case "Taxes":
		return numZero(t.Taxes)
	case "NetCash":
		return !t.NetCash.Valid
	case "CostBasis":
		return !t.CostBasis.Valid
	case "RealizedPnL":
		return !t.RealizedPnL.Valid
	case "Strike":
		return !t.Strike.Valid
	case "Expiry":
		return !t.Expiry.Valid
	case "PutCall":
		return t.PutCall == ""
	case "OpenClose":
		return t.OpenClose == ""
	case "OrderRef":
		return t.OrderRef == ""
	case "Currency":
		return t.Currency == ""
	case "Multiplier":
		return numZero(t.Multiplier)
	case "TradeDate":
		return t.TradeDate.IsZero()
	case "SettleDate":
		return !t.SettleDate.Valid
	default:
		return true // unknown field treated as zero
	}
}

func cashTxFieldZero(ct *CashTransaction, field string) bool {
	switch field {
	case "TransactionID":
		return ct.TransactionID == ""
	case "AccountID":
		return ct.AccountID == ""
	case "ConID":
		return ct.ConID == 0
	case "Symbol":
		return ct.Symbol == ""
	case "Type":
		return ct.Type == ""
	case "Amount":
		return numZero(ct.Amount)
	case "Currency":
		return ct.Currency == ""
	case "Description":
		return ct.Description == ""
	case "ReportDate":
		return ct.ReportDate.IsZero()
	case "SettleDate":
		return ct.SettleDate.IsZero()
	default:
		return true
	}
}

func optionEventFieldZero(oe *OptionEvent, field string) bool {
	switch field {
	case "TransactionType":
		return oe.TransactionType == ""
	case "AccountID":
		return oe.AccountID == ""
	case "ConID":
		return oe.ConID == 0
	case "Symbol":
		return oe.Symbol == ""
	case "Underlying":
		return oe.Underlying == ""
	case "UnderlyingID":
		return oe.UnderlyingID == 0
	case "Strike":
		return numZero(oe.Strike)
	case "Expiry":
		return !oe.Expiry.Valid
	case "PutCall":
		return oe.PutCall == ""
	case "Quantity":
		return numZero(oe.Quantity)
	case "Proceeds":
		return numZero(oe.Proceeds)
	case "RealizedPnL":
		return numZero(oe.RealizedPnL)
	case "TradeDate":
		return oe.TradeDate.IsZero()
	case "Currency":
		return oe.Currency == ""
	case "Multiplier":
		return numZero(oe.Multiplier)
	default:
		return true
	}
}

func commissionDetailFieldZero(cd *CommissionDetail, field string) bool {
	switch field {
	case "AccountID":
		return cd.AccountID == ""
	case "ConID":
		return cd.ConID == 0
	case "Symbol":
		return cd.Symbol == ""
	case "TradeID":
		return cd.TradeID == ""
	case "ExecID":
		return cd.ExecID == ""
	case "BrokerExecutionCharge":
		return numZero(cd.BrokerExecutionCharge)
	case "BrokerClearingCharge":
		return numZero(cd.BrokerClearingCharge)
	case "ThirdPartyExecutionCharge":
		return numZero(cd.ThirdPartyExecutionCharge)
	case "RegFINRATradingActivityFee":
		return numZero(cd.RegFINRATradingActivityFee)
	case "RegSection31TransactionFee":
		return numZero(cd.RegSection31TransactionFee)
	case "Currency":
		return cd.Currency == ""
	case "TradeDate":
		return cd.TradeDate.IsZero()
	default:
		return true
	}
}
