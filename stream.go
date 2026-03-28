package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trippwill/gbkr/internal/transport"
)

const keepaliveInterval = 25 * time.Second

// StreamObserver receives streaming lifecycle and activity notifications.
// Implementations must be safe for concurrent use.
type StreamObserver interface {
	OnMessageReceived(topic string)
	OnSubscribe(topic string)
	OnUnsubscribe(topic string)
	OnKeepaliveSent(err error)
}

// Stream manages a WebSocket connection to the IBKR Client Portal Gateway
// for real-time push updates. Obtained via [Client.Stream].
//
// Stream does not auto-reconnect. When the underlying WebSocket disconnects,
// all subscription channels are closed. Consumers should detect channel
// closure and re-dial via [Client.Stream] with appropriate backoff.
// See the package-level documentation for a reconnection example.
type Stream struct {
	ws        *transport.WSConn
	client    *Client
	observer  StreamObserver
	cancel    context.CancelFunc
	done      chan struct{}
	closeOnce sync.Once
	closed    atomic.Bool
}

// Stream opens a WebSocket connection to the IBKR gateway and returns a
// [Stream] for subscribing to real-time topics. The connection is closed
// when ctx is cancelled or [Stream.Close] is called.
func (c *Client) Stream(ctx context.Context) (*Stream, error) {
	wsURL, err := deriveWSURL(c.t.BaseURL)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	ws, err := transport.DialWS(ctx, wsURL, c.t.HTTPClient, c.t.Logger)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(ctx)

	s := &Stream{
		ws:       ws,
		client:   c,
		observer: c.streamObserver,
		cancel:   cancel,
		done:     make(chan struct{}),
	}

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
		s.closed.Store(true)
		s.cancel()
		start := time.Now()
		closeErr = s.ws.Close()
		s.client.emitOp(context.Background(), OpStreamDisconnect, nil, time.Since(start))
	})
	<-s.done
	return closeErr
}

// WsConn returns the underlying [transport.WSConn] for cross-package subscription.
// The return type is from internal/transport, so external consumers cannot
// reference it — this method is effectively module-internal.
func (s *Stream) WsConn() *transport.WSConn { return s.ws }

// EmitOp exposes operation logging for cross-package subscription helpers
// in the brokerage package.
func (s *Stream) EmitOp(op Operation, attrs ...slog.Attr) {
	s.client.emitOp(context.Background(), op, nil, 0, attrs...)
}

// Observer returns the stream's [StreamObserver], or nil if none was configured.
// Exposed for cross-package subscription helpers in the brokerage package.
func (s *Stream) Observer() StreamObserver { return s.observer }

// IsClosed reports whether the stream has been closed explicitly or
// the underlying WebSocket has disconnected.
func (s *Stream) IsClosed() bool {
	if s.closed.Load() {
		return true
	}
	select {
	case <-s.ws.Done():
		return true
	default:
		return false
	}
}

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
			err := s.ws.Send(ctx, "tic")
			if s.observer != nil {
				s.observer.OnKeepaliveSent(err)
			}
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

// sendTimeout is the maximum time allowed for a subscribe/unsubscribe send.
const sendTimeout = 5 * time.Second

// subscribe is the common subscription helper. It registers a topic handler,
// sends the subscribe command, emits the operation, and returns a cancel
// function that unsubscribes and closes the channel.
func subscribe[T any](s *Stream, topic string, bufSize int, dropOnFull bool, parse func(json.RawMessage) (T, error), attrs ...slog.Attr) (<-chan T, func(), error) {
	if s.IsClosed() {
		return nil, nil, ErrStreamNotConnected
	}

	ch := make(chan T, bufSize)

	var (
		cancelSub func()
		active    *atomic.Bool
	)
	cancelSub, active = s.ws.Subscribe(topic, func(data json.RawMessage) {
		v, err := parse(data)
		if err != nil {
			s.client.t.Logger.Warn("ws: malformed frame",
				slog.String("topic", topic),
				slog.String("error", err.Error()))
			return
		}
		if !active.Load() {
			return
		}
		if s.observer != nil {
			s.observer.OnMessageReceived(topic)
		}
		if dropOnFull {
			select {
			case ch <- v:
			default:
				// Drop oldest to make room.
				select {
				case <-ch:
				default:
				}
				ch <- v
			}
		} else {
			ch <- v
		}
	})

	// Send subscribe command to gateway.
	sendCtx, sendCancel := context.WithTimeout(context.Background(), sendTimeout)
	defer sendCancel()
	if err := s.ws.Send(sendCtx, topic); err != nil {
		active.Store(false)
		cancelSub()
		close(ch)
		return nil, nil, fmt.Errorf("subscribe %s: %w", topic, err)
	}

	allAttrs := make([]slog.Attr, 0, 1+len(attrs))
	allAttrs = append(allAttrs, slog.String("topic", topic))
	allAttrs = append(allAttrs, attrs...)
	s.client.emitOp(context.Background(), OpStreamSubscribe, nil, 0, allAttrs...)
	if s.observer != nil {
		s.observer.OnSubscribe(topic)
	}

	var cancelOnce sync.Once
	cancelFn := func() {
		cancelOnce.Do(func() {
			active.Store(false)
			cancelSub()
			// Send unsubscribe command using the "u" prefix convention.
			if len(topic) > 1 {
				unsub := "u" + topic[1:]
				unsubCtx, unsubCancel := context.WithTimeout(context.Background(), sendTimeout)
				defer unsubCancel()
				_ = s.ws.Send(unsubCtx, unsub)
			}
			s.client.emitOp(context.Background(), OpStreamUnsubscribe, nil, 0,
				slog.String("topic", topic))
			if s.observer != nil {
				s.observer.OnUnsubscribe(topic)
			}
			close(ch)
		})
	}

	// Auto-close on stream disconnect.
	go func() {
		<-s.ws.Done()
		cancelFn()
	}()

	return ch, cancelFn, nil
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

// ---------------------------------------------------------------------------
// Gateway-level subscriptions
// ---------------------------------------------------------------------------

// Notification represents a gateway notification message.
type Notification struct {
	ID   string `json:"id"`
	Date string `json:"date"`
	Text string `json:"text"`
}

// Notifications subscribes to gateway notification messages. The returned
// channel receives notifications; call cancel to unsubscribe and close it.
// Multiple calls return independent channels (fan-out).
func (s *Stream) Notifications() (updates <-chan Notification, cancel func(), err error) {
	return subscribe(s, "ntf", 32, true,
		func(data json.RawMessage) (Notification, error) {
			var n Notification
			return n, json.Unmarshal(data, &n)
		})
}

// AccountSummaryUpdate represents a streaming account summary update
// from the gateway topic sbd+{acctId}.
type AccountSummaryUpdate struct {
	AccountID AccountID                  `json:"-"`
	Fields    map[string]json.RawMessage `json:"-"`
}

// AccountSummary subscribes to account summary updates for the given account.
// Gateway topic: sbd+{acctId}.
func (s *Stream) AccountSummary(accountID AccountID) (updates <-chan AccountSummaryUpdate, cancel func(), err error) {
	topic := "sbd+" + string(accountID)
	return subscribe(s, topic, 16, true,
		func(data json.RawMessage) (AccountSummaryUpdate, error) {
			var raw map[string]json.RawMessage
			if err := json.Unmarshal(data, &raw); err != nil {
				return AccountSummaryUpdate{}, err
			}
			// Remove metadata fields.
			delete(raw, "topic")
			return AccountSummaryUpdate{AccountID: accountID, Fields: raw}, nil
		},
		slog.String("account_id", string(accountID)))
}

// Get returns the raw JSON value for a field key, or nil if absent.
func (u *AccountSummaryUpdate) Get(key string) json.RawMessage {
	return u.Fields[key]
}

// PnLUpdate represents a streaming portfolio P&L update
// from the gateway topic spl+{acctId}.
type PnLUpdate struct {
	AccountID     AccountID `json:"-"`
	DailyPnL      float64   `json:"dpl"`
	NetLiquidity  float64   `json:"nl"`
	UnrealizedPnL float64   `json:"upl"`
	RealizedPnL   float64   `json:"rpl"`
	ExcessLiq     float64   `json:"el"`
	MarginValue   float64   `json:"mv"`
}

// PortfolioPnL subscribes to portfolio P&L updates for the given account.
// Gateway topic: spl+{acctId}.
func (s *Stream) PortfolioPnL(accountID AccountID) (updates <-chan PnLUpdate, cancel func(), err error) {
	topic := "spl+" + string(accountID)
	return subscribe(s, topic, 16, true,
		func(data json.RawMessage) (PnLUpdate, error) {
			var u PnLUpdate
			if err := json.Unmarshal(data, &u); err != nil {
				return PnLUpdate{}, err
			}
			u.AccountID = accountID
			return u, nil
		},
		slog.String("account_id", string(accountID)))
}
