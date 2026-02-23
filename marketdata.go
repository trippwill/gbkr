package gbkr

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/trippwill/gbkr/models"
)

// MarketData provides read access to IBKR market data.
// IBKR path prefix: /iserver/marketdata/*
//
// No per-method permission check — access is gated by [Client.BrokerageSession].
type MarketData struct {
	c *Client
}

// MarketData returns a [*MarketData] handle.
// No per-method permission check — access is gated by [Client.BrokerageSession].
func (bc *BrokerageClient) MarketData() *MarketData {
	return &MarketData{c: bc.Client}
}

// Snapshot returns a live market data snapshot
// (GET /iserver/marketdata/snapshot).
func (m *MarketData) Snapshot(ctx context.Context, params models.SnapshotParams) ([]models.Snapshot, error) {
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
	if err := m.c.doGet(ctx, "/iserver/marketdata/snapshot", q, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// History returns historical OHLC bar data
// (GET /iserver/marketdata/history).
func (m *MarketData) History(ctx context.Context, params models.HistoryParams) (*models.HistoryResponse, error) {
	q := url.Values{}
	q.Set("conid", fmt.Sprintf("%d", int(params.ConID)))
	q.Set("period", params.Period.String())
	q.Set("bar", params.Bar.String())
	if params.Exchange != "" {
		q.Set("exchange", params.Exchange.String())
	}

	var result models.HistoryResponse
	if err := m.c.doGet(ctx, "/iserver/marketdata/history", q, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
