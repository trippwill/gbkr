package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// PortfolioReader provides read access to portfolio positions for a specific account.
// IBKR path prefix: /portfolio/{accountId}/*
//
// Gateway access — no permission check required.
type PortfolioReader interface {
	// AccountID returns the account this reader is scoped to.
	AccountID() models.AccountID

	// Positions returns positions for a page
	// (GET /portfolio/{accountId}/positions/{pageId}).
	Positions(ctx context.Context, page int) ([]models.Position, error)

	// Position returns a single position
	// (GET /portfolio/{accountId}/position/{conid}).
	Position(ctx context.Context, conID models.ConID) (*models.Position, error)

	// Summary returns a portfolio summary
	// (GET /portfolio/{accountId}/summary).
	Summary(ctx context.Context) (*models.PortfolioSummary, error)

	// Ledger returns the account ledger
	// (GET /portfolio/{accountId}/ledger).
	Ledger(ctx context.Context) (*models.Ledger, error)
}

// Portfolio returns a [PortfolioReader] scoped to the given account ID.
// Gateway access — no permission check required.
func (c *Client) Portfolio(accountID models.AccountID) PortfolioReader {
	return &portfolioReader{c: c, accountID: accountID}
}

type portfolioReader struct {
	c         *Client
	accountID models.AccountID
}

func (p *portfolioReader) AccountID() models.AccountID {
	return p.accountID
}

func (p *portfolioReader) Positions(ctx context.Context, page int) ([]models.Position, error) {
	var result []models.Position
	path := fmt.Sprintf("/portfolio/%s/positions/%d", url.PathEscape(string(p.accountID)), page)
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *portfolioReader) Position(ctx context.Context, conID models.ConID) (*models.Position, error) {
	var result models.Position
	path := fmt.Sprintf("/portfolio/%s/position/%d", url.PathEscape(string(p.accountID)), int(conID))
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *portfolioReader) Summary(ctx context.Context) (*models.PortfolioSummary, error) {
	var result models.PortfolioSummary
	path := fmt.Sprintf("/portfolio/%s/summary", url.PathEscape(string(p.accountID)))
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *portfolioReader) Ledger(ctx context.Context) (*models.Ledger, error) {
	var result models.Ledger
	path := fmt.Sprintf("/portfolio/%s/ledger", url.PathEscape(string(p.accountID)))
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
