package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/trippwill/gbkr/internal/transport"
)

const keepaliveInterval = 25 * time.Second

// Stream manages a WebSocket connection to the IBKR Client Portal Gateway
// for real-time push updates. Obtained via [Client.Stream].
type Stream struct {
	ws        *transport.WSConn
	client    *Client
	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
	closed    bool
}

// Stream opens a WebSocket connection to the IBKR gateway and returns a
// [Stream] for subscribing to real-time topics. The connection is closed
// when ctx is cancelled or [Stream.Close] is called.
func (c *Client) Stream(ctx context.Context) (*Stream, error) {
	wsURL, err := deriveWSURL(c.t.BaseURL)
	if err != nil {
		return nil, err
	}

	ws, err := transport.DialWS(ctx, wsURL, c.t.HTTPClient, c.t.Logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	s := &Stream{
		ws:     ws,
		client: c,
		cancel: cancel,
		done:   make(chan struct{}),
	}

	start := time.Now()
	c.emitOp(ctx, OpStreamConnect, nil, time.Since(start))

	go s.keepalive(ctx)
	go s.watchClose(ctx)

	return s, nil
}

// Close closes the stream and its underlying WebSocket connection.
// All subscription channels are closed. Idempotent.
func (s *Stream) Close() error {
	var closeErr error
	s.closeOnce.Do(func() {
		s.closed = true
		s.cancel()
		closeErr = s.ws.Close()
		start := time.Now()
		s.client.emitOp(context.Background(), OpStreamDisconnect, nil, time.Since(start))
	})
	<-s.done
	return closeErr
}

// wsConn returns the underlying WSConn for cross-package subscription.
// Accessible within the module but unexported from the public API.
func (s *Stream) WsConn() *transport.WSConn { return s.ws }

func (s *Stream) keepalive(ctx context.Context) {
	ticker := time.NewTicker(keepaliveInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.ws.Done():
			return
		case <-ticker.C:
			_ = s.ws.Send(ctx, "tic")
		}
	}
}

// watchClose waits for either context cancellation or WebSocket disconnect,
// then ensures cleanup and signals done.
func (s *Stream) watchClose(ctx context.Context) {
	defer close(s.done)
	select {
	case <-ctx.Done():
		s.ws.Close()
	case <-s.ws.Done():
		s.cancel()
	}
}

// deriveWSURL converts a REST base URL to a WebSocket URL.
//
//	https://host:port/v1/api → wss://host:port/v1/api/ws
//	http://host:port/v1/api  → ws://host:port/v1/api/ws
func deriveWSURL(baseURL string) (string, error) {
	switch {
	case strings.HasPrefix(baseURL, "https://"):
		return "wss://" + strings.TrimPrefix(baseURL, "https://") + "/ws", nil
	case strings.HasPrefix(baseURL, "http://"):
		return "ws://" + strings.TrimPrefix(baseURL, "http://") + "/ws", nil
	default:
		return "", fmt.Errorf("unsupported scheme in base URL: %s", baseURL)
	}
}

// Notification represents a gateway notification message.
type Notification struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Text string `json:"text"`
}

// Notifications subscribes to gateway notification messages and returns a
// channel that receives them. The channel is closed when the stream closes.
// Multiple calls return independent channels (fan-out).
func (s *Stream) Notifications() (<-chan Notification, error) {
	if s.closed {
		return nil, ErrStreamNotConnected
	}

	ch := make(chan Notification, 32)

	cancelSub := s.ws.Subscribe("ntf", func(data json.RawMessage) {
		var n Notification
		if err := json.Unmarshal(data, &n); err != nil {
			s.client.t.Logger.Warn("ws: malformed ntf frame",
				slog.String("error", err.Error()))
			return
		}
		select {
		case ch <- n:
		default:
			// Buffer full — drop oldest to make room.
			select {
			case <-ch:
			default:
			}
			ch <- n
		}
	})

	s.client.emitOp(context.Background(), OpStreamSubscribe, nil, 0,
		slog.String("topic", "ntf"))

	// Close channel when the stream disconnects.
	go func() {
		<-s.ws.Done()
		cancelSub()
		close(ch)
	}()

	return ch, nil
}
