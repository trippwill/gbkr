package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// Accounts provides discovery of IBKR accounts without scoping to a specific one.
// IBKR path prefix: /iserver/accounts, /iserver/account/pnl/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type Accounts struct {
	c *Client
}

// Accounts returns an [*Accounts] handle for querying accounts.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) Accounts() *Accounts {
	return &Accounts{c: bc.Client}
}

// List returns all tradable accounts
// (GET /iserver/accounts).
func (a *Accounts) List(ctx context.Context) (*models.AccountList, error) {
	var result models.AccountList
	if err := a.c.doGet(ctx, "/iserver/accounts", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PnL returns partitioned P&L data across accounts
// (GET /iserver/account/pnl/partitioned).
func (a *Accounts) PnL(ctx context.Context) (*models.PnLPartitioned, error) {
	var result models.PnLPartitioned
	if err := a.c.doGet(ctx, "/iserver/account/pnl/partitioned", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Account provides read access to a specific IBKR account.
// IBKR path prefix: /iserver/account/{accountId}/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type Account struct {
	c         *Client
	accountID models.AccountID
}

// Account returns an [*Account] handle scoped to the given account ID.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) Account(accountID models.AccountID) *Account {
	return &Account{c: bc.Client, accountID: accountID}
}

// ID returns the account this handle is scoped to.
func (a *Account) ID() models.AccountID {
	return a.accountID
}

// Summary returns the account summary
// (GET /iserver/account/{accountId}/summary).
func (a *Account) Summary(ctx context.Context) (*models.AccountSummary, error) {
	var result models.AccountSummary
	path := fmt.Sprintf("/iserver/account/%s/summary", url.PathEscape(string(a.accountID)))
	if err := a.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
