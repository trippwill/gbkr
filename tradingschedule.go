package gbkr

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
)

// TradingSchedule is the response for GET /contract/trading-schedule.
// It contains the exchange timezone and per-date schedule entries.
type TradingSchedule struct {
	// ExchangeTimezone is the IANA timezone of the exchange (e.g., "US/Eastern").
	ExchangeTimezone string
	// Schedules maps date strings ("YYYYMMDD") to their trading sessions.
	// Dates absent from this map are non-trading days (holidays, weekends).
	Schedules map[string]TradingDay
}

// TradingDay holds the trading hours for a single date.
type TradingDay struct {
	// LiquidHours are the regular (core) trading sessions.
	LiquidHours []TradingSession
	// ExtendedHours are the full trading sessions including pre/post market.
	ExtendedHours []TradingSession
}

// TradingSession represents a contiguous block of trading time.
type TradingSession struct {
	// Opening is the session start as a Unix epoch timestamp (seconds).
	Opening int64
	// Closing is the session end as a Unix epoch timestamp (seconds).
	Closing int64
	// CancelDayOrders indicates whether day orders are canceled at session close.
	CancelDayOrders bool
}

func (ts *TradingSchedule) UnmarshalJSON(data []byte) error {
	var aux struct {
		ExchangeTimezone *string               `json:"exchange_time_zone"`
		Schedules        map[string]TradingDay `json:"schedules"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	ts.ExchangeTimezone = jx.Deref(aux.ExchangeTimezone)
	ts.Schedules = aux.Schedules
	return nil
}

func (td *TradingDay) UnmarshalJSON(data []byte) error {
	var aux struct {
		LiquidHours   []TradingSession `json:"liquid_hours"`
		ExtendedHours []TradingSession `json:"extended_hours"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	td.LiquidHours = aux.LiquidHours
	td.ExtendedHours = aux.ExtendedHours
	return nil
}

func (s *TradingSession) UnmarshalJSON(data []byte) error {
	var raw struct {
		Opening         *int64 `json:"opening,omitempty"`
		Closing         *int64 `json:"closing,omitempty"`
		CancelDayOrders *bool  `json:"cancel_daily_orders,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.Opening = jx.Deref(raw.Opening)
	s.Closing = jx.Deref(raw.Closing)
	s.CancelDayOrders = jx.Deref(raw.CancelDayOrders)
	return nil
}

// TradingSchedule retrieves the trading schedule for a contract
// (GET /contract/trading-schedule).
// Returns schedule data for approximately 6 days surrounding the current
// trading day. Non-trading days (holidays, weekends) are absent from the
// response. This endpoint does not require a brokerage session.
func (c *Client) TradingSchedule(ctx context.Context, conID ConID, exchange Exchange) (*TradingSchedule, error) {
	start := time.Now()
	var result TradingSchedule
	q := url.Values{"conid": {conID.String()}}
	if exchange != "" {
		q.Set("exchange", exchange.String())
	}
	err := c.doGet(ctx, "/contract/trading-schedule", q, &result)
	c.emitOp(ctx, OpTradingSchedule, err, time.Since(start),
		slog.Int64("conid", int64(conID)),
		slog.String("exchange", string(exchange)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
