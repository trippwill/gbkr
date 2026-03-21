package flex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"
)

const defaultBaseURL = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService/"

// Client communicates with the IBKR Flex Web Service.
// Create one with [NewClient].
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
}

// Option configures a [Client].
type Option func(*Client)

// WithBaseURL overrides the default Flex Web Service URL.
// Useful for testing against a local server.
func WithBaseURL(u string) Option {
	return func(c *Client) { c.baseURL = u }
}

// WithHTTPClient sets the HTTP client used for requests.
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// WithLogger sets the structured logger for operation events.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) { c.logger = l }
}

// NewClient creates a Flex Web Service client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
		logger:     slog.Default(),
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

// SendRequest initiates report generation for the given query ID and returns
// a reference code for use with [Client.GetStatement].
func (c *Client) SendRequest(ctx context.Context, token, queryID string) (string, error) {
	start := time.Now()

	u, err := url.JoinPath(c.baseURL, "SendRequest")
	if err != nil {
		return "", fmt.Errorf("build send-request URL: %w", err)
	}

	q := url.Values{
		"t": {token},
		"q": {queryID},
		"v": {"3"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+q.Encode(), nil)
	if err != nil {
		return "", fmt.Errorf("create send-request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.emitOp(ctx, "FlexSendRequest", err, time.Since(start))
		return "", fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	parsed, err := parseSendRequestResponse(resp.Body)
	if err != nil {
		c.emitOp(ctx, "FlexSendRequest", err, time.Since(start))
		return "", err
	}

	if parsed.ErrorCode != 0 {
		rerr := ErrResponse(parsed.ErrorCode, parsed.ErrorMessage)
		c.emitOp(ctx, "FlexSendRequest", rerr, time.Since(start),
			slog.Int("errorCode", parsed.ErrorCode))
		return "", rerr
	}

	if parsed.ReferenceCode == "" {
		err := fmt.Errorf("send request returned empty reference code")
		c.emitOp(ctx, "FlexSendRequest", err, time.Since(start))
		return "", err
	}

	c.emitOp(ctx, "FlexSendRequest", nil, time.Since(start),
		slog.String("ref", parsed.ReferenceCode))
	return parsed.ReferenceCode, nil
}

// GetStatement retrieves a previously requested report by reference code.
// Returns [ErrReportNotReady] if the report is still being generated.
func (c *Client) GetStatement(ctx context.Context, token, referenceCode string) (*QueryResponse, error) {
	start := time.Now()

	u, err := url.JoinPath(c.baseURL, "GetStatement")
	if err != nil {
		return nil, fmt.Errorf("build get-statement URL: %w", err)
	}

	q := url.Values{
		"t": {token},
		"q": {referenceCode},
		"v": {"3"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u+"?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create get-statement: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.emitOp(ctx, "FlexGetStatement", err, time.Since(start))
		return nil, fmt.Errorf("get statement: %w", err)
	}
	defer resp.Body.Close()

	// Buffer the body so we can try parsing as both a report and an error response.
	// The GetStatement endpoint uses <FlexQueryResponse> for success and
	// <FlexStatementResponse> for errors — different root elements.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.emitOp(ctx, "FlexGetStatement", err, time.Since(start))
		return nil, fmt.Errorf("read get-statement body: %w", err)
	}

	// Try parsing as a full report first.
	result, err := ParseActivityStatement(bytes.NewReader(body))
	if err == nil {
		c.emitOp(ctx, "FlexGetStatement", nil, time.Since(start),
			slog.Int("statements", len(result.Statements)))
		return result, nil
	}

	// Parse failure — check if it's an error/status response.
	errResp, parseErr := parseSendRequestResponse(bytes.NewReader(body))
	if parseErr == nil && errResp.ErrorCode != 0 {
		rerr := ErrResponse(errResp.ErrorCode, errResp.ErrorMessage)
		c.emitOp(ctx, "FlexGetStatement", rerr, time.Since(start),
			slog.Int("errorCode", errResp.ErrorCode))
		return nil, rerr
	}

	// Neither parse succeeded — return the original parse error.
	c.emitOp(ctx, "FlexGetStatement", err, time.Since(start))
	return nil, err
}

// FetchOption configures the retry behavior of [Client.FetchReport].
type FetchOption func(*fetchConfig)

type fetchConfig struct {
	maxRetries   int
	initialDelay time.Duration
	backoffMult  float64
}

// WithMaxRetries sets the maximum number of GetStatement retries (default 3).
func WithMaxRetries(n int) FetchOption {
	return func(fc *fetchConfig) { fc.maxRetries = n }
}

// WithInitialDelay sets the delay before the first GetStatement attempt (default 5s).
func WithInitialDelay(d time.Duration) FetchOption {
	return func(fc *fetchConfig) { fc.initialDelay = d }
}

// WithBackoffMultiplier sets the backoff multiplier between retries (default 2.0).
func WithBackoffMultiplier(m float64) FetchOption {
	return func(fc *fetchConfig) { fc.backoffMult = m }
}

// FetchReport runs the full two-step retrieval protocol: SendRequest followed
// by polling GetStatement with exponential backoff until the report is ready
// or retries are exhausted.
func (c *Client) FetchReport(ctx context.Context, token, queryID string, opts ...FetchOption) (*QueryResponse, error) {
	cfg := fetchConfig{
		maxRetries:   3,
		initialDelay: 5 * time.Second,
		backoffMult:  2.0,
	}
	for _, o := range opts {
		o(&cfg)
	}

	ref, err := c.SendRequest(ctx, token, queryID)
	if err != nil {
		return nil, err
	}

	delay := cfg.initialDelay
	for attempt := range cfg.maxRetries + 1 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
		}

		result, err := c.GetStatement(ctx, token, ref)
		if err == nil {
			return result, nil
		}

		if !IsRetryable(err) {
			return nil, err
		}

		c.logger.LogAttrs(ctx, slog.LevelInfo, "flex report not ready, retrying",
			slog.Int("attempt", attempt+1),
			slog.Duration("nextDelay", time.Duration(float64(delay)*cfg.backoffMult)),
		)
		delay = time.Duration(float64(delay) * cfg.backoffMult)
	}

	return nil, fmt.Errorf("flex report not ready after %d retries: %w", cfg.maxRetries, ErrReportNotReady)
}

// emitOp logs a structured operation event. Mirrors the gbkr emitOp pattern.
func (c *Client) emitOp(ctx context.Context, op string, err error, dur time.Duration, attrs ...slog.Attr) {
	level := slog.LevelInfo
	allAttrs := make([]slog.Attr, 0, len(attrs)+3)
	allAttrs = append(allAttrs, slog.String("op", op), slog.Duration("dur", dur))
	if err != nil {
		level = slog.LevelWarn
		allAttrs = append(allAttrs, slog.String("error", err.Error()))
	}
	allAttrs = append(allAttrs, attrs...)
	c.logger.LogAttrs(ctx, level, "flex."+op, allAttrs...)
}
