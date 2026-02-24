package gbkr

import (
	"context"
	"fmt"
	"log/slog"
	"time"

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
	start := time.Now()
	var result models.ContractDetails
	path := fmt.Sprintf("/iserver/contract/%d/info", int(conID))
	err := r.c.doGet(ctx, path, nil, &result)
	r.c.emitOp(ctx, OpContractInfo, err, time.Since(start),
		slog.Int64("conid", int64(conID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
