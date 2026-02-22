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

func TestPosition_UnmarshalNumericStrikeAndMultiplier(t *testing.T) {
	t.Parallel()

	var p Position
	payload := []byte(`{
		"acctId":"U123",
		"conid":456,
		"assetClass":"OPT",
		"strike": 195.5,
		"multiplier": 100.0,
		"undConid": 265598,
		"expiry": "20240119",
		"putOrCall": "C"
	}`)
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Strike != 195.5 {
		t.Errorf("Strike = %f, want 195.5", p.Strike)
	}
	if p.Multiplier != 100.0 {
		t.Errorf("Multiplier = %f, want 100", p.Multiplier)
	}
	if p.UndConID != 265598 {
		t.Errorf("UndConID = %d, want 265598", p.UndConID)
	}
	if p.Expiry != "20240119" {
		t.Errorf("Expiry = %q", p.Expiry)
	}
	if p.PutOrCall != "C" {
		t.Errorf("PutOrCall = %q", p.PutOrCall)
	}
}

func TestPosition_UnmarshalEmptyStringStrike(t *testing.T) {
	t.Parallel()

	var p Position
	payload := []byte(`{"conid":1, "strike":"", "multiplier":""}`)
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Strike != 0 {
		t.Errorf("Strike = %f, want 0", p.Strike)
	}
	if p.Multiplier != 0 {
		t.Errorf("Multiplier = %f, want 0", p.Multiplier)
	}
}

func TestPosition_UnmarshalNilStrike(t *testing.T) {
	t.Parallel()

	var p Position
	payload := []byte(`{"conid":1}`)
	if err := json.Unmarshal(payload, &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Strike != 0 {
		t.Errorf("Strike = %f, want 0", p.Strike)
	}
}

func TestPosition_UnmarshalInvalidStrike(t *testing.T) {
	t.Parallel()

	var p Position
	payload := []byte(`{"conid":1, "strike":"not-a-number"}`)
	if err := json.Unmarshal(payload, &p); err == nil {
		t.Fatal("expected error for invalid strike")
	}
}

func TestPosition_UnmarshalUnsupportedStrikeType(t *testing.T) {
	t.Parallel()

	var p Position
	payload := []byte(`{"conid":1, "strike": true}`)
	if err := json.Unmarshal(payload, &p); err == nil {
		t.Fatal("expected error for boolean strike")
	}
}
