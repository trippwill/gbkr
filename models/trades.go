package models

import "encoding/json"

// TradeExecution represents a single trade execution returned by
// GET /iserver/account/trades.
type TradeExecution struct {
	// ExecutionID is the unique execution identifier.
	ExecutionID string
	// Symbol is the trading symbol.
	Symbol string
	// Side is the order side (e.g., "B" for buy, "S" for sell).
	Side string
	// OrderDescription is a human-readable order summary.
	OrderDescription string
	// TradeTime is the UTC trade time (format: "20231211-18:00:49").
	TradeTime string
	// TradeTimeEpoch is the epoch millisecond timestamp of the trade.
	TradeTimeEpoch int64
	// Size is the quantity traded.
	Size float64
	// Price is the execution price.
	Price string
	// OrderRef is the user-defined order reference.
	OrderRef string
	// Exchange is the execution exchange.
	Exchange Exchange
	// Commission is the trade commission.
	Commission string
	// NetAmount is the total net cost of the trade.
	NetAmount float64
	// Account is the account identifier.
	Account AccountID
	// CompanyName is the long company name.
	CompanyName string
	// ContractDesc is the local symbol / contract description.
	ContractDesc string
	// SecType is the security type (e.g., "STK", "OPT").
	SecType AssetClass
	// ListingExchange is the primary listing exchange.
	ListingExchange Exchange
	// ConID is the contract identifier.
	ConID ConID
}

func (t *TradeExecution) UnmarshalJSON(data []byte) error {
	var raw struct {
		ExecutionID      *string  `json:"execution_id,omitempty"`
		Symbol           *string  `json:"symbol,omitempty"`
		Side             *string  `json:"side,omitempty"`
		OrderDescription *string  `json:"order_description,omitempty"`
		TradeTime        *string  `json:"trade_time,omitempty"`
		TradeTimeEpoch   *int64   `json:"trade_time_r,omitempty"`
		Size             *float64 `json:"size,omitempty"`
		Price            *string  `json:"price,omitempty"`
		OrderRef         *string  `json:"order_ref,omitempty"`
		Exchange         *string  `json:"exchange,omitempty"`
		Commission       *string  `json:"commission,omitempty"`
		NetAmount        *float64 `json:"net_amount,omitempty"`
		Account          *string  `json:"account,omitempty"`
		CompanyName      *string  `json:"company_name,omitempty"`
		ContractDesc     *string  `json:"contract_description_1,omitempty"`
		SecType          *string  `json:"sec_type,omitempty"`
		ListingExchange  *string  `json:"listing_exchange,omitempty"`
		ConID            *int     `json:"conid,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t.ExecutionID = deref(raw.ExecutionID)
	t.Symbol = deref(raw.Symbol)
	t.Side = deref(raw.Side)
	t.OrderDescription = deref(raw.OrderDescription)
	t.TradeTime = deref(raw.TradeTime)
	t.TradeTimeEpoch = deref(raw.TradeTimeEpoch)
	t.Size = deref(raw.Size)
	t.Price = deref(raw.Price)
	t.OrderRef = deref(raw.OrderRef)
	t.Exchange = Exchange(deref(raw.Exchange))
	t.Commission = deref(raw.Commission)
	t.NetAmount = deref(raw.NetAmount)
	t.Account = AccountID(deref(raw.Account))
	t.CompanyName = deref(raw.CompanyName)
	t.ContractDesc = deref(raw.ContractDesc)
	t.SecType = AssetClass(deref(raw.SecType))
	t.ListingExchange = Exchange(deref(raw.ListingExchange))
	t.ConID = ConID(deref(raw.ConID))
	return nil
}

// TransactionHistoryRequest is the request body for POST /pa/transactions.
type TransactionHistoryRequest struct {
	AcctIDs  []string `json:"acctIds"`
	ConIDs   []int    `json:"conids"`
	Currency string   `json:"currency"`
	Days     string   `json:"days,omitempty"`
}

// TransactionHistoryResponse is the response from POST /pa/transactions.
type TransactionHistoryResponse struct {
	// Currency is the reporting currency.
	Currency Currency
	// From is the epoch start time of the range.
	From int64
	// To is the epoch end time of the range.
	To int64
	// IncludesRealTime indicates if trades are up to date.
	IncludesRealTime bool
	// Transactions is the list of transactions.
	Transactions []Transaction
}

func (r *TransactionHistoryResponse) UnmarshalJSON(data []byte) error {
	var raw struct {
		Currency         *string          `json:"currency,omitempty"`
		From             *int64           `json:"from,omitempty"`
		To               *int64           `json:"to,omitempty"`
		IncludesRealTime *bool            `json:"includesRealTime,omitempty"`
		Transactions     *json.RawMessage `json:"transactions,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Currency = Currency(deref(raw.Currency))
	r.From = deref(raw.From)
	r.To = deref(raw.To)
	r.IncludesRealTime = deref(raw.IncludesRealTime)
	if raw.Transactions != nil {
		if err := json.Unmarshal(*raw.Transactions, &r.Transactions); err != nil {
			return err
		}
	}
	return nil
}

// Transaction represents a single transaction from POST /pa/transactions.
type Transaction struct {
	// Date is the human-readable datetime of the transaction.
	Date string
	// Currency is the traded instrument currency.
	Currency Currency
	// FxRate is the forex conversion rate.
	FxRate float64
	// Price is the price per share.
	Price float64
	// Qty is the quantity traded (negative for sells).
	Qty int
	// Account is the account identifier.
	Account AccountID
	// Amount is the total trade value.
	Amount float64
	// ConID is the contract identifier.
	ConID ConID
	// Type is the order side (e.g., "Sell", "Buy").
	Type string
	// Desc is the long company name.
	Desc string
}

func (t *Transaction) UnmarshalJSON(data []byte) error {
	var raw struct {
		Date     *string  `json:"date,omitempty"`
		Currency *string  `json:"cur,omitempty"`
		FxRate   *float64 `json:"fxRate,omitempty"`
		Price    *float64 `json:"pr,omitempty"`
		Qty      *int     `json:"qty,omitempty"`
		Account  *string  `json:"acctid,omitempty"`
		Amount   *float64 `json:"amt,omitempty"`
		ConID    *int     `json:"conid,omitempty"`
		Type     *string  `json:"type,omitempty"`
		Desc     *string  `json:"desc,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	t.Date = deref(raw.Date)
	t.Currency = Currency(deref(raw.Currency))
	t.FxRate = deref(raw.FxRate)
	t.Price = deref(raw.Price)
	t.Qty = deref(raw.Qty)
	t.Account = AccountID(deref(raw.Account))
	t.Amount = deref(raw.Amount)
	t.ConID = ConID(deref(raw.ConID))
	t.Type = deref(raw.Type)
	t.Desc = deref(raw.Desc)
	return nil
}
