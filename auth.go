package gbkr

import (
	"context"

	"github.com/trippwill/gbkr/models"
)

// SessionClient provides brokerage session management.
// Gateway authentication is a prerequisite handled outside this library;
// this interface manages the brokerage session elevation required for
// iserver endpoints.
type SessionClient interface {
	// InitBrokerageSession elevates to a brokerage session via SSO/DH
	// (POST /iserver/auth/ssodh/init). Required before using iserver endpoints.
	InitBrokerageSession(ctx context.Context, req *models.SSOInitRequest) (*models.SessionStatus, error)

	// SessionStatus checks the current brokerage session status (POST /iserver/auth/status).
	SessionStatus(ctx context.Context) (*models.SessionStatus, error)
}

// requiredSessionPermissions lists the permissions needed for SessionClient.
var requiredSessionPermissions = []Permission{
	{AreaAuth, ResourceSession, ActionRead},
	{AreaAuth, ResourceSession, ActionWrite},
}

// Session returns a [SessionClient] if the client has the required permissions.
func Session(c *Client) (SessionClient, error) {
	if err := checkPermissions(c, requiredSessionPermissions...); err != nil {
		return nil, err
	}
	return &sessionClient{c: c}, nil
}

type sessionClient struct {
	c *Client
}

func (s *sessionClient) InitBrokerageSession(ctx context.Context, req *models.SSOInitRequest) (*models.SessionStatus, error) {
	var result models.SessionStatus
	if err := s.c.doPost(ctx, "/iserver/auth/ssodh/init", req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (s *sessionClient) SessionStatus(ctx context.Context) (*models.SessionStatus, error) {
	var result models.SessionStatus
	if err := s.c.doPost(ctx, "/iserver/auth/status", nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
