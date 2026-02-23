package gbkr

import (
	"net/http"
	"testing"
)

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithHTTPClient(custom),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient != custom {
		t.Error("WithHTTPClient did not set the client")
	}
}
