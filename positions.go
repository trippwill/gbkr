package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// PositionReader provides read access to portfolio positions for a specific account.
type PositionReader interface {
	// AccountID returns the account this reader is scoped to.
	AccountID() models.AccountID

	// ListPositions returns positions for a page (GET /portfolio/{accountId}/positions/{pageId}).
	ListPositions(ctx context.Context, page int) ([]models.Position, error)

	// Position returns a single position (GET /portfolio/{accountid}/position/{conid}).
	Position(ctx context.Context, conID models.ConID) (*models.Position, error)

	// PortfolioSummary returns a portfolio summary (GET /portfolio/{accountId}/summary).
	PortfolioSummary(ctx context.Context) (*models.PortfolioSummary, error)

	// Ledger returns the account ledger (GET /portfolio/{accountId}/ledger).
	Ledger(ctx context.Context) (*models.Ledger, error)
}

// requiredPositionPermissions lists the permissions needed for PositionReader.
var requiredPositionPermissions = []Permission{
	{AreaPortfolio, ResourcePositions, ActionRead},
}

// newPositionReader creates a PositionReader with permission checking.
// Used by both the standalone constructor and AccountReader.Positions().
func newPositionReader(c *Client, accountID models.AccountID) (PositionReader, error) {
	if err := checkPermissions(c, requiredPositionPermissions...); err != nil {
		return nil, err
	}
	return &positionReader{c: c, accountID: accountID}, nil
}

// Positions returns a [PositionReader] scoped to the given account ID.
// Requires: portfolio.positions.read.
func Positions(c *Client, accountID models.AccountID) (PositionReader, error) {
	return newPositionReader(c, accountID)
}

type positionReader struct {
	c         *Client
	accountID models.AccountID
}

func (p *positionReader) AccountID() models.AccountID {
	return p.accountID
}

func (p *positionReader) ListPositions(ctx context.Context, page int) ([]models.Position, error) {
	var result []models.Position
	path := fmt.Sprintf("/portfolio/%s/positions/%d", url.PathEscape(string(p.accountID)), page)
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (p *positionReader) Position(ctx context.Context, conID models.ConID) (*models.Position, error) {
	var result models.Position
	path := fmt.Sprintf("/portfolio/%s/position/%d", url.PathEscape(string(p.accountID)), int(conID))
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *positionReader) PortfolioSummary(ctx context.Context) (*models.PortfolioSummary, error) {
	var result models.PortfolioSummary
	path := fmt.Sprintf("/portfolio/%s/summary", url.PathEscape(string(p.accountID)))
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (p *positionReader) Ledger(ctx context.Context) (*models.Ledger, error) {
	var result models.Ledger
	path := fmt.Sprintf("/portfolio/%s/ledger", url.PathEscape(string(p.accountID)))
	if err := p.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
