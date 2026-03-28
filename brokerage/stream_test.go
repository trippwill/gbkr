package brokerage_test

import (
	"context"
	"encoding/json"
	"fmt"
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
		// Field "86" should be filtered out since we only requested FieldLast and FieldBid.
		if _, ok := update.Fields[brokerage.SnapshotField("86")]; ok {
			t.Error("field 86 should have been filtered out")
		}
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

func TestSubscribeMarketDataHistory(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		// Read the subscribe command.
		_, _, _ = conn.Read(context.Background())

		// Send a historical bar.
		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"smh+265598","t":"20260328 16:00:00","o":"142.50","h":"145.00","l":"141.25","c":"144.75","v":"1250000"}`
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

	ch, cancel, err := brokerage.SubscribeMarketDataHistory(stream, 265598)
	if err != nil {
		t.Fatalf("SubscribeMarketDataHistory: %v", err)
	}
	defer cancel()

	select {
	case bar := <-ch:
		if bar.ConID != 265598 {
			t.Errorf("ConID = %d, want 265598", bar.ConID)
		}
		if bar.Time != "20260328 16:00:00" {
			t.Errorf("Time = %q, want %q", bar.Time, "20260328 16:00:00")
		}
		if len(bar.Open) == 0 {
			t.Error("Open should not be empty")
		}
		if len(bar.High) == 0 {
			t.Error("High should not be empty")
		}
		if len(bar.Low) == 0 {
			t.Error("Low should not be empty")
		}
		if len(bar.Close) == 0 {
			t.Error("Close should not be empty")
		}
		if len(bar.Volume) == 0 {
			t.Error("Volume should not be empty")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for history bar")
	}
}

func TestSubscribeMarketDataHistory_Cancel(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	stream := newTestStream(t, srv)

	ch, cancel, err := brokerage.SubscribeMarketDataHistory(stream, 265598)
	if err != nil {
		t.Fatalf("SubscribeMarketDataHistory: %v", err)
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

func TestSubscribeMarketDataHistory_ClosedStream(t *testing.T) {
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

	_, _, err := brokerage.SubscribeMarketDataHistory(stream, 265598)
	if err == nil {
		t.Fatal("expected error on closed stream")
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

func TestSubscribeOrders_ClosedStream(t *testing.T) {
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

	_, _, err := brokerage.SubscribeOrders(stream)
	if err == nil {
		t.Fatal("expected error on closed stream")
	}
}

func TestSubscribeTrades_ClosedStream(t *testing.T) {
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

	_, _, err := brokerage.SubscribeTrades(stream)
	if err == nil {
		t.Fatal("expected error on closed stream")
	}
}

func TestSubscribeMarketData_NoMatchingFields(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		// Send a tick with only unrequested fields (we request FieldLast="31").
		noMatch := `{"topic":"smd+265598","conid":265598,"999":"42.00"}`
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(noMatch))

		// Then send a valid tick to verify subscription still works.
		time.Sleep(50 * time.Millisecond)
		valid := `{"topic":"smd+265598","conid":265598,"31":"150.00"}`
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(valid))

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
	defer cancel()

	// The no-match tick should be silently filtered; we receive only the valid one.
	select {
	case update := <-ch:
		if _, ok := update.Fields[brokerage.FieldLast]; !ok {
			t.Error("missing FieldLast in valid update")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for update after no-match tick")
	}
}

func TestSubscribeMarketData_Overflow(t *testing.T) {
	const total = 70
	serverDone := make(chan struct{})

	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		for i := range total {
			msg := fmt.Sprintf(`{"topic":"smd+265598","conid":265598,"31":"%d.00"}`, i)
			if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
				return
			}
		}
		close(serverDone)

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
	defer cancel()

	// Wait for server to finish sending, then let callbacks process.
	select {
	case <-serverDone:
	case <-time.After(3 * time.Second):
		t.Fatal("server didn't finish sending")
	}
	time.Sleep(300 * time.Millisecond)

	// Drain the channel — should have at most 64 items (buffer size).
	count := 0
drain:
	for {
		select {
		case <-ch:
			count++
		default:
			break drain
		}
	}

	if count == 0 {
		t.Error("expected some updates")
	}
	if count > 64 {
		t.Errorf("received %d updates, expected at most 64 (buffer size)", count)
	}
	t.Logf("received %d/%d updates (overflow dropped %d)", count, total, total-count)
}

func TestSubscribeOrders_ParseError(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		// Send type-mismatched JSON: account as number instead of string.
		bad := `{"topic":"sor","account":123,"status":"Filled"}`
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(bad))

		time.Sleep(50 * time.Millisecond)
		good := `{"topic":"sor","orderId":"456","account":"U1234567","status":"Submitted"}`
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(good))

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

	// Bad message should be skipped; good message should arrive.
	select {
	case update := <-ch:
		if update.Account != "U1234567" {
			t.Errorf("Account = %q, want %q", update.Account, "U1234567")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for order update after parse error")
	}
}

func TestSubscribeTrades_ParseError(t *testing.T) {
	srv := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		// Send type-mismatched JSON: symbol as number instead of string.
		bad := `{"topic":"str","symbol":123}`
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(bad))

		time.Sleep(50 * time.Millisecond)
		good := `{"topic":"str","execution_id":"e2","symbol":"MSFT","side":"SELL","size":"50","price":"320.00"}`
		_ = conn.Write(context.Background(), websocket.MessageText, []byte(good))

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
		if update.Symbol != "MSFT" {
			t.Errorf("Symbol = %q, want %q", update.Symbol, "MSFT")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for trade update after parse error")
	}
}
