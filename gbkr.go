package gbkr

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/trippwill/gbkr/internal/transport"
)

// Client is the base IBKR API client. It holds transport configuration.
// Provides read-only capabilities and session elevation via [brokerage.NewSession].
type Client struct {
	t              *transport.Transport
	pacingObserver PacingObserver // staging field; attached to PacingPolicy during init
	streamObserver StreamObserver // staging field; attached to Stream during Stream()
	pacingDisabled bool           // set by WithRateLimit(nil) to suppress default init
	pacing         *PacingPolicy
}

// NewClient creates a new IBKR API client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		t: &transport.Transport{
			HTTPClient: http.DefaultClient,
		},
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}
	if c.t.BaseURL == "" {
		return nil, ErrBaseURLRequired
	}

	// Initialize logger with "gbkr" group.
	if c.t.Logger == nil {
		c.t.Logger = slog.Default()
	}
	c.t.Logger = c.t.Logger.WithGroup("gbkr")

	// Initialize pacing policy if not explicitly set by WithRateLimit.
	if c.pacing == nil && !c.pacingDisabled {
		c.pacing = newDefaultPacingPolicy()
	}
	if c.pacing != nil && c.pacingObserver != nil {
		c.pacing.observer = c.pacingObserver
	}
	c.pacingObserver = nil // clear staging field

	// Wire pacing as a transport request hook.
	if c.pacing != nil {
		c.t.Hook = c.pacing
	}

	return c, nil
}

// Transport returns the underlying transport for use by subpackages within this module.
func (c *Client) Transport() *transport.Transport {
	return c.t
}

func (c *Client) doGet(ctx context.Context, path string, query url.Values, result any) error {
	return c.t.Get(ctx, path, query, result)
}

func (c *Client) doPost(ctx context.Context, path string, body any, result any) error {
	return c.t.Post(ctx, path, body, result)
}
