package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Client is the base IBKR API client. It holds transport configuration.
// Provides read-only capabilities and session elevation via [BrokerageSession].
type Client struct {
	baseURL        string
	httpClient     *http.Client
	pacingObserver PacingObserver // staging field; attached to PacingPolicy during init
	pacingDisabled bool           // set by WithRateLimit(nil) to suppress default init
	pacing         *PacingPolicy
}

// BrokerageClient provides capabilities requiring a brokerage session.
// Embeds [*Client], inheriting all read-only capabilities.
// Obtained via [Client.BrokerageSession].
type BrokerageClient struct {
	*Client
}

// NewClient creates a new IBKR API client with the given options.
func NewClient(opts ...Option) (*Client, error) {
	c := &Client{
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, fmt.Errorf("applying option: %w", err)
		}
	}
	if c.baseURL == "" {
		return nil, ErrBaseURLRequired
	}

	// Initialize pacing policy if not explicitly set by WithRateLimit.
	if c.pacing == nil && !c.pacingDisabled {
		c.pacing = newDefaultPacingPolicy()
	}
	if c.pacing != nil && c.pacingObserver != nil {
		c.pacing.observer = c.pacingObserver
	}
	c.pacingObserver = nil // clear staging field

	return c, nil
}

func (c *Client) doGet(ctx context.Context, path string, query url.Values, result any) error {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	return c.doRequest(req, result)
}

func (c *Client) doPost(ctx context.Context, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.doRequest(req, result)
}

func (c *Client) doRequest(req *http.Request, result any) error {
	if c.pacing != nil {
		path := req.URL.Path
		if err := c.pacing.waitForSlot(req.Context(), req.Method, path); err != nil {
			return err
		}
		defer c.pacing.releaseSlot(req.Method, path)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ErrAPI(resp.StatusCode, resp.Status, data)
	}

	if result != nil && len(data) > 0 {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}
