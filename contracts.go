package gbkr

import (
	"context"
	"fmt"
	"net/url"

	"github.com/trippwill/gbkr/models"
)

// ContractReader provides read access to contract details and search.
type ContractReader interface {
	// Info returns full contract details (GET /iserver/contract/{conid}/info).
	Info(ctx context.Context, conID models.ConID) (*models.ContractDetails, error)

	// Search finds contracts matching the query string (GET /iserver/secdef/search).
	Search(ctx context.Context, symbol string) ([]models.ContractSearchResult, error)
}

var requiredContractPermissions = []Permission{
	{AreaTrading, ResourceContracts, ActionRead},
}

// Contracts returns a [ContractReader] for querying contract information.
// Requires: trading.contracts.read.
func (bc *BrokerageClient) Contracts() (ContractReader, error) {
	if err := checkPermissions(bc.Client, requiredContractPermissions...); err != nil {
		return nil, err
	}
	return &contractReader{c: bc.Client}, nil
}

type contractReader struct {
	c *Client
}

func (r *contractReader) Info(ctx context.Context, conID models.ConID) (*models.ContractDetails, error) {
	var result models.ContractDetails
	path := fmt.Sprintf("/iserver/contract/%d/info", int(conID))
	if err := r.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *contractReader) Search(ctx context.Context, symbol string) ([]models.ContractSearchResult, error) {
	query := url.Values{}
	query.Set("symbol", symbol)
	var result []models.ContractSearchResult
	if err := r.c.doGet(ctx, "/iserver/secdef/search", query, &result); err != nil {
		return nil, err
	}
	return result, nil
}
