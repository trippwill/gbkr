package gbkr

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/trippwill/gbkr/models"
)

// MarketData provides read access to IBKR market data.
// IBKR path prefix: /iserver/marketdata/*
type MarketData struct {
	c *Client
}

// MarketData returns a [*MarketData] handle.
func (bc *BrokerageClient) MarketData() *MarketData {
	return &MarketData{c: bc.Client}
}

// Snapshot returns a live market data snapshot
// (GET /iserver/marketdata/snapshot).
func (m *MarketData) Snapshot(ctx context.Context, params models.SnapshotParams) ([]models.Snapshot, error) {
	start := time.Now()
	q := url.Values{}
	if len(params.ConIDs) > 0 {
		ids := make([]string, len(params.ConIDs))
		for i, id := range params.ConIDs {
			ids[i] = fmt.Sprintf("%d", int(id))
		}
		q.Set("conids", strings.Join(ids, ","))
	}
	if len(params.Fields) > 0 {
		fs := make([]string, len(params.Fields))
		for i, f := range params.Fields {
			fs[i] = f.String()
		}
		q.Set("fields", strings.Join(fs, ","))
	}

	var result []models.Snapshot
	err := m.c.doGet(ctx, "/iserver/marketdata/snapshot", q, &result)
	m.c.emitOp(ctx, OpMarketDataSnapshot, err, time.Since(start))
	if err != nil {
		return nil, err
	}
	return result, nil
}

// History returns historical OHLC bar data
// (GET /iserver/marketdata/history).
func (m *MarketData) History(ctx context.Context, params models.HistoryParams) (*models.HistoryResponse, error) {
	start := time.Now()
	q := url.Values{}
	q.Set("conid", fmt.Sprintf("%d", int(params.ConID)))
	q.Set("period", params.Period.String())
	q.Set("bar", params.Bar.String())
	if params.Exchange != "" {
		q.Set("exchange", params.Exchange.String())
	}

	var result models.HistoryResponse
	err := m.c.doGet(ctx, "/iserver/marketdata/history", q, &result)
	m.c.emitOp(ctx, OpMarketDataHistory, err, time.Since(start),
		slog.Int64("conid", int64(params.ConID)))
	if err != nil {
		return nil, err
	}
	return &result, nil
}
