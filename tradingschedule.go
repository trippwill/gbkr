package gbkr

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"time"

	"github.com/trippwill/gbkr/internal/jx"
	"github.com/trippwill/gbkr/when"
)

// TradingSchedule is the response for GET /contract/trading-schedule.
// It contains the exchange timezone and per-date schedule entries.
type TradingSchedule struct {
	// ExchangeTimezone is the IANA timezone of the exchange (e.g., "US/Eastern").
	ExchangeTimezone string
	// Schedules maps dates to their trading sessions.
	// Dates absent from this map are non-trading days (holidays, weekends).
	Schedules map[when.Date]TradingDay
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
	// Opening is the session start time.
	Opening when.DateTime
	// Closing is the session end time.
	Closing when.DateTime
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
	if len(aux.Schedules) > 0 {
		ts.Schedules = make(map[when.Date]TradingDay, len(aux.Schedules))
		for k, v := range aux.Schedules {
			d, err := when.ParseDate(k)
			if err != nil {
				return fmt.Errorf("schedule date key %q: %w", k, err)
			}
			ts.Schedules[d] = v
		}
	}
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
	if raw.Opening != nil {
		s.Opening = when.DateTimeFromEpoch(*raw.Opening)
	}
	if raw.Closing != nil {
		s.Closing = when.DateTimeFromEpoch(*raw.Closing)
	}
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
