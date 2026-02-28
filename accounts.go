package gbkr

import (
	"context"
	"encoding/json"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
)

// PortfolioAccount describes an account returned by the portfolio accounts endpoint.
type PortfolioAccount struct {
	AccountID AccountID
	Title     string
	Desc      string
	Currency  Currency
	Type      AccountType
	IsProp    bool
	IsMulti   bool
}

// UnmarshalJSON implements [json.Unmarshaler].
func (a *PortfolioAccount) UnmarshalJSON(data []byte) error {
	var raw struct {
		AccountID *string `json:"accountId"`
		Title     *string `json:"accountTitle"`
		Desc      *string `json:"desc"`
		Currency  *string `json:"currency"`
		Type      *string `json:"type"`
		IsProp    *bool   `json:"isProp"`
		IsMulti   *bool   `json:"isMulti"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	a.AccountID = AccountID(jx.Deref(raw.AccountID))
	a.Title = jx.Deref(raw.Title)
	a.Desc = jx.Deref(raw.Desc)
	a.Currency = Currency(jx.Deref(raw.Currency))
	a.Type = AccountType(jx.Deref(raw.Type))
	a.IsProp = jx.Deref(raw.IsProp)
	a.IsMulti = jx.Deref(raw.IsMulti)
	return nil
}

// PortfolioAccounts provides gateway-level account discovery.
// IBKR path: /portfolio/accounts
//
// This is distinct from [brokerage.Accounts] which uses /iserver/accounts
// and requires an elevated brokerage session. PortfolioAccounts works with
// the basic gateway session and should be called before other /portfolio/*
// endpoints to initialize account context.
type PortfolioAccounts struct {
	c *Client
}

// PortfolioAccounts returns a [*PortfolioAccounts] handle for gateway-level account discovery.
func (c *Client) PortfolioAccounts() *PortfolioAccounts {
	return &PortfolioAccounts{c: c}
}

// List returns accounts available at the gateway (portfolio) level
// (GET /portfolio/accounts).
func (a *PortfolioAccounts) List(ctx context.Context) ([]PortfolioAccount, error) {
	start := time.Now()
	var result []PortfolioAccount
	err := a.c.doGet(ctx, "/portfolio/accounts", nil, &result)
	a.c.emitOp(ctx, OpPortfolioAccounts, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return result, nil
}
