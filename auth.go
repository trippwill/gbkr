package gbkr

import (
	"context"
	"time"

	"github.com/trippwill/gbkr/models"
)

// SessionStatus checks the current brokerage session status
// (POST /iserver/auth/status).
func (c *Client) SessionStatus(ctx context.Context) (*models.SessionStatus, error) {
	start := time.Now()
	var result models.SessionStatus
	err := c.doPost(ctx, "/iserver/auth/status", nil, &result)
	c.emitOp(ctx, OpSessionStatus, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// BrokerageSession elevates to a brokerage session via SSO/DH
// (POST /iserver/auth/ssodh/init) and returns a [*BrokerageClient].
// Each call performs a fresh handshake (no caching).
func (c *Client) BrokerageSession(ctx context.Context, req *models.SSOInitRequest) (*BrokerageClient, error) {
	start := time.Now()
	var result models.SessionStatus
	err := c.doPost(ctx, "/iserver/auth/ssodh/init", req, &result)
	c.emitOp(ctx, OpBrokerageSession, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &BrokerageClient{Client: c}, nil
}
