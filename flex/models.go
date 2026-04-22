package flex

import "github.com/trippwill/gbkr/num"

// QueryResponse is the top-level envelope returned by the Flex Web Service
// GetStatement endpoint for a successfully generated report.
type QueryResponse struct {
	QueryName  string      // Name of the Flex Query template that generated this report
	Type       string      // Report type code (e.g., "AF" for Activity Flex)
	Statements []Statement // One statement per account included in the query
}

// Statement contains the report data for a single account.
type Statement struct {
	AccountID     string // IBKR account identifier (e.g., "U1234567")
	FromDate      string // Report start date
	ToDate        string // Report end date
	Period        string // Report period label (e.g., "YTD", "Q1", "LastMonth")
	WhenGenerated string // Timestamp when the report was generated

	Trades            []Trade
	CashTransactions  []CashTransaction
	OptionEvents      []OptionEvent
	CommissionDetails []CommissionDetail
}

// Trade represents a single execution from the Trades section of an
// Activity Flex Query.
//
// Financial fields use [num.Num] for exact decimal arithmetic. Fields that may
// be absent from the Flex XML (depending on query template configuration) use
// [num.NullNum]. Field names use domain vocabulary; the original IBKR Flex
// attribute name is noted where they differ.
type Trade struct {
	TransactionID string      // IBKR: transactionID
	TradeID       string      // IBKR: tradeID
	OrderID       string      // IBKR: ibOrderID
	ExecID        string      // IBKR: ibExecID
	AccountID     string      // IBKR: accountId
	ConID         int64       // IBKR: conid — contract identifier
	Symbol        string      // IBKR: symbol
	Underlying    string      // IBKR: underlyingSymbol — empty for non-derivative trades
	UnderlyingID  int64       // IBKR: underlyingConid — 0 for non-derivative trades
	AssetClass    string      // IBKR: assetCategory — "STK", "OPT", etc.
	Side          string      // IBKR: buySell — "BUY" or "SELL"
	Quantity      num.Num     // Signed: positive for buys, negative for sells
	Price         num.Num     // IBKR: tradePrice — per-unit execution price
	TradeMoney    num.Num     // IBKR: tradeMoney — quantity × price
	Proceeds      num.Num     // Net proceeds (negative for buys)
	Commission    num.Num     // IBKR: ibCommission — typically negative
	Taxes         num.Num     // Regulatory taxes
	NetCash       num.NullNum // Net cash impact; absent if not in query template
	CostBasis     num.NullNum // FIFO cost basis; absent if not computed
	RealizedPnL   num.NullNum // IBKR: fifoPnlRealized — FIFO realized P&L
	Strike        num.NullNum // Option strike price; absent for stock trades
	Expiry        string      // Option expiry date; empty for stock trades
	PutCall       string      // "C" or "P" for options; empty for stock trades
	OpenClose     string      // IBKR: openCloseIndicator — "O" or "C"
	OrderRef      string      // IBKR: orderReference — user-assigned order ref
	Currency      string      // Settlement currency (e.g., "USD")
	Multiplier    num.Num     // Contract multiplier (1 for stock, 100 for US equity options)
	TradeDate     string      // Execution date
	SettleDate    string      // Settlement date; may be empty
}

// CashTransaction represents a dividend, interest charge, fee, or other
// cash movement from the Cash Transactions section of an Activity Flex Query.
type CashTransaction struct {
	TransactionID string  // IBKR: transactionID — may be empty on SUMMARY rows
	AccountID     string  // IBKR: accountId
	ConID         int64   // IBKR: conid — 0 for account-level entries (e.g., margin interest)
	Symbol        string  // IBKR: symbol — empty for account-level entries
	Type          string  // IBKR: type — "Dividends", "Withholding Tax", "Broker Interest Paid", etc.
	Amount        num.Num // Signed: positive for credits, negative for debits
	Currency      string  // Settlement currency
	Description   string  // IBKR: description — human-readable details
	ReportDate    string  // Date the transaction was reported
	SettleDate    string  // Settlement date
}

// OptionEvent represents an option exercise, assignment, or expiration
// from the Option Exercises, Assignments & Expirations (OptionEAE) section.
type OptionEvent struct {
	TransactionType string  // "Exercise", "Assignment", or "Expiration"
	AccountID       string  // IBKR: accountId
	ConID           int64   // IBKR: conid — option contract identifier
	Symbol          string  // IBKR: symbol — OCC-style option symbol
	Underlying      string  // IBKR: underlyingSymbol
	UnderlyingID    int64   // IBKR: underlyingConid
	Strike          num.Num // Option strike price
	Expiry          string  // Option expiry date
	PutCall         string  // "C" or "P"
	Quantity        num.Num // Signed: positive for long, negative for short
	Proceeds        num.Num // Cash proceeds from the event
	RealizedPnL     num.Num // IBKR: realizedPnl
	TradeDate       string  // Date the event occurred
	Currency        string  // Settlement currency
	Multiplier      num.Num // Contract multiplier (typically 100)
}

// CommissionDetail provides a granular fee breakdown for a single trade
// from the Commission Details section. All charge fields are [num.Num]
// representing the dollar amount of each fee component.
type CommissionDetail struct {
	AccountID                  string  // IBKR: accountId
	ConID                      int64   // IBKR: conid
	Symbol                     string  // IBKR: symbol
	TradeID                    string  // IBKR: tradeID — links to [Trade.TradeID]
	ExecID                     string  // IBKR: execID — links to [Trade.ExecID]
	BrokerExecutionCharge      num.Num // Broker execution fee
	BrokerClearingCharge       num.Num // Broker clearing fee
	ThirdPartyExecutionCharge  num.Num // Exchange/third-party execution fee
	RegFINRATradingActivityFee num.Num // FINRA TAF
	RegSection31TransactionFee num.Num // SEC Section 31 fee
	Currency                   string  // Fee currency
	TradeDate                  string  // Trade execution date
}

// sendRequestResponse is the XML envelope returned by the SendRequest endpoint.
type sendRequestResponse struct {
	Status        string `xml:"Status"`
	ReferenceCode string `xml:"ReferenceCode"`
	URL           string `xml:"Url"`
	ErrorCode     int    `xml:"ErrorCode"`
	ErrorMessage  string `xml:"ErrorMessage"`
}
