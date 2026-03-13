package transport_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/coder/websocket"
)

// testWSServer starts an httptest.Server that upgrades connections to WebSocket
// and runs the provided handler. The returned server URL uses "ws://" scheme.
func testWSServer(t *testing.T, handler func(conn *websocket.Conn)) (*httptest.Server, string) {
	t.Helper()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			InsecureSkipVerify: true,
		})
		if err != nil {
			t.Logf("testWSServer: accept error: %v", err)
			return
		}
		handler(conn)
	}))

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	t.Cleanup(srv.Close)
	return srv, wsURL
}
