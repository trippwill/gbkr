package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

// Client is the core IBKR API client. It holds transport configuration
// and granted permissions but exposes no API methods directly.
// Use capability constructors (Auth, Accounts, Positions, MarketData)
// to obtain typed interfaces.
type Client struct {
	baseURL     string
	httpClient  *http.Client
	permissions PermissionSet
	prompter    Prompter
	mu          sync.Mutex
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
	return c, nil
}

// Permissions returns the client's granted permission set.
func (c *Client) Permissions() PermissionSet {
	return c.permissions
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
