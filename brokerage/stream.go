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
	"github.com/trippwill/gbkr/num"
)

// MarketDataUpdate represents a streaming market data tick.
// Field keys are [SnapshotField] constants; values are [FieldValue]
// using the same accessor API as REST /iserver/marketdata/snapshot.
type MarketDataUpdate struct {
	ConID  gbkr.ConID
	Fields map[SnapshotField]FieldValue
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
			Fields: make(map[SnapshotField]FieldValue, len(fields)),
		}

		for key, val := range raw {
			sf := SnapshotField(key)
			if _, ok := requested[sf]; ok {
				update.Fields[sf] = FieldValue{raw: val}
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
// Topic: smh+{conid}.
type HistorySubscriptionBar struct {
	ConID  gbkr.ConID
	Time   string
	Open   num.Num
	High   num.Num
	Low    num.Num
	Close  num.Num
	Volume num.Num
}

// wireHistoryBar is the JSON wire format for streaming historical bars.
type wireHistoryBar struct {
	Time   string  `json:"t"`
	Open   num.Num `json:"o"`
	High   num.Num `json:"h"`
	Low    num.Num `json:"l"`
	Close  num.Num `json:"c"`
	Volume num.Num `json:"v"`
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
		wire := wireHistoryBar{
			Open:   num.Zero(),
			High:   num.Zero(),
			Low:    num.Zero(),
			Close:  num.Zero(),
			Volume: num.Zero(),
		}
		if err := json.Unmarshal(data, &wire); err != nil {
			return
		}
		bar := HistorySubscriptionBar{
			ConID:  conID,
			Time:   wire.Time,
			Open:   wire.Open,
			High:   wire.High,
			Low:    wire.Low,
			Close:  wire.Close,
			Volume: wire.Volume,
		}

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
			// TODO: Validate umh unsub protocol against live gateway.
			// If IBKR requires a server-assigned subscription ID instead
			// of conid for smh, update this to capture it from the ack.
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
	OrderID   gbkr.OrderID
	Account   gbkr.AccountID
	Status    string
	FilledQty num.Num
	AvgPrice  num.Num
	// Additional raw fields for extensibility.
	Raw json.RawMessage `json:"-"`
}

// wireOrderUpdate is the JSON wire format for order status updates.
type wireOrderUpdate struct {
	OrderID   json.RawMessage `json:"orderId"`
	Account   string          `json:"account"`
	Status    string          `json:"status"`
	FilledQty num.Num         `json:"filledQuantity"`
	AvgPrice  num.Num         `json:"avgPrice"`
}

// SubscribeOrders subscribes to live order status updates.
// Topic: sor. Uses blocking channel semantics (loss-intolerant).
func SubscribeOrders(s *gbkr.Stream) (updates <-chan OrderUpdate, cancel func(), err error) {
	return brokerageSubscribe(s, "sor", "sor+{}", 32,
		func(data json.RawMessage) (OrderUpdate, error) {
			wire := wireOrderUpdate{
				FilledQty: num.Zero(),
				AvgPrice:  num.Zero(),
			}
			if err := json.Unmarshal(data, &wire); err != nil {
				return OrderUpdate{}, err
			}
			return OrderUpdate{
				OrderID:   gbkr.OrderID(normalizeJSONString(wire.OrderID)),
				Account:   gbkr.AccountID(wire.Account),
				Status:    wire.Status,
				FilledQty: wire.FilledQty,
				AvgPrice:  wire.AvgPrice,
				Raw:       data,
			}, nil
		})
}

// TradeUpdate represents a streaming trade execution update. Topic: str.
type TradeUpdate struct {
	ExecutionID string
	ConID       gbkr.ConID
	Symbol      string
	Side        string
	Quantity    num.Num
	Price       num.Num
	// Additional raw fields for extensibility.
	Raw json.RawMessage `json:"-"`
}

// wireTradeUpdate is the JSON wire format for trade execution updates.
type wireTradeUpdate struct {
	ExecutionID string  `json:"execution_id"`
	ConID       int     `json:"conid"`
	Symbol      string  `json:"symbol"`
	Side        string  `json:"side"`
	Quantity    num.Num `json:"size"`
	Price       num.Num `json:"price"`
}

// SubscribeTrades subscribes to live trade execution updates.
// Topic: str. Uses blocking channel semantics (loss-intolerant).
func SubscribeTrades(s *gbkr.Stream) (updates <-chan TradeUpdate, cancel func(), err error) {
	return brokerageSubscribe(s, "str", "str+{}", 32,
		func(data json.RawMessage) (TradeUpdate, error) {
			wire := wireTradeUpdate{
				Quantity: num.Zero(),
				Price:    num.Zero(),
			}
			if err := json.Unmarshal(data, &wire); err != nil {
				return TradeUpdate{}, err
			}
			return TradeUpdate{
				ExecutionID: wire.ExecutionID,
				ConID:       gbkr.ConID(wire.ConID),
				Symbol:      wire.Symbol,
				Side:        wire.Side,
				Quantity:    wire.Quantity,
				Price:       wire.Price,
				Raw:         data,
			}, nil
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

// normalizeJSONString extracts a string from a json.RawMessage that may
// be either a JSON string ("123") or a JSON number (123). Returns the
// unquoted string value in both cases.
func normalizeJSONString(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return s
	}
	// Not a JSON string — return the raw text (handles numeric values).
	return strings.TrimSpace(string(raw))
}
