package brokerage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/internal/jx"
)

// Accounts provides discovery of IBKR accounts without scoping to a specific one.
// IBKR path prefix: /iserver/accounts, /iserver/account/pnl/*
type Accounts struct {
	c *Client
}

// Accounts returns an [*Accounts] handle for querying accounts.
func (c *Client) Accounts() *Accounts {
	return &Accounts{c: c}
}

// List returns all tradable accounts
// (GET /iserver/accounts).
func (a *Accounts) List(ctx context.Context) (*AccountList, error) {
	start := time.Now()
	var result AccountList
	err := a.c.doGet(ctx, "/iserver/accounts", nil, &result)
	a.c.emitOp(ctx, gbkr.OpListAccounts, err, time.Since(start), slog.Int("count", len(result.Accounts)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PnL returns partitioned P&L data across accounts
// (GET /iserver/account/pnl/partitioned).
func (a *Accounts) PnL(ctx context.Context) (*PnLPartitioned, error) {
	start := time.Now()
	var result PnLPartitioned
	err := a.c.doGet(ctx, "/iserver/account/pnl/partitioned", nil, &result)
	a.c.emitOp(ctx, gbkr.OpAccountPnL, err, time.Since(start), slog.Int("count", len(result.AcctPnL)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Account provides read access to a specific IBKR account.
// IBKR path prefix: /iserver/account/{accountId}/*
type Account struct {
	c         *Client
	accountID gbkr.AccountID
}

// Account returns an [*Account] handle scoped to the given account ID.
func (c *Client) Account(accountID gbkr.AccountID) *Account {
	return &Account{c: c, accountID: accountID}
}

// ID returns the account this handle is scoped to.
func (a *Account) ID() gbkr.AccountID {
	return a.accountID
}

// Summary returns the account summary
// (GET /iserver/account/{accountId}/summary).
func (a *Account) Summary(ctx context.Context) (*AccountSummary, error) {
	start := time.Now()
	var result AccountSummary
	path := fmt.Sprintf("/iserver/account/%s/summary", url.PathEscape(string(a.accountID)))
	err := a.c.doGet(ctx, path, nil, &result)
	a.c.emitOp(ctx, gbkr.OpAccountSummary, err, time.Since(start),
		slog.String("account_id", string(a.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// AccountList is the response for GET /iserver/accounts.
type AccountList struct {
	Accounts      []gbkr.AccountID // Array of all accessible account IDs.
	SelectedAcct  gbkr.AccountID   // The currently selected account. (API: "selectedAccount")
	AllowFeatures AllowFeatures    // Feature flags for the account.
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
		a.Accounts = make([]gbkr.AccountID, len(*raw.Accounts))
		for i, id := range *raw.Accounts {
			a.Accounts[i] = gbkr.AccountID(id)
		}
	}
	a.SelectedAcct = gbkr.AccountID(jx.Deref(raw.SelectedAcct))
	if raw.AllowFeatures != nil {
		a.AllowFeatures = *raw.AllowFeatures
	}
	return nil
}

// AllowFeatures describes feature flags for the account.
type AllowFeatures struct {
	AllowCrypto       bool   // Whether crypto currencies can be traded.
	AllowFXConv       bool   // Whether currency conversion is allowed.
	AllowEventTrading bool   // Whether Event Trader can be used.
	AllowTypeAhead    bool   // Whether Type-Ahead support is available.
	AllowedAssetTypes string // List of asset types the account can trade.
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
	f.AllowCrypto = jx.Deref(raw.AllowCrypto)
	f.AllowFXConv = jx.Deref(raw.AllowFXConv)
	f.AllowEventTrading = jx.Deref(raw.AllowEventTrading)
	f.AllowTypeAhead = jx.Deref(raw.AllowTypeAhead)
	f.AllowedAssetTypes = jx.Deref(raw.AllowedAssetTypes)
	return nil
}

// AccountSummary is the response for GET /iserver/account/{accountId}/summary.
type AccountSummary struct {
	AccountReady bool                      // Indicates if the account is ready. (API: "accountready")
	AccountType  gbkr.AccountType          // Account type classification. (API: "accounttype")
	AccountID    gbkr.AccountID            // Account identifier. (API: "accountId")
	Currency     gbkr.Currency             // Account base currency.
	Sections     map[string][]SummaryField `json:"-"` // Dynamic sections with summary fields.
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
	s.AccountReady = jx.Deref(raw.AccountReady)
	s.AccountType = gbkr.AccountType(jx.Deref(raw.AccountType))
	s.AccountID = gbkr.AccountID(jx.Deref(raw.AccountID))
	s.Currency = gbkr.Currency(jx.Deref(raw.Currency))
	return nil
}

// SummaryField represents a single field within an account summary section.
type SummaryField struct {
	Amount    float64       // Numeric amount.
	Currency  gbkr.Currency // Currency code.
	IsNull    bool          // Whether the value is null.
	Severity  int           // Severity level.
	Timestamp int64         // Epoch timestamp.
	Value     string        // Display value.
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
	f.Amount = jx.Deref(raw.Amount)
	f.Currency = gbkr.Currency(jx.Deref(raw.Currency))
	f.IsNull = jx.Deref(raw.IsNull)
	f.Severity = jx.Deref(raw.Severity)
	f.Timestamp = jx.Deref(raw.Timestamp)
	f.Value = jx.Deref(raw.Value)
	return nil
}

// PnLPartitioned is the response for GET /iserver/account/pnl/partitioned.
type PnLPartitioned struct {
	AcctPnL map[gbkr.AccountID]PnLEntry `json:"acctPnl,omitempty"`
}

// PnLEntry holds profit/loss data for a single account.
//
// Response for GET /iserver/account/pnl/partitioned (within the acctPnl map).
// Several fields use abbreviated API names; friendly Go names are provided.
type PnLEntry struct {
	DailyPnL        float64 // API: "dpl" — daily profit/loss.
	NetLiquidation  float64 // API: "nl" — net liquidity value.
	UnrealizedPnL   float64 // API: "upl" — unrealized profit/loss.
	RealizedPnL     float64 // API: "rpl" — realized profit/loss.
	ExcessLiquidity float64 // API: "el" — excess liquidity.
	MarginValue     float64 // API: "mv" — margin value.
}

func (e *PnLEntry) UnmarshalJSON(data []byte) error {
	var raw struct {
		DailyPnL        *float64 `json:"dpl,omitempty"`
		NetLiquidation  *float64 `json:"nl,omitempty"`
		UnrealizedPnL   *float64 `json:"upl,omitempty"`
		RealizedPnL     *float64 `json:"rpl,omitempty"`
		ExcessLiquidity *float64 `json:"el,omitempty"`
		MarginValue     *float64 `json:"mv,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.DailyPnL = jx.Deref(raw.DailyPnL)
	e.NetLiquidation = jx.Deref(raw.NetLiquidation)
	e.UnrealizedPnL = jx.Deref(raw.UnrealizedPnL)
	e.RealizedPnL = jx.Deref(raw.RealizedPnL)
	e.ExcessLiquidity = jx.Deref(raw.ExcessLiquidity)
	e.MarginValue = jx.Deref(raw.MarginValue)
	return nil
}
