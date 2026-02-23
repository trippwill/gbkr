package gbkr

import (
	"crypto/tls"
	"net/http"
)

// Option configures a Client.
type Option func(*Client) error

// WithBaseURL sets the IBKR API base URL.
func WithBaseURL(url string) Option {
	return func(c *Client) error {
		c.baseURL = url
		return nil
	}
}

// WithHTTPClient sets a custom HTTP client for transport.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) error {
		c.httpClient = hc
		return nil
	}
}

// WithInsecureTLS creates an HTTP client that skips TLS verification.
// Useful for local gateway connections.
func WithInsecureTLS() Option {
	return func(c *Client) error {
		c.httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true, //nolint:gosec // intentional for local gateway
				},
			},
		}
		return nil
	}
}

// WithRateLimit sets a custom [PacingPolicy] on the client. Pass nil
// (or [NoPacing]) to disable pacing entirely.
func WithRateLimit(policy *PacingPolicy) Option {
	return func(c *Client) error {
		if policy == nil {
			c.pacingDisabled = true
			c.pacing = nil
		} else {
			c.pacing = policy
		}
		return nil
	}
}

// WithPacingObserver sets a [PacingObserver] that receives notifications
// about pacing waits, cache hits, and cache misses.
func WithPacingObserver(obs PacingObserver) Option {
	return func(c *Client) error {
		c.pacingObserver = obs
		return nil
	}
}
