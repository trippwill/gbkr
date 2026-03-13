package transport_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/trippwill/gbkr/internal/transport"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

func TestDialWS_Success(t *testing.T) {
	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		// Hold connection open until client disconnects.
		for {
			_, _, err := conn.Read(context.Background())
			if err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}
	defer ws.Close()

	select {
	case <-ws.Done():
		t.Fatal("Done() should not be closed on a connected WSConn")
	default:
	}
}

func TestDialWS_BadURL(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := transport.DialWS(ctx, "ws://127.0.0.1:0/nope", http.DefaultClient, testLogger())
	if err == nil {
		t.Fatal("expected error for bad URL")
	}
	if !errors.Is(err, transport.ErrWSClosed) {
		t.Errorf("expected ErrWSClosed, got %v", err)
	}
}

func TestWSConn_Send(t *testing.T) {
	received := make(chan string, 1)

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		_, data, err := conn.Read(context.Background())
		if err != nil {
			return
		}
		received <- string(data)
		// Drain remaining reads so close handshake works.
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}
	defer ws.Close()

	if err := ws.Send(ctx, "tic"); err != nil {
		t.Fatalf("Send: %v", err)
	}

	select {
	case msg := <-received:
		if msg != "tic" {
			t.Errorf("got %q, want %q", msg, "tic")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for message")
	}
}

func TestWSConn_Subscribe(t *testing.T) {
	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		msg := `{"topic":"ntf","id":"1","text":"hello"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		// Read until client closes.
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}
	defer ws.Close()

	received := make(chan json.RawMessage, 1)
	cancelSub, _ := ws.Subscribe("ntf", func(data json.RawMessage) {
		received <- data
	})
	defer cancelSub()

	select {
	case data := <-received:
		var msg struct {
			Topic string `json:"topic"`
			Text  string `json:"text"`
		}
		if err := json.Unmarshal(data, &msg); err != nil {
			t.Fatalf("Unmarshal: %v", err)
		}
		if msg.Text != "hello" {
			t.Errorf("text = %q, want %q", msg.Text, "hello")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for dispatch")
	}
}

func TestWSConn_Subscribe_MultipleHandlers(t *testing.T) {
	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		msg := `{"topic":"ntf","id":"1","text":"fanout"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}
	defer ws.Close()

	var count atomic.Int32
	var wg sync.WaitGroup
	wg.Add(2)

	cancel1, _ := ws.Subscribe("ntf", func(_ json.RawMessage) {
		count.Add(1)
		wg.Done()
	})
	defer cancel1()

	cancel2, _ := ws.Subscribe("ntf", func(_ json.RawMessage) {
		count.Add(1)
		wg.Done()
	})
	defer cancel2()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		if c := count.Load(); c != 2 {
			t.Errorf("handler count = %d, want 2", c)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for both handlers")
	}
}

func TestWSConn_Subscribe_Cancel(t *testing.T) {
	ready := make(chan struct{})

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		<-ready
		msg := `{"topic":"ntf","id":"1","text":"after-cancel"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}
	defer ws.Close()

	called := atomic.Bool{}
	cancelSub, _ := ws.Subscribe("ntf", func(_ json.RawMessage) {
		called.Store(true)
	})

	// Cancel before the server sends the message.
	cancelSub()
	close(ready)

	// Give the read loop time to process the message.
	time.Sleep(100 * time.Millisecond)

	if called.Load() {
		t.Error("handler was called after cancel")
	}
}

func TestWSConn_Close_Idempotent(t *testing.T) {
	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		for {
			_, _, err := conn.Read(context.Background())
			if err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}

	if err := ws.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}

	// Second close should not panic or block.
	if err := ws.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}

	select {
	case <-ws.Done():
	default:
		t.Fatal("Done() should be closed after Close()")
	}
}

func TestWSConn_ServerDisconnect(t *testing.T) {
	serverDone := make(chan struct{})

	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		conn.Close(websocket.StatusNormalClosure, "bye")
		close(serverDone)
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}

	select {
	case <-ws.Done():
		// Read loop exited — expected.
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for Done() after server disconnect")
	}
}

func TestWSConn_Send_AfterClose(t *testing.T) {
	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		for {
			_, _, err := conn.Read(context.Background())
			if err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}

	ws.Close()

	err = ws.Send(ctx, "tic")
	if err == nil {
		t.Fatal("expected error sending on closed connection")
	}
	if !errors.Is(err, transport.ErrWSClosed) {
		t.Errorf("expected ErrWSClosed, got %v", err)
	}
}

func TestWSConn_UnmatchedTopic(t *testing.T) {
	_, wsURL := testWSServer(t, func(conn *websocket.Conn) {
		defer conn.Close(websocket.StatusNormalClosure, "")
		// Send message with no matching handler.
		msg := `{"topic":"unknown","data":"ignored"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg)); err != nil {
			return
		}
		// Then send one with a matching handler.
		msg2 := `{"topic":"ntf","id":"1","text":"found"}`
		if err := conn.Write(context.Background(), websocket.MessageText, []byte(msg2)); err != nil {
			return
		}
		for {
			if _, _, err := conn.Read(context.Background()); err != nil {
				return
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	ws, err := transport.DialWS(ctx, wsURL, http.DefaultClient, testLogger())
	if err != nil {
		t.Fatalf("DialWS: %v", err)
	}
	defer ws.Close()

	received := make(chan struct{}, 1)
	cancelSub, _ := ws.Subscribe("ntf", func(_ json.RawMessage) {
		received <- struct{}{}
	})
	defer cancelSub()

	select {
	case <-received:
		// The "unknown" topic was silently skipped; "ntf" was dispatched.
	case <-time.After(2 * time.Second):
		t.Fatal("timed out — unmatched topic may have blocked dispatch")
	}
}
