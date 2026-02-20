package models

import "encoding/json"

// AccountList is the response for GET /iserver/accounts.
type AccountList struct {
	Accounts      []AccountID
	SelectedAcct  AccountID
	AllowFeatures AllowFeatures
}

func (a *AccountList) UnmarshalJSON(data []byte) error {
	var raw struct {
		Accounts      *[]string      `json:"accounts,omitempty"`
		SelectedAcct  *string        `json:"selectedAccount,omitempty"`
		AllowFeatures *AllowFeatures `json:"allowFeatures,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw.Accounts != nil {
		a.Accounts = make([]AccountID, len(*raw.Accounts))
		for i, id := range *raw.Accounts {
			a.Accounts[i] = AccountID(id)
		}
	}
	a.SelectedAcct = AccountID(deref(raw.SelectedAcct))
	if raw.AllowFeatures != nil {
		a.AllowFeatures = *raw.AllowFeatures
	}
	return nil
}

// AllowFeatures describes feature flags for the account.
type AllowFeatures struct {
	AllowCrypto       bool
	AllowFXConv       bool
	AllowEventTrading bool
	AllowTypeAhead    bool
	AllowedAssetTypes string
}

func (f *AllowFeatures) UnmarshalJSON(data []byte) error {
	var raw struct {
		AllowCrypto       *bool   `json:"allowCrypto,omitempty"`
		AllowFXConv       *bool   `json:"allowFXConv,omitempty"`
		AllowEventTrading *bool   `json:"allowEventTrading,omitempty"`
		AllowTypeAhead    *bool   `json:"allowTypeAhead,omitempty"`
		AllowedAssetTypes *string `json:"allowedAssetTypes,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.AllowCrypto = deref(raw.AllowCrypto)
	f.AllowFXConv = deref(raw.AllowFXConv)
	f.AllowEventTrading = deref(raw.AllowEventTrading)
	f.AllowTypeAhead = deref(raw.AllowTypeAhead)
	f.AllowedAssetTypes = deref(raw.AllowedAssetTypes)
	return nil
}

// AccountSummary is the response for GET /iserver/account/{accountId}/summary.
type AccountSummary struct {
	AccountReady bool
	AccountType  AccountType
	AccountID    AccountID
	Currency     Currency
	Sections     map[string][]SummaryField `json:"-"`
}

func (s *AccountSummary) UnmarshalJSON(data []byte) error {
	var raw struct {
		AccountReady *bool   `json:"accountready,omitempty"`
		AccountType  *string `json:"accounttype,omitempty"`
		AccountID    *string `json:"accountId,omitempty"`
		Currency     *string `json:"currency,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.AccountReady = deref(raw.AccountReady)
	s.AccountType = AccountType(deref(raw.AccountType))
	s.AccountID = AccountID(deref(raw.AccountID))
	s.Currency = Currency(deref(raw.Currency))
	return nil
}

// SummaryField represents a single field within an account summary section.
type SummaryField struct {
	Amount    float64
	Currency  Currency
	IsNull    bool
	Severity  int
	Timestamp int64
	Value     string
}

func (f *SummaryField) UnmarshalJSON(data []byte) error {
	var raw struct {
		Amount    *float64 `json:"amount,omitempty"`
		Currency  *string  `json:"currency,omitempty"`
		IsNull    *bool    `json:"isNull,omitempty"`
		Severity  *int     `json:"severity,omitempty"`
		Timestamp *int64   `json:"timestamp,omitempty"`
		Value     *string  `json:"value,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	f.Amount = deref(raw.Amount)
	f.Currency = Currency(deref(raw.Currency))
	f.IsNull = deref(raw.IsNull)
	f.Severity = deref(raw.Severity)
	f.Timestamp = deref(raw.Timestamp)
	f.Value = deref(raw.Value)
	return nil
}

// PnLPartitioned is the response for GET /iserver/account/pnl/partitioned.
type PnLPartitioned struct {
	AcctPnL map[AccountID]PnLEntry `json:"acctPnl,omitempty"`
}

// PnLEntry holds profit/loss data for a single account.
type PnLEntry struct {
	DailyPnL        float64
	NetLiquidation  float64
	UnrealizedPnL   float64
	RealizedPnL     float64
	ExcessLiquidity float64
	MarketValue     float64
}

func (e *PnLEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		DailyPnL        *float64 `json:"dpl,omitempty"`
		NetLiquidation  *float64 `json:"nl,omitempty"`
		UnrealizedPnL   *float64 `json:"upl,omitempty"`
		RealizedPnL     *float64 `json:"rpl,omitempty"`
		ExcessLiquidity *float64 `json:"el,omitempty"`
		MarketValue     *float64 `json:"mv,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.DailyPnL = deref(raw.DailyPnL)
	e.NetLiquidation = deref(raw.NetLiquidation)
	e.UnrealizedPnL = deref(raw.UnrealizedPnL)
	e.RealizedPnL = deref(raw.RealizedPnL)
	e.ExcessLiquidity = deref(raw.ExcessLiquidity)
	e.MarketValue = deref(raw.MarketValue)
	return nil
}
