package brokerage

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr"
)

// Client provides capabilities requiring an elevated brokerage session.
type Client struct {
	client *gbkr.Client
}

// NewSession elevates to a brokerage session via SSO/DH
// (POST /iserver/auth/ssodh/init) and returns a [*Client].
func NewSession(ctx context.Context, client *gbkr.Client, req *SSOInitRequest) (*Client, error) {
	start := time.Now()
	t := client.Transport()
	var result gbkr.SessionStatus
	err := t.Post(ctx, "/iserver/auth/ssodh/init", req, &result)
	t.EmitOp(ctx, string(gbkr.OpBrokerageSession), err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return &Client{client: client}, nil
}

// SSOInitRequest is the request body for POST /iserver/auth/ssodh/init.
type SSOInitRequest struct {
	// Compete determines if other brokerage sessions should be disconnected.
	Compete *bool `json:"compete,omitempty"`

	// Publish publishes the brokerage session token when the session is initialized.
	Publish *bool `json:"publish,omitempty"`
}

func (c *Client) emitOp(ctx context.Context, op gbkr.Operation, err error, dur time.Duration, attrs ...slog.Attr) {
	c.client.Transport().EmitOp(ctx, string(op), err, dur, attrs...)
}

func (c *Client) doGet(ctx context.Context, path string, query url.Values, result any) error {
	return c.client.Transport().Get(ctx, path, query, result)
}
