package gbkr

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr/models"
)

// Accounts provides discovery of IBKR accounts without scoping to a specific one.
// IBKR path prefix: /iserver/accounts, /iserver/account/pnl/*
type Accounts struct {
	c *Client
}

// Accounts returns an [*Accounts] handle for querying accounts.
func (bc *BrokerageClient) Accounts() *Accounts {
	return &Accounts{c: bc.Client}
}

// List returns all tradable accounts
// (GET /iserver/accounts).
func (a *Accounts) List(ctx context.Context) (*models.AccountList, error) {
	start := time.Now()
	var result models.AccountList
	err := a.c.doGet(ctx, "/iserver/accounts", nil, &result)
	a.c.emitOp(ctx, OpListAccounts, err, time.Since(start), slog.Int("count", len(result.Accounts)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// PnL returns partitioned P&L data across accounts
// (GET /iserver/account/pnl/partitioned).
func (a *Accounts) PnL(ctx context.Context) (*models.PnLPartitioned, error) {
	start := time.Now()
	var result models.PnLPartitioned
	err := a.c.doGet(ctx, "/iserver/account/pnl/partitioned", nil, &result)
	a.c.emitOp(ctx, OpAccountPnL, err, time.Since(start), slog.Int("count", len(result.AcctPnL)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Account provides read access to a specific IBKR account.
// IBKR path prefix: /iserver/account/{accountId}/*
type Account struct {
	c         *Client
	accountID models.AccountID
}

// Account returns an [*Account] handle scoped to the given account ID.
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
	start := time.Now()
	var result models.AccountSummary
	path := fmt.Sprintf("/iserver/account/%s/summary", url.PathEscape(string(a.accountID)))
	err := a.c.doGet(ctx, path, nil, &result)
	a.c.emitOp(ctx, OpAccountSummary, err, time.Since(start),
		slog.String("account_id", string(a.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
