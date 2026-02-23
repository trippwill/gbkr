package gbkr

import (
	"context"

	"github.com/trippwill/gbkr/models"
)

// SessionStatus checks the current brokerage session status
// (POST /iserver/auth/status).
// Gateway access — no permission check required.
func (c *Client) SessionStatus(ctx context.Context) (*models.SessionStatus, error) {
	var result models.SessionStatus
	if err := c.doPost(ctx, "/iserver/auth/status", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// BrokerageSession elevates to a brokerage session via SSO/DH
// (POST /iserver/auth/ssodh/init) and returns a [*BrokerageClient].
// Each call performs a fresh handshake (no caching).
// Requires: brokerage:read.
func (c *Client) BrokerageSession(ctx context.Context, req *models.SSOInitRequest) (*BrokerageClient, error) {
	if err := checkPermissions(c, ScopeBrokerage, LevelRead); err != nil {
		return nil, err
	}
	var result models.SessionStatus
	if err := c.doPost(ctx, "/iserver/auth/ssodh/init", req, &result); err != nil {
		return nil, err
	}
	return &BrokerageClient{Client: c}, nil
}
