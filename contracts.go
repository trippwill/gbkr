package gbkr

import (
	"context"
	"fmt"

	"github.com/trippwill/gbkr/models"
)

// Contracts provides read access to contract details.
// IBKR path prefix: /iserver/contract/{conid}/*
type Contracts struct {
	c *Client
}

// Contracts returns a [*Contracts] handle for querying contract information.
func (bc *BrokerageClient) Contracts() *Contracts {
	return &Contracts{c: bc.Client}
}

// Info returns full contract details
// (GET /iserver/contract/{conid}/info).
func (r *Contracts) Info(ctx context.Context, conID models.ConID) (*models.ContractDetails, error) {
	var result models.ContractDetails
	path := fmt.Sprintf("/iserver/contract/%d/info", int(conID))
	if err := r.c.doGet(ctx, path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
