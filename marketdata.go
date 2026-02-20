package gbkr

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/trippwill/gbkr/models"
)

// MarketDataReader provides read access to IBKR market data.
type MarketDataReader interface {
	// Snapshot returns a live market data snapshot (GET /iserver/marketdata/snapshot).
	Snapshot(ctx context.Context, params models.SnapshotParams) ([]models.Snapshot, error)

	// History returns historical OHLC bar data (GET /iserver/marketdata/history).
	History(ctx context.Context, params models.HistoryParams) (*models.HistoryResponse, error)
}

// requiredMarketDataPermissions lists the permissions needed for MarketDataReader.
var requiredMarketDataPermissions = []Permission{
	{AreaTrading, ResourceMarketData, ActionRead},
}

// MarketData returns a MarketDataReader if the client has the required permissions.
func MarketData(c *Client) (MarketDataReader, error) {
	if err := checkPermissions(c, requiredMarketDataPermissions...); err != nil {
		return nil, err
	}
	return &marketDataReader{c: c}, nil
}

type marketDataReader struct {
	c *Client
}

func (m *marketDataReader) Snapshot(ctx context.Context, params models.SnapshotParams) ([]models.Snapshot, error) {
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

func (m *marketDataReader) History(ctx context.Context, params models.HistoryParams) (*models.HistoryResponse, error) {
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
