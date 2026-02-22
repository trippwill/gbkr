package models

import (
	"encoding/json"
	"testing"
)

func TestPosition_UnmarshalFlexibleStrikeAndMultiplier(t *testing.T) {
	t.Parallel()

	var p Position
	payload := []byte(`{
		"acctId":"U123",
		"conid":123,
		"assetClass":"OPT",
		"ticker":"AAPL",
		"strike":"190",
		"multiplier":"100"
	}`)
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("unmarshal position: %v", err)
	}
	if p.Strike != 190 {
		t.Fatalf("expected strike 190, got %v", p.Strike)
	}
	if p.Multiplier != 100 {
		t.Fatalf("expected multiplier 100, got %v", p.Multiplier)
	}
}
