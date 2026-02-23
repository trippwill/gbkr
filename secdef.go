package gbkr

import (
	"context"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// SecurityDefinitionReader provides search access to security definitions.
// IBKR path prefix: /iserver/secdef/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type SecurityDefinitionReader interface {
	// Search finds contracts matching the query string
	// (GET /iserver/secdef/search).
	Search(ctx context.Context, symbol string) ([]models.ContractSearchResult, error)
}

// SecurityDefinitions returns a [SecurityDefinitionReader] for searching security definitions.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) SecurityDefinitions() SecurityDefinitionReader {
	return &securityDefinitionReader{c: bc.Client}
}

type securityDefinitionReader struct {
	c *Client
}

func (r *securityDefinitionReader) Search(ctx context.Context, symbol string) ([]models.ContractSearchResult, error) {
	query := url.Values{}
	query.Set("symbol", symbol)
	var result []models.ContractSearchResult
	if err := r.c.doGet(ctx, "/iserver/secdef/search", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
