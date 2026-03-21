package flex

// QueryResponse is the top-level XML envelope returned by the Flex Web
// Service GetStatement endpoint for a successfully generated report.
type QueryResponse struct {
	QueryName  string      `xml:"queryName,attr"`
	Type       string      `xml:"type,attr"`
	Statements []Statement `xml:"FlexStatements>FlexStatement"`
}

// Statement contains the report data for a single account.
type Statement struct {
	AccountID     string `xml:"accountId,attr"`
	FromDate      string `xml:"fromDate,attr"`
	ToDate        string `xml:"toDate,attr"`
	Period        string `xml:"period,attr"`
	WhenGenerated string `xml:"whenGenerated,attr"`

	Trades            []Trade            `xml:"Trades>Trade"`
	CashTransactions  []CashTransaction  `xml:"CashTransactions>CashTransaction"`
	OptionEvents      []OptionEvent      `xml:"OptionEAE>OptionEAE"`
	CommissionDetails []CommissionDetail `xml:"CommissionDetails>CommissionDetail"`
}

// Trade represents a single execution from the Trades section of an
// Activity Flex Query. Fields use IBKR's camelCase attribute names.
type Trade struct {
	TransactionID      string  `xml:"transactionID,attr"`
	TradeID            string  `xml:"tradeID,attr"`
	IBOrderID          string  `xml:"ibOrderID,attr"`
	IBExecID           string  `xml:"ibExecID,attr"`
	AccountID          string  `xml:"accountId,attr"`
	ConID              int64   `xml:"conid,attr"`
	Symbol             string  `xml:"symbol,attr"`
	UnderlyingSymbol   string  `xml:"underlyingSymbol,attr"`
	UnderlyingConID    int64   `xml:"underlyingConid,attr"`
	AssetCategory      string  `xml:"assetCategory,attr"`
	BuySell            string  `xml:"buySell,attr"`
	Quantity           float64 `xml:"quantity,attr"`
	TradePrice         float64 `xml:"tradePrice,attr"`
	TradeMoney         float64 `xml:"tradeMoney,attr"`
	Proceeds           float64 `xml:"proceeds,attr"`
	IBCommission       float64 `xml:"ibCommission,attr"`
	Taxes              float64 `xml:"taxes,attr"`
	NetCash            float64 `xml:"netCash,attr"`
	FIFOPnlRealized    float64 `xml:"fifoPnlRealized,attr"`
	CostBasis          float64 `xml:"costBasis,attr"`
	Strike             float64 `xml:"strike,attr"`
	Expiry             string  `xml:"expiry,attr"`
	PutCall            string  `xml:"putCall,attr"`
	OpenCloseIndicator string  `xml:"openCloseIndicator,attr"`
	OrderReference     string  `xml:"orderReference,attr"`
	TradeDate          string  `xml:"tradeDate,attr"`
	SettleDate         string  `xml:"settleDate,attr"`
	Currency           string  `xml:"currency,attr"`
	Multiplier         float64 `xml:"multiplier,attr"`
}

// CashTransaction represents a dividend, interest charge, fee, or other
// cash movement from the Cash Transactions section.
type CashTransaction struct {
	TransactionID string  `xml:"transactionID,attr"`
	AccountID     string  `xml:"accountId,attr"`
	ConID         int64   `xml:"conid,attr"`
	Symbol        string  `xml:"symbol,attr"`
	Type          string  `xml:"type,attr"`
	Amount        float64 `xml:"amount,attr"`
	Currency      string  `xml:"currency,attr"`
	Description   string  `xml:"description,attr"`
	ReportDate    string  `xml:"reportDate,attr"`
	SettleDate    string  `xml:"settleDate,attr"`
}

// OptionEvent represents an option exercise, assignment, or expiration
// from the Option Exercises, Assignments & Expirations section.
type OptionEvent struct {
	TransactionType  string  `xml:"transactionType,attr"`
	AccountID        string  `xml:"accountId,attr"`
	ConID            int64   `xml:"conid,attr"`
	Symbol           string  `xml:"symbol,attr"`
	UnderlyingSymbol string  `xml:"underlyingSymbol,attr"`
	UnderlyingConID  int64   `xml:"underlyingConid,attr"`
	Strike           float64 `xml:"strike,attr"`
	Expiry           string  `xml:"expiry,attr"`
	PutCall          string  `xml:"putCall,attr"`
	Quantity         float64 `xml:"quantity,attr"`
	Proceeds         float64 `xml:"proceeds,attr"`
	RealizedPnl      float64 `xml:"realizedPnl,attr"`
	TradeDate        string  `xml:"tradeDate,attr"`
	Currency         string  `xml:"currency,attr"`
	Multiplier       float64 `xml:"multiplier,attr"`
}

// CommissionDetail provides a granular fee breakdown for a single trade
// from the Commission Details section.
type CommissionDetail struct {
	AccountID                  string  `xml:"accountId,attr"`
	ConID                      int64   `xml:"conid,attr"`
	Symbol                     string  `xml:"symbol,attr"`
	TradeID                    string  `xml:"tradeID,attr"`
	ExecID                     string  `xml:"execID,attr"`
	BrokerExecutionCharge      float64 `xml:"brokerExecutionCharge,attr"`
	BrokerClearingCharge       float64 `xml:"brokerClearingCharge,attr"`
	ThirdPartyExecutionCharge  float64 `xml:"thirdPartyExecutionCharge,attr"`
	RegFINRATradingActivityFee float64 `xml:"regFINRATradingActivityFee,attr"`
	RegSection31TransactionFee float64 `xml:"regSection31TransactionFee,attr"`
	Currency                   string  `xml:"currency,attr"`
	TradeDate                  string  `xml:"tradeDate,attr"`
}

// sendRequestResponse is the XML envelope returned by the SendRequest endpoint.
type sendRequestResponse struct {
	Status        string `xml:"Status"`
	ReferenceCode string `xml:"ReferenceCode"`
	URL           string `xml:"Url"`
	ErrorCode     int    `xml:"ErrorCode"`
	ErrorMessage  string `xml:"ErrorMessage"`
}
