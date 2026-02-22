package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// AccountLister provides discovery of IBKR accounts without scoping to a specific one.
// IBKR path prefix: /iserver/accounts, /iserver/account/pnl/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type AccountLister interface {
	// ListAccounts returns all tradable accounts
	// (GET /iserver/accounts).
	ListAccounts(ctx context.Context) (*models.AccountList, error)

	// AccountPnL returns partitioned P&L data across accounts
	// (GET /iserver/account/pnl/partitioned).
	AccountPnL(ctx context.Context) (*models.PnLPartitioned, error)
}

// Accounts returns an [AccountLister] if the client has the required permissions.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) Accounts() AccountLister {
	return &accountLister{c: bc.Client}
}

type accountLister struct {
	c *Client
}

func (a *accountLister) ListAccounts(ctx context.Context) (*models.AccountList, error) {
	var result models.AccountList
	if err := a.c.doGet(ctx, "/iserver/accounts", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (a *accountLister) AccountPnL(ctx context.Context) (*models.PnLPartitioned, error) {
	var result models.PnLPartitioned
	if err := a.c.doGet(ctx, "/iserver/account/pnl/partitioned", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// AccountReader provides read access to a specific IBKR account.
// IBKR path prefix: /iserver/account/{accountId}/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type AccountReader interface {
	// AccountID returns the account this reader is scoped to.
	AccountID() models.AccountID

	// Summary returns the account summary
	// (GET /iserver/account/{accountId}/summary).
	Summary(ctx context.Context) (*models.AccountSummary, error)
}

// Account returns an [AccountReader] scoped to the given account ID.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) Account(accountID models.AccountID) AccountReader {
	return &accountReader{c: bc.Client, accountID: accountID}
}

type accountReader struct {
	c         *Client
	accountID models.AccountID
}

func (a *accountReader) AccountID() models.AccountID {
	return a.accountID
}

func (a *accountReader) Summary(ctx context.Context) (*models.AccountSummary, error) {
	var result models.AccountSummary
	path := fmt.Sprintf("/iserver/account/%s/summary", url.PathEscape(string(a.accountID)))
	if err := a.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
