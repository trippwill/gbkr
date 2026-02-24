package gbkr

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr/models"
)

// Portfolio provides read access to portfolio positions for a specific account.
// IBKR path prefix: /portfolio/{accountId}/*
type Portfolio struct {
	c         *Client
	accountID models.AccountID
}

// Portfolio returns a [*Portfolio] handle scoped to the given account ID.
func (c *Client) Portfolio(accountID models.AccountID) *Portfolio {
	return &Portfolio{c: c, accountID: accountID}
}

// ID returns the account this handle is scoped to.
func (p *Portfolio) ID() models.AccountID {
	return p.accountID
}

// Positions returns positions for a page
// (GET /portfolio/{accountId}/positions/{pageId}).
func (p *Portfolio) Positions(ctx context.Context, page int) ([]models.Position, error) {
	start := time.Now()
	var result []models.Position
	path := fmt.Sprintf("/portfolio/%s/positions/%d", url.PathEscape(string(p.accountID)), page)
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioPositions, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// Position returns a single position
// (GET /portfolio/{accountId}/position/{conid}).
func (p *Portfolio) Position(ctx context.Context, conID models.ConID) (*models.Position, error) {
	start := time.Now()
	var result models.Position
	path := fmt.Sprintf("/portfolio/%s/position/%d", url.PathEscape(string(p.accountID)), int(conID))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioPosition, err, time.Since(start),
		slog.String("account_id", string(p.accountID)),
		slog.Int64("conid", int64(conID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Summary returns a portfolio summary
// (GET /portfolio/{accountId}/summary).
func (p *Portfolio) Summary(ctx context.Context) (*models.PortfolioSummary, error) {
	start := time.Now()
	var result models.PortfolioSummary
	path := fmt.Sprintf("/portfolio/%s/summary", url.PathEscape(string(p.accountID)))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioSummary, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// Ledger returns the account ledger
// (GET /portfolio/{accountId}/ledger).
func (p *Portfolio) Ledger(ctx context.Context) (*models.Ledger, error) {
	start := time.Now()
	var result models.Ledger
	path := fmt.Sprintf("/portfolio/%s/ledger", url.PathEscape(string(p.accountID)))
	err := p.c.doGet(ctx, path, nil, &result)
	p.c.emitOp(ctx, OpPortfolioLedger, err, time.Since(start),
		slog.String("account_id", string(p.accountID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
