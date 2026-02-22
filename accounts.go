package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// AccountLister provides discovery of IBKR accounts without scoping to a specific one.
type AccountLister interface {
	// ListAccounts returns all tradable accounts (GET /iserver/accounts).
	ListAccounts(ctx context.Context) (*models.AccountList, error)

	// AccountPnL returns partitioned P&L data across accounts (GET /iserver/account/pnl/partitioned).
	AccountPnL(ctx context.Context) (*models.PnLPartitioned, error)
}

// requiredAccountListerPermissions lists the permissions needed for AccountLister.
var requiredAccountListerPermissions = []Permission{
	{AreaTrading, ResourceAccounts, ActionRead},
}

// Accounts returns an [AccountLister] if the client has the required permissions.
// Requires: trading.accounts.read.
func (bc *BrokerageClient) Accounts() (AccountLister, error) {
	if err := checkPermissions(bc.Client, requiredAccountListerPermissions...); err != nil {
		return nil, err
	}
	return &accountLister{c: bc.Client}, nil
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
type AccountReader interface {
	// AccountID returns the account this reader is scoped to.
	AccountID() models.AccountID

	// Summary returns the account summary (GET /iserver/account/{accountId}/summary).
	Summary(ctx context.Context) (*models.AccountSummary, error)

	// Positions returns a PositionReader for this account if the client
	// has the required portfolio position permissions.
	Positions() (PositionReader, error)
}

// requiredAccountReaderPermissions lists the permissions needed for AccountReader.
var requiredAccountReaderPermissions = []Permission{
	{AreaTrading, ResourceAccounts, ActionRead},
}

// Account returns an [AccountReader] scoped to the given account ID.
// Requires: trading.accounts.read.
func (bc *BrokerageClient) Account(accountID models.AccountID) (AccountReader, error) {
	if err := checkPermissions(bc.Client, requiredAccountReaderPermissions...); err != nil {
		return nil, err
	}
	return &accountReader{c: bc.Client, accountID: accountID}, nil
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

func (a *accountReader) Positions() (PositionReader, error) {
	return newPositionReader(a.c, a.accountID)
}
