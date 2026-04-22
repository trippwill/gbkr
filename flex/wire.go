package flex

// Wire types for XML deserialization of IBKR Flex Query responses.
// All fields are string because XML attributes deserialize as text;
// num.FromString handles numeric parsing uniformly in the mapping layer.
//
// The xml* prefix convention distinguishes wire types from their exported
// domain counterparts in models.go.

type xmlQueryResponse struct {
	QueryName  string         `xml:"queryName,attr"`
	Type       string         `xml:"type,attr"`
	Statements []xmlStatement `xml:"FlexStatements>FlexStatement"`
}

type xmlStatement struct {
	AccountID     string                `xml:"accountId,attr"`
	FromDate      string                `xml:"fromDate,attr"`
	ToDate        string                `xml:"toDate,attr"`
	Period        string                `xml:"period,attr"`
	WhenGenerated string                `xml:"whenGenerated,attr"`
	Trades        []xmlTrade            `xml:"Trades>Trade"`
	CashTxns      []xmlCashTransaction  `xml:"CashTransactions>CashTransaction"`
	OptionEvents  []xmlOptionEvent      `xml:"OptionEAE>OptionEAE"`
	Commissions   []xmlCommissionDetail `xml:"CommissionDetails>CommissionDetail"`
}

type xmlTrade struct {
	TransactionID      string `xml:"transactionID,attr"`
	TradeID            string `xml:"tradeID,attr"`
	IBOrderID          string `xml:"ibOrderID,attr"`
	IBExecID           string `xml:"ibExecID,attr"`
	AccountID          string `xml:"accountId,attr"`
	ConID              string `xml:"conid,attr"`
	Symbol             string `xml:"symbol,attr"`
	UnderlyingSymbol   string `xml:"underlyingSymbol,attr"`
	UnderlyingConID    string `xml:"underlyingConid,attr"`
	AssetCategory      string `xml:"assetCategory,attr"`
	BuySell            string `xml:"buySell,attr"`
	Quantity           string `xml:"quantity,attr"`
	TradePrice         string `xml:"tradePrice,attr"`
	TradeMoney         string `xml:"tradeMoney,attr"`
	Proceeds           string `xml:"proceeds,attr"`
	IBCommission       string `xml:"ibCommission,attr"`
	Taxes              string `xml:"taxes,attr"`
	NetCash            string `xml:"netCash,attr"`
	FIFOPnlRealized    string `xml:"fifoPnlRealized,attr"`
	CostBasis          string `xml:"costBasis,attr"`
	Strike             string `xml:"strike,attr"`
	Expiry             string `xml:"expiry,attr"`
	PutCall            string `xml:"putCall,attr"`
	OpenCloseIndicator string `xml:"openCloseIndicator,attr"`
	OrderReference     string `xml:"orderReference,attr"`
	TradeDate          string `xml:"tradeDate,attr"`
	SettleDate         string `xml:"settleDate,attr"`
	Currency           string `xml:"currency,attr"`
	Multiplier         string `xml:"multiplier,attr"`
}

type xmlCashTransaction struct {
	TransactionID string `xml:"transactionID,attr"`
	AccountID     string `xml:"accountId,attr"`
	ConID         string `xml:"conid,attr"`
	Symbol        string `xml:"symbol,attr"`
	Type          string `xml:"type,attr"`
	Amount        string `xml:"amount,attr"`
	Currency      string `xml:"currency,attr"`
	Description   string `xml:"description,attr"`
	ReportDate    string `xml:"reportDate,attr"`
	SettleDate    string `xml:"settleDate,attr"`
}

type xmlOptionEvent struct {
	TransactionType  string `xml:"transactionType,attr"`
	AccountID        string `xml:"accountId,attr"`
	ConID            string `xml:"conid,attr"`
	Symbol           string `xml:"symbol,attr"`
	UnderlyingSymbol string `xml:"underlyingSymbol,attr"`
	UnderlyingConID  string `xml:"underlyingConid,attr"`
	Strike           string `xml:"strike,attr"`
	Expiry           string `xml:"expiry,attr"`
	PutCall          string `xml:"putCall,attr"`
	Quantity         string `xml:"quantity,attr"`
	Proceeds         string `xml:"proceeds,attr"`
	RealizedPnl      string `xml:"realizedPnl,attr"`
	TradeDate        string `xml:"tradeDate,attr"`
	Currency         string `xml:"currency,attr"`
	Multiplier       string `xml:"multiplier,attr"`
}

type xmlCommissionDetail struct {
	AccountID                  string `xml:"accountId,attr"`
	ConID                      string `xml:"conid,attr"`
	Symbol                     string `xml:"symbol,attr"`
	TradeID                    string `xml:"tradeID,attr"`
	ExecID                     string `xml:"execID,attr"`
	BrokerExecutionCharge      string `xml:"brokerExecutionCharge,attr"`
	BrokerClearingCharge       string `xml:"brokerClearingCharge,attr"`
	ThirdPartyExecutionCharge  string `xml:"thirdPartyExecutionCharge,attr"`
	RegFINRATradingActivityFee string `xml:"regFINRATradingActivityFee,attr"`
	RegSection31TransactionFee string `xml:"regSection31TransactionFee,attr"`
	Currency                   string `xml:"currency,attr"`
	TradeDate                  string `xml:"tradeDate,attr"`
}
