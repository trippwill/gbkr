package flex

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
)

const (
	defaultBaseURL                = "https://ndcdyn.interactivebrokers.com/AccountManagement/FlexWebService/"
	defaultMaxResponseBytes int64 = 64 << 20
	httpErrorPreviewBytes   int64 = 4 << 10
)

var reportFileSeq atomic.Uint64

// Client communicates with the IBKR Flex Web Service.
// Create one with [NewClient].
type Client struct {
	baseURL          string
	httpClient       *http.Client
	logger           *slog.Logger
	reportDir        string
	maxResponseBytes int64
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

// WithReportDir sets a directory path where raw GetStatement response bodies
// are saved to disk before parsing. Files are named
// flex-<timestamp>-<seq>-<sanitized-refcode>[-err|-csv].<xml|csv>. This is
// useful for debugging unexpected parse failures or inspecting query output.
// The directory is created if it does not exist.
// Pass an empty string to disable (the default).
func WithReportDir(dir string) Option {
	return func(c *Client) { c.reportDir = dir }
}

// WithMaxResponseBytes bounds how many bytes [Client.GetStatement] buffers
// before parsing and optional report saving. Values less than 1 leave the
// default limit in place.
func WithMaxResponseBytes(n int64) Option {
	return func(c *Client) {
		if n > 0 {
			c.maxResponseBytes = n
		}
	}
}

// NewClient creates a Flex Web Service client.
func NewClient(opts ...Option) *Client {
	c := &Client{
		baseURL:          defaultBaseURL,
		httpClient:       http.DefaultClient,
		logger:           slog.Default(),
		maxResponseBytes: defaultMaxResponseBytes,
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

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		preview, truncated, readErr := readPreview(resp.Body, httpErrorPreviewBytes)
		if readErr != nil {
			c.emitOp(ctx, "FlexSendRequest", readErr, time.Since(start),
				slog.Int("statusCode", resp.StatusCode))
			return "", fmt.Errorf("read send-request error body: %w", readErr)
		}
		err = formatHTTPError("send request", resp.StatusCode, resp.Status, preview, truncated)
		c.emitOp(ctx, "FlexSendRequest", err, time.Since(start),
			slog.Int("statusCode", resp.StatusCode),
			slog.Int("bodyLen", len(preview)),
			slog.Bool("bodyPreviewTruncated", truncated))
		return "", err
	}

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

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		preview, truncated, readErr := readPreview(resp.Body, httpErrorPreviewBytes)
		if readErr != nil {
			c.emitOp(ctx, "FlexGetStatement", readErr, time.Since(start),
				slog.Int("statusCode", resp.StatusCode))
			return nil, fmt.Errorf("read get-statement error body: %w", readErr)
		}
		err = formatHTTPError("get statement", resp.StatusCode, resp.Status, preview, truncated)
		c.emitOp(ctx, "FlexGetStatement", err, time.Since(start),
			slog.Int("statusCode", resp.StatusCode),
			slog.Int("bodyLen", len(preview)),
			slog.Bool("bodyPreviewTruncated", truncated))
		return nil, err
	}

	// Buffer the body so we can try parsing as both a report and an error response.
	// The GetStatement endpoint uses <FlexQueryResponse> for success and
	// <FlexStatementResponse> for errors — different root elements.
	body, truncated, err := readPreview(resp.Body, c.maxResponseBytes)
	if err != nil {
		c.emitOp(ctx, "FlexGetStatement", err, time.Since(start))
		return nil, fmt.Errorf("read get-statement body: %w", err)
	}
	if truncated {
		err = fmt.Errorf("get statement response exceeds configured limit of %d bytes", c.maxResponseBytes)
		c.emitOp(ctx, "FlexGetStatement", err, time.Since(start),
			slog.Int64("maxResponseBytes", c.maxResponseBytes),
			slog.Int("bodyLen", len(body)))
		return nil, err
	}

	// Empty body means the report is still being generated.
	if len(body) == 0 {
		c.emitOp(ctx, "FlexGetStatement", ErrReportNotReady, time.Since(start),
			slog.Int("bodyLen", 0))
		return nil, ErrReportNotReady
	}

	// Try parsing as a full report first.
	result, err := ParseActivityStatement(bytes.NewReader(body))
	if err == nil {
		c.saveReport(ctx, referenceCode, body, "", ".xml")
		c.emitOp(ctx, "FlexGetStatement", nil, time.Since(start),
			slog.Int("statements", len(result.Statements)))
		return result, nil
	}

	// Parse failure — check if it's an error/status response.
	errResp, parseErr := parseSendRequestResponse(bytes.NewReader(body))
	if parseErr == nil && errResp.ErrorCode != 0 {
		rerr := ErrResponse(errResp.ErrorCode, errResp.ErrorMessage)
		c.saveReport(ctx, referenceCode, body, "-err", ".xml")
		c.emitOp(ctx, "FlexGetStatement", rerr, time.Since(start),
			slog.Int("errorCode", errResp.ErrorCode))
		return nil, rerr
	}

	// Neither parse succeeded — check for common misconfigurations before
	// falling through to a generic error.

	// CSV output: IBKR returns quoted comma-separated values when the query
	// Format is not set to XML. The response starts with '"' and contains no '<'.
	if len(body) > 0 && body[0] == '"' && !bytes.Contains(body, []byte{'<'}) {
		c.saveReport(ctx, referenceCode, body, "-csv", ".csv")
		c.emitOp(ctx, "FlexGetStatement", ErrWrongFormat, time.Since(start),
			slog.Int("bodyLen", len(body)))
		return nil, ErrWrongFormat
	}

	// Save body for inspection and log prefix for context.
	c.saveReport(ctx, referenceCode, body, "-err", ".xml")
	preview := string(body)
	if len(preview) > 200 {
		preview = preview[:200]
	}
	c.emitOp(ctx, "FlexGetStatement", err, time.Since(start),
		slog.Int("bodyLen", len(body)),
		slog.String("bodyPrefix", preview))
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
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, ctx.Err()
		case <-timer.C:
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

// saveReport writes body to reportDir using a timestamped unique filename.
// nameSuffix should be "" on success, "-err" on parse failure, or "-csv" when
// persisting a CSV diagnostic response.
// Errors writing the file are logged but not returned.
func (c *Client) saveReport(ctx context.Context, refCode string, body []byte, nameSuffix, ext string) {
	if c.reportDir == "" {
		return
	}
	if err := os.MkdirAll(c.reportDir, 0o750); err != nil {
		c.logger.WarnContext(ctx, "flex: could not create report dir", "dir", c.reportDir, "error", err)
		return
	}
	if ext == "" {
		ext = ".xml"
	}
	ts := time.Now().UTC().Format("20060102-150405.000000000")
	seq := reportFileSeq.Add(1)
	name := fmt.Sprintf("flex-%s-%06d-%s%s%s", ts, seq, sanitizeFilenamePart(refCode), nameSuffix, ext)
	path := filepath.Join(c.reportDir, name)
	if err := os.WriteFile(path, body, 0o600); err != nil {
		c.logger.WarnContext(ctx, "flex: could not save report", "path", path, "error", err)
		return
	}
	c.logger.InfoContext(ctx, "flex: saved report", "path", path, "bytes", len(body))
}

func readPreview(r io.Reader, limit int64) ([]byte, bool, error) {
	data, err := io.ReadAll(io.LimitReader(r, limit+1))
	if err != nil {
		return nil, false, err
	}
	truncated := int64(len(data)) > limit
	if truncated {
		data = data[:limit]
	}
	return data, truncated, nil
}

func formatHTTPError(action string, statusCode int, status string, preview []byte, truncated bool) error {
	if status == "" {
		status = fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	}
	bodyPreview := strings.Join(strings.Fields(string(preview)), " ")
	if bodyPreview == "" {
		return fmt.Errorf("%s: unexpected HTTP %s", action, status)
	}
	if truncated {
		bodyPreview += "..."
	}
	return fmt.Errorf("%s: unexpected HTTP %s: %s", action, status, bodyPreview)
}

func sanitizeFilenamePart(s string) string {
	if s == "" {
		return "report"
	}
	var b strings.Builder
	b.Grow(len(s))
	lastUnderscore := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '-',
			r == '_':
			b.WriteRune(r)
			lastUnderscore = false
		default:
			if !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
	}
	sanitized := strings.Trim(b.String(), "_")
	if sanitized == "" {
		return "report"
	}
	return sanitized
}
