package brokerage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/trippwill/gbkr"
	"github.com/trippwill/gbkr/internal/transport"
)

// MarketDataUpdate represents a streaming market data tick.
// Field keys are [SnapshotField] constants; values are raw JSON
// matching the REST /iserver/marketdata/snapshot field encoding.
type MarketDataUpdate struct {
	ConID  gbkr.ConID
	Fields map[SnapshotField]json.RawMessage
}

// sendTimeout is the maximum time allowed for a subscribe/unsubscribe send.
const sendTimeout = 5 * time.Second

// SubscribeMarketData subscribes to streaming market data for the given
// contract. Only the requested fields are included in updates (when the
// gateway provides them). Topic: smd+{conid}.
//
// The stream must have been obtained from a [gbkr.Client] that has an
// active brokerage session.
func SubscribeMarketData(s *gbkr.Stream, conID gbkr.ConID, fields ...SnapshotField) (updates <-chan MarketDataUpdate, cancel func(), err error) {
	if s.IsClosed() {
		return nil, nil, gbkr.ErrStreamNotConnected
	}

	fieldStrs := make([]string, len(fields))
	for i, f := range fields {
		fieldStrs[i] = `"` + string(f) + `"`
	}
	topic := fmt.Sprintf("smd+%d", conID)
	subCmd := fmt.Sprintf(`smd+%d+{"fields":[%s]}`, conID, strings.Join(fieldStrs, ","))

	ws := s.WsConn()
	ch := make(chan MarketDataUpdate, 64)

	requested := make(map[SnapshotField]struct{}, len(fields))
	for _, f := range fields {
		requested[f] = struct{}{}
	}

	var (
		cancelSub func()
		active    *atomic.Bool
	)
	cancelSub, active = ws.Subscribe(topic, func(data json.RawMessage) {
		if !active.Load() {
			return
		}
		if obs := s.Observer(); obs != nil {
			obs.OnMessageReceived(topic)
		}
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(data, &raw); err != nil {
			return
		}

		update := MarketDataUpdate{
			ConID:  conID,
			Fields: make(map[SnapshotField]json.RawMessage, len(fields)),
		}

		for key, val := range raw {
			sf := SnapshotField(key)
			if _, ok := requested[sf]; ok {
				update.Fields[sf] = val
			}
		}

		if len(update.Fields) == 0 {
			return
		}

		// Drop oldest on overflow (market data is lossy-ok).
		select {
		case ch <- update:
		default:
			select {
			case <-ch:
			default:
			}
			ch <- update
		}
	})

	sendCtx, sendCancel := context.WithTimeout(context.Background(), sendTimeout)
	defer sendCancel()
	if err := ws.Send(sendCtx, subCmd); err != nil {
		active.Store(false)
		cancelSub()
		close(ch)
		return nil, nil, fmt.Errorf("subscribe %s: %w", topic, err)
	}
	s.EmitOp(gbkr.OpStreamSubscribe,
		slog.String("topic", topic),
		slog.Int("conid", int(conID)))
	if obs := s.Observer(); obs != nil {
		obs.OnSubscribe(topic)
	}

	var cancelOnce sync.Once
	cancelFn := func() {
		cancelOnce.Do(func() {
			active.Store(false)
			cancelSub()
			unsub := fmt.Sprintf("umd+%d+{}", conID)
			unsubCtx, unsubCancel := context.WithTimeout(context.Background(), sendTimeout)
			defer unsubCancel()
			_ = ws.Send(unsubCtx, unsub)
			s.EmitOp(gbkr.OpStreamUnsubscribe,
				slog.String("topic", topic),
				slog.Int("conid", int(conID)))
			if obs := s.Observer(); obs != nil {
				obs.OnUnsubscribe(topic)
			}
			close(ch)
		})
	}

	go func() {
		<-ws.Done()
		cancelFn()
	}()

	return ch, cancelFn, nil
}

// HistorySubscriptionBar represents a streaming historical bar update.
// Topic: smh+{conid}. Numeric fields use [json.RawMessage] to preserve
// the gateway's flexible encoding (string or number).
type HistorySubscriptionBar struct {
	ConID  gbkr.ConID      `json:"-"`
	Time   string          `json:"t"`
	Open   json.RawMessage `json:"o"`
	High   json.RawMessage `json:"h"`
	Low    json.RawMessage `json:"l"`
	Close  json.RawMessage `json:"c"`
	Volume json.RawMessage `json:"v"`
}

// SubscribeMarketDataHistory subscribes to streaming historical bar updates
// for the given contract. Topic: smh+{conid}.
//
// Uses drop-oldest semantics with a 64-element buffer (lossy-ok).
func SubscribeMarketDataHistory(s *gbkr.Stream, conID gbkr.ConID) (updates <-chan HistorySubscriptionBar, cancel func(), err error) {
	if s.IsClosed() {
		return nil, nil, gbkr.ErrStreamNotConnected
	}

	topic := fmt.Sprintf("smh+%d", conID)
	subCmd := topic

	ws := s.WsConn()
	ch := make(chan HistorySubscriptionBar, 64)

	var (
		cancelSub func()
		active    *atomic.Bool
	)
	cancelSub, active = ws.Subscribe(topic, func(data json.RawMessage) {
		if !active.Load() {
			return
		}
		var bar HistorySubscriptionBar
		if err := json.Unmarshal(data, &bar); err != nil {
			return
		}
		bar.ConID = conID

		if obs := s.Observer(); obs != nil {
			obs.OnMessageReceived(topic)
		}

		// Drop oldest on overflow (historical bars are lossy-ok).
		select {
		case ch <- bar:
		default:
			select {
			case <-ch:
			default:
			}
			ch <- bar
		}
	})

	sendCtx, sendCancel := context.WithTimeout(context.Background(), sendTimeout)
	defer sendCancel()
	if err := ws.Send(sendCtx, subCmd); err != nil {
		active.Store(false)
		cancelSub()
		close(ch)
		return nil, nil, fmt.Errorf("subscribe %s: %w", topic, err)
	}
	s.EmitOp(gbkr.OpStreamSubscribe,
		slog.String("topic", topic),
		slog.Int("conid", int(conID)))
	if obs := s.Observer(); obs != nil {
		obs.OnSubscribe(topic)
	}

	var cancelOnce sync.Once
	cancelFn := func() {
		cancelOnce.Do(func() {
			active.Store(false)
			cancelSub()
			unsub := fmt.Sprintf("umh+%d+{}", conID)
			unsubCtx, unsubCancel := context.WithTimeout(context.Background(), sendTimeout)
			defer unsubCancel()
			_ = ws.Send(unsubCtx, unsub)
			s.EmitOp(gbkr.OpStreamUnsubscribe,
				slog.String("topic", topic),
				slog.Int("conid", int(conID)))
			if obs := s.Observer(); obs != nil {
				obs.OnUnsubscribe(topic)
			}
			close(ch)
		})
	}

	go func() {
		<-ws.Done()
		cancelFn()
	}()

	return ch, cancelFn, nil
}

// OrderUpdate represents a streaming order status update. Topic: sor.
type OrderUpdate struct {
	OrderID   json.RawMessage `json:"orderId"`
	Account   string          `json:"account"`
	Status    string          `json:"status"`
	FilledQty json.RawMessage `json:"filledQuantity"`
	AvgPrice  json.RawMessage `json:"avgPrice"`
	// Additional raw fields for extensibility.
	Raw json.RawMessage `json:"-"`
}

// SubscribeOrders subscribes to live order status updates.
// Topic: sor. Uses blocking channel semantics (loss-intolerant).
func SubscribeOrders(s *gbkr.Stream) (updates <-chan OrderUpdate, cancel func(), err error) {
	return brokerageSubscribe(s, "sor", "sor+{}", 32,
		func(data json.RawMessage) (OrderUpdate, error) {
			var u OrderUpdate
			if err := json.Unmarshal(data, &u); err != nil {
				return OrderUpdate{}, err
			}
			u.Raw = data
			return u, nil
		})
}

// TradeUpdate represents a streaming trade execution update. Topic: str.
type TradeUpdate struct {
	ExecutionID string          `json:"execution_id"`
	ConID       json.RawMessage `json:"conid"`
	Symbol      string          `json:"symbol"`
	Side        string          `json:"side"`
	Quantity    json.RawMessage `json:"size"`
	Price       json.RawMessage `json:"price"`
	// Additional raw fields for extensibility.
	Raw json.RawMessage `json:"-"`
}

// SubscribeTrades subscribes to live trade execution updates.
// Topic: str. Uses blocking channel semantics (loss-intolerant).
func SubscribeTrades(s *gbkr.Stream) (updates <-chan TradeUpdate, cancel func(), err error) {
	return brokerageSubscribe(s, "str", "str+{}", 32,
		func(data json.RawMessage) (TradeUpdate, error) {
			var u TradeUpdate
			if err := json.Unmarshal(data, &u); err != nil {
				return TradeUpdate{}, err
			}
			u.Raw = data
			return u, nil
		})
}

// brokerageSubscribe is a helper for brokerage-level topic subscriptions.
// Uses blocking channel semantics (loss-intolerant).
func brokerageSubscribe[T any](s *gbkr.Stream, topic, subCmd string, bufSize int, parse func(json.RawMessage) (T, error)) (<-chan T, func(), error) {
	if s.IsClosed() {
		return nil, nil, gbkr.ErrStreamNotConnected
	}

	ws := s.WsConn()
	ch := make(chan T, bufSize)

	var (
		cancelSub func()
		active    *atomic.Bool
	)
	cancelSub, active = ws.Subscribe(topic, func(data json.RawMessage) {
		v, err := parse(data)
		if err != nil {
			return
		}
		if !active.Load() {
			return
		}
		if obs := s.Observer(); obs != nil {
			obs.OnMessageReceived(topic)
		}
		ch <- v
	})

	sendCtx, sendCancel := context.WithTimeout(context.Background(), sendTimeout)
	defer sendCancel()
	if err := ws.Send(sendCtx, subCmd); err != nil {
		active.Store(false)
		cancelSub()
		close(ch)
		return nil, nil, fmt.Errorf("subscribe %s: %w", topic, err)
	}
	s.EmitOp(gbkr.OpStreamSubscribe, slog.String("topic", topic))
	if obs := s.Observer(); obs != nil {
		obs.OnSubscribe(topic)
	}

	var cancelOnce sync.Once
	cancelFn := func() {
		cancelOnce.Do(func() {
			active.Store(false)
			cancelSub()
			if len(topic) > 1 {
				unsub := "u" + topic[1:]
				unsubCtx, unsubCancel := context.WithTimeout(context.Background(), sendTimeout)
				defer unsubCancel()
				_ = ws.Send(unsubCtx, unsub)
			}
			s.EmitOp(gbkr.OpStreamUnsubscribe, slog.String("topic", topic))
			if obs := s.Observer(); obs != nil {
				obs.OnUnsubscribe(topic)
			}
			close(ch)
		})
	}

	go watchDone(ws, cancelFn)

	return ch, cancelFn, nil
}

func watchDone(ws *transport.WSConn, cancelFn func()) {
	<-ws.Done()
	cancelFn()
}
