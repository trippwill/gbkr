package brokerage

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr"
)

// SecurityDefinitions provides search access to security definitions.
// IBKR path prefix: /iserver/secdef/*
type SecurityDefinitions struct {
	c *Client
}

// SecurityDefinitions returns a [*SecurityDefinitions] handle for searching security definitions.
func (c *Client) SecurityDefinitions() *SecurityDefinitions {
	return &SecurityDefinitions{c: c}
}

// Search finds contracts matching the query string
// (GET /iserver/secdef/search).
func (r *SecurityDefinitions) Search(ctx context.Context, symbol string) ([]ContractSearchResult, error) {
	start := time.Now()
	query := url.Values{}
	query.Set("symbol", symbol)
	var result []ContractSearchResult
	err := r.c.doGet(ctx, "/iserver/secdef/search", query, &result)
	r.c.emitOp(ctx, gbkr.OpSecuritySearch, err, time.Since(start),
		slog.String("symbol", symbol))
	if err != nil {
		return nil, err
	}
	return result, nil
}
