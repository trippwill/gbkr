package flex

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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
