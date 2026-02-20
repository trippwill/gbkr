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

// WithPermissions sets a static permission set.
func WithPermissions(ps PermissionSet) Option {
	return func(c *Client) error {
		c.permissions = ps
		return nil
	}
}

// WithPermissionsFromFile loads permissions from a YAML file.
func WithPermissionsFromFile(path string) Option {
	return func(c *Client) error {
		ps, err := LoadPermissionsFromFile(path)
		if err != nil {
			return err
		}
		c.permissions = ps
		return nil
	}
}

// WithPrompter sets a [Prompter] for JIT permission granting. When a
// capability constructor needs permissions not in the static set, the
// prompter is called with the missing permissions.
func WithPrompter(p Prompter) Option {
	return func(c *Client) error {
		c.prompter = p
		return nil
	}
}

// WithInteractivePrompt enables JIT permission prompting via stderr/stdin.
// Convenience wrapper for WithPrompter with an [InteractivePrompter].
func WithInteractivePrompt() Option {
	return WithPrompter(InteractivePrompter{})
}
