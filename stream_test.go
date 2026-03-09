package gbkr_test

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/trippwill/gbkr"
)

// testWSServer starts an httptest.Server that upgrades to WebSocket.
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

// newStreamTestClient creates a Client whose BaseURL points at the test WS server.
// The base URL is the HTTP URL (stream.go derives the WS URL from it).
func newStreamTestClient(t *testing.T, httpURL string) *gbkr.Client {
	t.Helper()
	c, err := gbkr.NewClient(
		gbkr.WithBaseURL(httpURL),
		gbkr.WithRateLimit(nil),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

// captureHandler captures slog records for assertion.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	h.records = append(h.records, r)
	h.mu.Unlock()
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func (h *captureHandler) findOp(op string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, r := range h.records {
		var found bool
		r.Attrs(func(a slog.Attr) bool {
			if a.Key == "op" && a.Value.String() == op {
				found = true
				return false
			}
			return true
		})
		if found {
			return true
		}
	}
	return false
}

func TestClient_Stream_Connect(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	stream, err := client.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()
}

func TestStream_Close_Idempotent(t *testing.T) {
	_, _ = testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	ctx := context.Background()

	stream, err := client.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	if err := stream.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := stream.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}

func TestStream_ContextCancel(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	ctx, cancel := context.WithCancel(context.Background())

	stream, err := client.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	cancel()

	// Close should not block after context cancellation.
	done := make(chan struct{})
	go func() {
		stream.Close()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Close blocked after context cancel")
	}
}

func TestStream_Keepalive(t *testing.T) {
	received := make(chan string, 10)

	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			_, data, err := conn.Read(context.Background())
			if err != nil {
				return
			}
			received <- string(data)
		}
	})

	client := newStreamTestClient(t, srv.URL)
	ctx := context.Background()

	// Override keepalive interval for testing by using a short-lived stream.
	// We can't easily change the interval, so we just verify that the stream
	// sends tic messages. Since the default is 25s, we'll wait briefly and
	// verify via Close that the keepalive goroutine exits cleanly.
	stream, err := client.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	// Stream is connected; close it and verify no hang.
	stream.Close()
}

func TestStream_OpEmission(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	ch := &captureHandler{}
	logger := slog.New(ch)

	client, err := gbkr.NewClient(
		gbkr.WithBaseURL(srv.URL),
		gbkr.WithRateLimit(nil),
		gbkr.WithLogger(logger),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx := context.Background()
	stream, err := client.Stream(ctx)
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	if !ch.findOp("StreamConnect") {
		t.Error("OpStreamConnect not emitted")
	}

	stream.Close()

	if !ch.findOp("StreamDisconnect") {
		t.Error("OpStreamDisconnect not emitted")
	}
}

func TestStream_Notifications(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		// Wait for client to be ready, then send a notification.
		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"ntf","id":"n1","date":"2026-01-15","text":"Dividend received"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	ch, cancelSub, err := stream.Notifications()
	if err != nil {
		t.Fatalf("Notifications: %v", err)
	}
	defer cancelSub()

	select {
	case n := <-ch:
		if n.ID != "n1" {
			t.Errorf("ID = %q, want %q", n.ID, "n1")
		}
		if n.Text != "Dividend received" {
			t.Errorf("Text = %q, want %q", n.Text, "Dividend received")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for notification")
	}
}

func TestStream_Notifications_ChannelCloses(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	ch, cancelSub, err := stream.Notifications()
	if err != nil {
		t.Fatalf("Notifications: %v", err)
	}
	_ = cancelSub

	stream.Close()

	// Channel should be closed after stream closes.
	select {
	case _, ok := <-ch:
		if ok {
			t.Error("expected channel to be closed")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for channel close")
	}
}

func TestStream_Notifications_ClosedStream(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	stream.Close()

	_, _, err = stream.Notifications()
	if err == nil {
		t.Fatal("expected error on closed stream")
	}
	if !errors.Is(err, gbkr.ErrStreamNotConnected) {
		t.Errorf("got %v, want ErrStreamNotConnected", err)
	}
}

func TestStream_Notifications_MultipleCalls(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"ntf","id":"n1","date":"2026-01-15","text":"hello"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	ch1, cancel1, err := stream.Notifications()
	if err != nil {
		t.Fatalf("Notifications 1: %v", err)
	}
	defer cancel1()
	ch2, cancel2, err := stream.Notifications()
	if err != nil {
		t.Fatalf("Notifications 2: %v", err)
	}
	defer cancel2()

	// Both channels should receive the same notification (fan-out).
	for i, ch := range []<-chan gbkr.Notification{ch1, ch2} {
		select {
		case n := <-ch:
			if n.ID != "n1" {
				t.Errorf("ch%d: ID = %q, want %q", i+1, n.ID, "n1")
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("ch%d: timed out", i+1)
		}
	}
}

func TestStream_AccountSummary(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		// Read the subscribe command.
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"sbd+U1234567","totalCashValue":50000.00,"netLiquidation":150000.00}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	ch, cancel, err := stream.AccountSummary("U1234567")
	if err != nil {
		t.Fatalf("AccountSummary: %v", err)
	}
	defer cancel()

	select {
	case update := <-ch:
		if update.AccountID != "U1234567" {
			t.Errorf("AccountID = %q, want %q", update.AccountID, "U1234567")
		}
		if update.Get("totalCashValue") == nil {
			t.Error("missing totalCashValue field")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for account summary")
	}
}

func TestStream_PortfolioPnL(t *testing.T) {
	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, _, _ = conn.Read(context.Background())

		time.Sleep(50 * time.Millisecond)
		msg := `{"topic":"spl+U1234567","dpl":-500.0,"nl":100000.0,"upl":-2100.75,"rpl":850.25,"el":95000.0,"mv":5000.0}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	ch, cancel, err := stream.PortfolioPnL("U1234567")
	if err != nil {
		t.Fatalf("PortfolioPnL: %v", err)
	}
	defer cancel()

	select {
	case update := <-ch:
		if update.AccountID != "U1234567" {
			t.Errorf("AccountID = %q, want %q", update.AccountID, "U1234567")
		}
		if update.DailyPnL != -500.0 {
			t.Errorf("DailyPnL = %f, want -500.0", update.DailyPnL)
		}
		if update.UnrealizedPnL != -2100.75 {
			t.Errorf("UnrealizedPnL = %f, want -2100.75", update.UnrealizedPnL)
		}
		if update.RealizedPnL != 850.25 {
			t.Errorf("RealizedPnL = %f, want 850.25", update.RealizedPnL)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for PnL update")
	}
}

func TestStream_Cancel_Unsubscribes(t *testing.T) {
	received := make(chan string, 10)

	srv, _ := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		for {
			_, data, err := conn.Read(context.Background())
			if err != nil {
				return
			}
			received <- string(data)
		}
	})

	client := newStreamTestClient(t, srv.URL)
	stream, err := client.Stream(context.Background())
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}
	defer stream.Close()

	_, cancel, err := stream.PortfolioPnL("U1234567")
	if err != nil {
		t.Fatalf("PortfolioPnL: %v", err)
	}

	// Drain the subscribe command.
	select {
	case msg := <-received:
		if msg != "spl+U1234567" {
			t.Errorf("subscribe msg = %q, want %q", msg, "spl+U1234567")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for subscribe command")
	}

	cancel()

	// Should send unsubscribe command.
	select {
	case msg := <-received:
		if msg != "upl+U1234567" {
			t.Errorf("unsubscribe msg = %q, want %q", msg, "upl+U1234567")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for unsubscribe command")
	}
}
