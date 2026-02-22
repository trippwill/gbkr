package gbkr

import (
	"context"
	"fmt"

	"github.com/trippwill/gbkr/models"
)

// ContractReader provides read access to contract details.
// IBKR path prefix: /iserver/contract/{conid}/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type ContractReader interface {
	// Info returns full contract details
	// (GET /iserver/contract/{conid}/info).
	Info(ctx context.Context, conID models.ConID) (*models.ContractDetails, error)
}

// Contracts returns a [ContractReader] for querying contract information.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) Contracts() ContractReader {
	return &contractReader{c: bc.Client}
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
