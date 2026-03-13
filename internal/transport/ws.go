package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
)

// topicHeader is used for lightweight topic extraction from incoming frames.
type topicHeader struct {
	Topic string `json:"topic"`
}

// WSConn wraps a WebSocket connection with topic-based message dispatch.
type WSConn struct {
	conn      *websocket.Conn
	writeMu   sync.Mutex
	handlers  sync.Map // topic string → *handlerList
	done      chan struct{}
	closeOnce sync.Once
	cancelCtx context.CancelFunc // cancels the readLoop context
	logger    *slog.Logger
}

// handlerList holds fan-out handlers for a single topic.
type handlerList struct {
	mu       sync.Mutex
	handlers map[int]*handler
	nextID   int
}

type handler struct {
	fn     func(json.RawMessage)
	active *atomic.Bool // set to false before channel close to prevent send-after-close
}

func newHandlerList() *handlerList {
	return &handlerList{handlers: make(map[int]*handler)}
}

func (hl *handlerList) add(fn func(json.RawMessage)) (int, *atomic.Bool) {
	active := &atomic.Bool{}
	active.Store(true)
	hl.mu.Lock()
	defer hl.mu.Unlock()
	id := hl.nextID
	hl.nextID++
	hl.handlers[id] = &handler{fn: fn, active: active}
	return id, active
}

func (hl *handlerList) remove(id int) bool {
	hl.mu.Lock()
	defer hl.mu.Unlock()
	delete(hl.handlers, id)
	return len(hl.handlers) == 0
}

func (hl *handlerList) dispatch(data json.RawMessage) {
	hl.mu.Lock()
	// Snapshot handlers under lock to avoid holding it during callbacks.
	type entry struct {
		fn     func(json.RawMessage)
		active *atomic.Bool
	}
	entries := make([]entry, 0, len(hl.handlers))
	for _, h := range hl.handlers {
		entries = append(entries, entry{fn: h.fn, active: h.active})
	}
	hl.mu.Unlock()

	for _, e := range entries {
		if e.active.Load() {
			e.fn(data)
		}
	}
}

// DialWS establishes a WebSocket connection and starts the read loop.
// The provided ctx is used only for dialing; the read loop runs on an
// internal context that is cancelled when Close is called.
func DialWS(ctx context.Context, wsURL string, httpClient *http.Client, logger *slog.Logger) (*WSConn, error) {
	opts := &websocket.DialOptions{
		HTTPClient: httpClient,
	}

	conn, _, err := websocket.Dial(ctx, wsURL, opts)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrWSClosed, err)
	}

	// Derive a non-cancelling context from the dial context so the read loop
	// is not tied to the caller's deadline/cancellation.
	readCtx, cancelRead := context.WithCancel(context.WithoutCancel(ctx))

	ws := &WSConn{
		conn:      conn,
		done:      make(chan struct{}),
		cancelCtx: cancelRead,
		logger:    logger,
	}

	go ws.readLoop(readCtx)

	return ws, nil
}

// Send writes a text message to the WebSocket connection.
// Thread-safe; multiple goroutines may call Send concurrently.
func (ws *WSConn) Send(ctx context.Context, msg string) error {
	ws.writeMu.Lock()
	defer ws.writeMu.Unlock()

	select {
	case <-ws.done:
		return ErrWSClosed
	default:
	}

	if err := ws.conn.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
		return fmt.Errorf("%w: %w", ErrWSSend, err)
	}
	return nil
}

// Subscribe registers a handler for messages matching the given topic.
// Returns a cancel function that deregisters the handler, and an active
// flag that is set to false before deregistration. Callers that close
// channels in their cancel path should call active.Store(false) first
// to prevent in-flight dispatch from sending on a closed channel.
// Multiple handlers may be registered for the same topic (fan-out).
func (ws *WSConn) Subscribe(topic string, fn func(json.RawMessage)) (cancel func(), active *atomic.Bool) {
	actual, _ := ws.handlers.LoadOrStore(topic, newHandlerList())
	hl := actual.(*handlerList)
	id, active := hl.add(fn)

	return func() {
		if empty := hl.remove(id); empty {
			ws.handlers.Delete(topic)
		}
	}, active
}

// Done returns a channel that is closed when the WebSocket connection is lost.
func (ws *WSConn) Done() <-chan struct{} {
	return ws.done
}

// Close closes the WebSocket connection and waits for the read loop to exit.
// Idempotent — safe to call multiple times.
func (ws *WSConn) Close() error {
	var closeErr error
	ws.closeOnce.Do(func() {
		ws.cancelCtx()
		closeErr = ws.conn.Close(websocket.StatusNormalClosure, "")
	})
	<-ws.done
	return closeErr
}

func (ws *WSConn) readLoop(ctx context.Context) {
	defer close(ws.done)
	for {
		_, data, err := ws.conn.Read(ctx)
		if err != nil {
			return
		}

		var hdr topicHeader
		if err := json.Unmarshal(data, &hdr); err != nil {
			ws.logger.Warn("ws: malformed frame", slog.String("error", err.Error()))
			continue
		}
		if hdr.Topic == "" {
			continue
		}

		if val, ok := ws.handlers.Load(hdr.Topic); ok {
			val.(*handlerList).dispatch(json.RawMessage(data))
		}
	}
}
