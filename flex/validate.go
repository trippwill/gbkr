package flex

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
		return t.Quantity.IsZero()
	case "Price":
		return t.Price.IsZero()
	case "TradeMoney":
		return t.TradeMoney.IsZero()
	case "Proceeds":
		return t.Proceeds.IsZero()
	case "Commission":
		return t.Commission.IsZero()
	case "Taxes":
		return t.Taxes.IsZero()
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
		return t.Multiplier.IsZero()
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
		return ct.Amount.IsZero()
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
		return oe.Strike.IsZero()
	case "Expiry":
		return !oe.Expiry.Valid
	case "PutCall":
		return oe.PutCall == ""
	case "Quantity":
		return oe.Quantity.IsZero()
	case "Proceeds":
		return oe.Proceeds.IsZero()
	case "RealizedPnL":
		return oe.RealizedPnL.IsZero()
	case "TradeDate":
		return oe.TradeDate.IsZero()
	case "Currency":
		return oe.Currency == ""
	case "Multiplier":
		return oe.Multiplier.IsZero()
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
		return cd.BrokerExecutionCharge.IsZero()
	case "BrokerClearingCharge":
		return cd.BrokerClearingCharge.IsZero()
	case "ThirdPartyExecutionCharge":
		return cd.ThirdPartyExecutionCharge.IsZero()
	case "RegFINRATradingActivityFee":
		return cd.RegFINRATradingActivityFee.IsZero()
	case "RegSection31TransactionFee":
		return cd.RegSection31TransactionFee.IsZero()
	case "Currency":
		return cd.Currency == ""
	case "TradeDate":
		return cd.TradeDate.IsZero()
	default:
		return true
	}
}
