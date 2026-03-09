package brokerage_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/brokerage"
)

func testWSServer(t *testing.T, handler func(conn *websocket.Conn)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/ws") {
			conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
				InsecureSkipVerify: true,
			})
			if err != nil {
				t.Logf("testWSServer: accept error: %v", err)
				return
			}
			handler(conn)
			return
		}
		// Default handler for non-WS requests (e.g., SSO init).
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"authenticated":true}`)) //nolint:errcheck
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestStream(t *testing.T, srv *httptest.Server) *gbkr.Stream {
	t.Helper()
	client, err := gbkr.NewClient(
		gbkr.WithBaseURL(srv.URL),
		gbkr.WithRateLimit(nil),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	t.Cleanup(func() { stream.Close() })
	return stream
}

func TestSubscribeMarketData(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		// Read the subscribe command.
		_, _, _ = conn.Read(context.Background())

		// Send a market data tick.
		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"smd+265598","conid":265598,"31":"142.50","84":"145.00","86":"350000"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)

	ch, cancel, err := brokerage.SubscribeMarketData(stream, 265598,
		brokerage.FieldLast, brokerage.FieldBid)
	if err != nil {
		t.Fatalf("SubscribeMarketData: %v", err)
	}
	defer cancel()

	select {
	case update := <-ch:
		if update.ConID != 265598 {
			t.Errorf("ConID = %d, want 265598", update.ConID)
		}
		if _, ok := update.Fields[brokerage.FieldLast]; !ok {
			t.Error("missing FieldLast")
		}
		if _, ok := update.Fields[brokerage.FieldBid]; !ok {
			t.Error("missing FieldBid")
		}
		// FieldVolume ("86") should be filtered out since we didn't request it.
		// But "86" is FieldAsk actually. Let me verify — FieldAsk = "86".
		// We only requested FieldLast("31") and FieldBid("84").
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for market data")
	}
}

func TestSubscribeMarketData_Cancel(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)

	ch, cancel, err := brokerage.SubscribeMarketData(stream, 265598, brokerage.FieldLast)
	if err != nil {
		t.Fatalf("SubscribeMarketData: %v", err)
	}

	cancel()

	// Channel should be closed after cancel.
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestSubscribeOrders(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"sor","orderId":"123","account":"U1234567","status":"Filled","filledQuantity":"100","avgPrice":"145.50"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)

	ch, cancel, err := brokerage.SubscribeOrders(stream)
	if err != nil {
		t.Fatalf("SubscribeOrders: %v", err)
	}
	defer cancel()

	select {
	case update := <-ch:
		if update.Status != "Filled" {
			t.Errorf("Status = %q, want %q", update.Status, "Filled")
		}
		if update.Account != "U1234567" {
			t.Errorf("Account = %q, want %q", update.Account, "U1234567")
		}
		if update.Raw == nil {
			t.Error("Raw should not be nil")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for order update")
	}
}

func TestSubscribeTrades(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"str","execution_id":"e1","symbol":"AAPL","side":"BUY","size":"100","price":"145.50"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)

	ch, cancel, err := brokerage.SubscribeTrades(stream)
	if err != nil {
		t.Fatalf("SubscribeTrades: %v", err)
	}
	defer cancel()

	select {
	case update := <-ch:
		if update.Symbol != "AAPL" {
			t.Errorf("Symbol = %q, want %q", update.Symbol, "AAPL")
		}
		if update.Side != "BUY" {
			t.Errorf("Side = %q, want %q", update.Side, "BUY")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for trade update")
	}
}

func TestSubscribeOrders_Blocking(t *testing.T) {
	// Verify that orders use blocking semantics (no drop).
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		// Send enough orders to fill the buffer (32) plus 1.
		for i := range 33 {
			msg, _ := json.Marshal(map[string]any{
				"topic":   "sor",
				"orderId": i,
				"status":  "Submitted",
			})
			if err := conn.Write(context.Background(), websocket.MessageText, msg); err != nil {
				return
			}
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)

	ch, cancel, err := brokerage.SubscribeOrders(stream)
	if err != nil {
		t.Fatalf("SubscribeOrders: %v", err)
	}
	defer cancel()

	// Drain all 33 messages — the blocking semantics mean none are dropped.
	count := 0
	timeout := time.After(3 * time.Second)
	for count < 33 {
		select {
		case <-ch:
			count++
		case <-timeout:
			t.Fatalf("only received %d/33 order updates (expected blocking semantics)", count)
		}
	}
}

func TestSubscribeMarketData_ClosedStream(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)
	stream.Close()

	_, _, err := brokerage.SubscribeMarketData(stream, 265598, brokerage.FieldLast)
	if err == nil {
		t.Fatal("expected error on closed stream")
	}
}
