package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/trippwill/gbkr/when"
)

func TestTradingSchedule(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/contract/trading-schedule" {
				t.Errorf("path = %q, want /contract/trading-schedule", r.URL.Path)
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			if r.URL.Query().Get("conid") != "265598" {
				t.Errorf("conid = %q, want 265598", r.URL.Query().Get("conid"))
			}
			if r.URL.Query().Get("exchange") != "ISLAND" {
				t.Errorf("exchange = %q, want ISLAND", r.URL.Query().Get("exchange"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"exchange_time_zone": "US/Eastern",
				"schedules": map[string]any{
					"20260303": map[string]any{
						"liquid_hours": []map[string]any{
							{"opening": 1766068200, "closing": 1766095200},
						},
						"extended_hours": []map[string]any{
							{"opening": 1766012400, "closing": 1766095200, "cancel_daily_orders": true},
						},
					},
					"20260304": map[string]any{
						"liquid_hours": []map[string]any{
							{"opening": 1766154600, "closing": 1766181600},
						},
						"extended_hours": []map[string]any{
							{"opening": 1766098800, "closing": 1766181600, "cancel_daily_orders": true},
						},
					},
				},
			})
		})
		defer srv.Close()

		result, err := c.TradingSchedule(context.Background(), 265598, "ISLAND")
		if err != nil {
			t.Fatal(err)
		}
		if result.ExchangeTimezone != "US/Eastern" {
			t.Errorf("timezone = %q, want US/Eastern", result.ExchangeTimezone)
		}
		if len(result.Schedules) != 2 {
			t.Fatalf("schedules count = %d, want 2", len(result.Schedules))
		}
		day, ok := result.Schedules[when.NewDate(2026, 3, 3)]
		if !ok {
			t.Fatal("missing schedule for 2026-03-03")
		}
		if len(day.LiquidHours) != 1 {
			t.Fatalf("liquid_hours count = %d, want 1", len(day.LiquidHours))
		}
		if !day.LiquidHours[0].Opening.Equal(when.DateTimeFromEpoch(1766068200)) {
			t.Errorf("liquid opening = %v, want epoch 1766068200", day.LiquidHours[0].Opening)
		}
		if !day.LiquidHours[0].Closing.Equal(when.DateTimeFromEpoch(1766095200)) {
			t.Errorf("liquid closing = %v, want epoch 1766095200", day.LiquidHours[0].Closing)
		}
		if len(day.ExtendedHours) != 1 {
			t.Fatalf("extended_hours count = %d, want 1", len(day.ExtendedHours))
		}
		if !day.ExtendedHours[0].CancelDayOrders {
			t.Error("expected CancelDayOrders=true")
		}
	})

	t.Run("no_exchange", func(t *testing.T) {
		c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("exchange") != "" {
				t.Errorf("exchange should be empty, got %q", r.URL.Query().Get("exchange"))
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"exchange_time_zone": "US/Central",
				"schedules":          map[string]any{},
			})
		})
		defer srv.Close()

		result, err := c.TradingSchedule(context.Background(), 265598, "")
		if err != nil {
			t.Fatal(err)
		}
		if result.ExchangeTimezone != "US/Central" {
			t.Errorf("timezone = %q, want US/Central", result.ExchangeTimezone)
		}
	})

	t.Run("holiday_absent", func(t *testing.T) {
		// Holidays are represented by absent dates in the schedule map.
		c, srv := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"exchange_time_zone": "US/Eastern",
				"schedules": map[string]any{
					"20261224": map[string]any{
						"liquid_hours": []map[string]any{
							{"opening": 1766586600, "closing": 1766600100},
						},
					},
					// Dec 25 absent = Christmas holiday
					"20261226": map[string]any{
						"liquid_hours": []map[string]any{
							{"opening": 1766759400, "closing": 1766786400},
						},
					},
				},
			})
		})
		defer srv.Close()

		result, err := c.TradingSchedule(context.Background(), 756733, "NYSE")
		if err != nil {
			t.Fatal(err)
		}
		if _, ok := result.Schedules[when.NewDate(2026, 12, 25)]; ok {
			t.Error("expected Dec 25 to be absent (holiday)")
		}
		// Dec 24 should have an early close
		xmasEve, ok := result.Schedules[when.NewDate(2026, 12, 24)]
		if !ok {
			t.Fatalf("expected schedule for 2026-12-24 (Christmas Eve) to be present")
		}
		if len(xmasEve.LiquidHours) == 0 {
			t.Fatalf("expected at least one liquid hours entry for 2026-12-24 (Christmas Eve)")
		}
		if !xmasEve.LiquidHours[0].Closing.Equal(when.DateTimeFromEpoch(1766600100)) {
			t.Errorf("Christmas Eve close = %v, expected early close timestamp", xmasEve.LiquidHours[0].Closing)
		}
	})
}
