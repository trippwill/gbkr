package flex

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func writeXML(t *testing.T, w http.ResponseWriter, data []byte) {
	t.Helper()
	w.Header().Set("Content-Type", "application/xml")
	if _, err := w.Write(data); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

func TestSendRequest_Success(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "send_request_success.xml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/SendRequest" {
			t.Errorf("path = %q, want /SendRequest", r.URL.Path)
		}
		if got := r.URL.Query().Get("t"); got != "TESTTOKEN" {
			t.Errorf("token = %q, want %q", got, "TESTTOKEN")
		}
		if got := r.URL.Query().Get("q"); got != "QUERYID" {
			t.Errorf("queryID = %q, want %q", got, "QUERYID")
		}
		if got := r.URL.Query().Get("v"); got != "3" {
			t.Errorf("version = %q, want %q", got, "3")
		}
		writeXML(t, w, data)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	ref, err := c.SendRequest(context.Background(), "TESTTOKEN", "QUERYID")
	if err != nil {
		t.Fatalf("SendRequest: %v", err)
	}
	if ref != "1234567890" {
		t.Errorf("reference = %q, want %q", ref, "1234567890")
	}
}

func TestSendRequest_TokenExpired(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "send_request_token_expired.xml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXML(t, w, data)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	_, err = c.SendRequest(context.Background(), "EXPIRED", "QUERYID")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrTokenExpired) {
		t.Errorf("error = %v, want ErrTokenExpired", err)
	}
}

func TestSendRequest_HTTPStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "gateway offline", http.StatusBadGateway)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	_, err := c.SendRequest(context.Background(), "TOKEN", "QUERYID")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unexpected HTTP 502 Bad Gateway") {
		t.Fatalf("error = %v, want HTTP status context", err)
	}
	if !strings.Contains(err.Error(), "gateway offline") {
		t.Fatalf("error = %v, want body preview", err)
	}
}

func TestGetStatement_Success(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "activity_statement.xml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/GetStatement" {
			t.Errorf("path = %q, want /GetStatement", r.URL.Path)
		}
		writeXML(t, w, data)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	resp, err := c.GetStatement(context.Background(), "TOKEN", "REF123")
	if err != nil {
		t.Fatalf("GetStatement: %v", err)
	}
	if len(resp.Statements) != 1 {
		t.Fatalf("Statements = %d, want 1", len(resp.Statements))
	}
	if resp.Statements[0].AccountID != "U1234567" {
		t.Errorf("AccountID = %q, want %q", resp.Statements[0].AccountID, "U1234567")
	}
}

func TestGetStatement_HTTPStatusError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	_, err := c.GetStatement(context.Background(), "TOKEN", "REF123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "unexpected HTTP 503 Service Unavailable") {
		t.Fatalf("error = %v, want HTTP status context", err)
	}
	if !strings.Contains(err.Error(), "service unavailable") {
		t.Fatalf("error = %v, want body preview", err)
	}
}

func TestGetStatement_ResponseTooLarge(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		if _, err := w.Write([]byte(strings.Repeat("x", 33))); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(
		WithBaseURL(srv.URL+"/"),
		WithMaxResponseBytes(32),
	)
	_, err := c.GetStatement(context.Background(), "TOKEN", "REF123")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "exceeds configured limit of 32 bytes") {
		t.Fatalf("error = %v, want response size limit error", err)
	}
}

func TestGetStatement_WrongFormatSavesCSV(t *testing.T) {
	body := []byte("\"Date\",\"Symbol\"\n\"2026-03-21\",\"SPY\"\n")
	dir := t.TempDir()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/csv")
		if _, err := w.Write(body); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	c := NewClient(
		WithBaseURL(srv.URL+"/"),
		WithReportDir(dir),
	)
	_, err := c.GetStatement(context.Background(), "TOKEN", "../unsafe/ref")
	if !errors.Is(err, ErrWrongFormat) {
		t.Fatalf("GetStatement error = %v, want ErrWrongFormat", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("saved file count = %d, want 1", len(entries))
	}
	name := entries[0].Name()
	if strings.Contains(name, "/") || strings.Contains(name, "..") {
		t.Fatalf("saved filename %q is not sanitized", name)
	}
	if !strings.Contains(name, "unsafe_ref") {
		t.Fatalf("saved filename %q does not contain sanitized refcode", name)
	}
	if !strings.HasSuffix(name, "-csv.csv") {
		t.Fatalf("saved filename %q does not use CSV suffix/extension", name)
	}
}

func TestFetchReport_Success(t *testing.T) {
	sendData, err := os.ReadFile(filepath.Join("testdata", "send_request_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	stmtData, err := os.ReadFile(filepath.Join("testdata", "activity_statement.xml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/SendRequest":
			writeXML(t, w, sendData)
		case "/GetStatement":
			writeXML(t, w, stmtData)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	resp, err := c.FetchReport(context.Background(), "TOKEN", "QID",
		WithInitialDelay(10*time.Millisecond),
		WithMaxRetries(1),
	)
	if err != nil {
		t.Fatalf("FetchReport: %v", err)
	}
	if len(resp.Statements) != 1 {
		t.Fatalf("Statements = %d, want 1", len(resp.Statements))
	}
}

func TestFetchReport_RetryThenSuccess(t *testing.T) {
	sendData, err := os.ReadFile(filepath.Join("testdata", "send_request_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	stmtData, err := os.ReadFile(filepath.Join("testdata", "activity_statement.xml"))
	if err != nil {
		t.Fatal(err)
	}
	errorData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<FlexStatementResponse><Status>Warn</Status><ErrorCode>1019</ErrorCode><ErrorMessage>Statement generation in progress. Please try again shortly.</ErrorMessage></FlexStatementResponse>`)

	var getAttempts atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/SendRequest":
			writeXML(t, w, sendData)
		case "/GetStatement":
			n := getAttempts.Add(1)
			if n <= 2 {
				writeXML(t, w, errorData)
			} else {
				writeXML(t, w, stmtData)
			}
		}
	}))
	defer srv.Close()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	resp, err := c.FetchReport(context.Background(), "TOKEN", "QID",
		WithInitialDelay(10*time.Millisecond),
		WithBackoffMultiplier(1.0),
		WithMaxRetries(3),
	)
	if err != nil {
		t.Fatalf("FetchReport: %v", err)
	}
	if len(resp.Statements) != 1 {
		t.Fatalf("Statements = %d, want 1", len(resp.Statements))
	}
	if got := getAttempts.Load(); got != 3 {
		t.Errorf("GetStatement attempts = %d, want 3", got)
	}
}

func TestFetchReport_ContextCanceled(t *testing.T) {
	sendData, err := os.ReadFile(filepath.Join("testdata", "send_request_success.xml"))
	if err != nil {
		t.Fatal(err)
	}
	errorData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<FlexStatementResponse><Status>Warn</Status><ErrorCode>1019</ErrorCode><ErrorMessage>Not ready</ErrorMessage></FlexStatementResponse>`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/SendRequest":
			writeXML(t, w, sendData)
		case "/GetStatement":
			writeXML(t, w, errorData)
		}
	}))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	c := NewClient(WithBaseURL(srv.URL + "/"))
	_, err = c.FetchReport(ctx, "TOKEN", "QID",
		WithInitialDelay(200*time.Millisecond),
		WithMaxRetries(5),
	)
	if err == nil {
		t.Fatal("expected error from context cancellation")
	}
}

func TestSaveReport_UsesUniqueSafeNames(t *testing.T) {
	dir := t.TempDir()
	c := NewClient(WithReportDir(dir))

	c.saveReport(context.Background(), "../unsafe/ref", []byte("<xml/>"), "", ".xml")
	c.saveReport(context.Background(), "../unsafe/ref", []byte("<xml/>"), "", ".xml")

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("saved file count = %d, want 2", len(entries))
	}
	seen := map[string]struct{}{}
	for _, entry := range entries {
		name := entry.Name()
		if _, ok := seen[name]; ok {
			t.Fatalf("duplicate saved filename %q", name)
		}
		seen[name] = struct{}{}
		if strings.Contains(name, "/") || strings.Contains(name, "..") {
			t.Fatalf("saved filename %q is not sanitized", name)
		}
		if !strings.Contains(name, "unsafe_ref") {
			t.Fatalf("saved filename %q does not contain sanitized refcode", name)
		}
		if !strings.HasSuffix(name, ".xml") {
			t.Fatalf("saved filename %q does not use XML extension", name)
		}
	}
}

func TestGetStatement_WithReportDir_WritesXMLFile(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "activity_statement.xml"))
	if err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXML(t, w, data)
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := NewClient(WithBaseURL(srv.URL+"/"), WithReportDir(dir))
	_, err = c.GetStatement(context.Background(), "TOKEN", "REF999")
	if err != nil {
		t.Fatalf("GetStatement: %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("file count = %d, want 1", len(entries))
	}
	name := entries[0].Name()
	if !contains(name, "REF999") {
		t.Errorf("filename %q does not contain ref code", name)
	}
	if !hasSuffix(name, ".xml") {
		t.Errorf("filename %q does not end with .xml", name)
	}
	// Success files must not have an -err or -csv infix.
	if contains(name, "-err") || contains(name, "-csv") {
		t.Errorf("success file should have no error suffix, got %q", name)
	}
}

func TestGetStatement_CSVResponse_WritesCsvSuffixedFile(t *testing.T) {
	// IBKR returns a CSV body when the query Format is not set to XML.
	// The body starts with '"' and contains no '<'.
	csvBody := []byte(`"Status","DataSet","ErrorCode","ErrorMessage"` + "\r\n" + `"Success","","",""` + "\r\n")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		if _, err := w.Write(csvBody); err != nil {
			t.Fatalf("write response: %v", err)
		}
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := NewClient(WithBaseURL(srv.URL+"/"), WithReportDir(dir))
	_, err := c.GetStatement(context.Background(), "TOKEN", "CSVREF")
	if !errors.Is(err, ErrWrongFormat) {
		t.Fatalf("error = %v, want ErrWrongFormat", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("file count = %d, want 1", len(entries))
	}
	name := entries[0].Name()
	if !contains(name, "-csv") {
		t.Errorf("filename %q should contain -csv suffix, got %q", name, name)
	}
}

func TestGetStatement_ErrorResponse_WritesErrSuffixedFile(t *testing.T) {
	// IBKR returns a status/error XML response when the query fails
	// (e.g. expired token, query not found). This parses as a FlexError.
	errBody := []byte(`<?xml version="1.0" encoding="UTF-8"?>` +
		`<FlexStatementResponse><Status>Fail</Status>` +
		`<ErrorCode>1012</ErrorCode>` +
		`<ErrorMessage>Account not found or not permitted.</ErrorMessage>` +
		`</FlexStatementResponse>`)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeXML(t, w, errBody)
	}))
	defer srv.Close()

	dir := t.TempDir()
	c := NewClient(WithBaseURL(srv.URL+"/"), WithReportDir(dir))
	_, err := c.GetStatement(context.Background(), "TOKEN", "ERRREF")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("file count = %d, want 1", len(entries))
	}
	name := entries[0].Name()
	if !contains(name, "-err") {
		t.Errorf("filename %q should contain -err suffix", name)
	}
}

// contains reports whether substr appears in s.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr))
}

func containsAt(s, substr string) bool {
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// hasSuffix reports whether s ends with suffix.
func hasSuffix(s, suffix string) bool {
	return len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"report not ready", ErrReportNotReady, true},
		{"rate limited", ErrRateLimited, true},
		{"token expired", ErrTokenExpired, false},
		{"query not found", ErrQueryNotFound, false},
		{"response error 1019", ErrResponse(1019, "not ready"), true},
		{"response error 1018", ErrResponse(1018, "rate limit"), true},
		{"response error 1012", ErrResponse(1012, "expired"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsRetryable(tt.err); got != tt.want {
				t.Errorf("IsRetryable(%v) = %v, want %v", tt.err, got, tt.want)
			}
		})
	}
}
