package gbkr

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadPermissions_Valid(t *testing.T) {
	yaml := `
permissions:
  - scope: brokerage
    level: read
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}
	if !ps.Has(ScopeBrokerage, LevelRead) {
		t.Error("should grant brokerage:read")
	}
}

func TestLoadPermissions_Trade(t *testing.T) {
	yaml := `
permissions:
  - scope: brokerage
    level: trade
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}
	if !ps.Has(ScopeBrokerage, LevelTrade) {
		t.Error("should grant brokerage:trade")
	}
	// Trade implies read.
	if !ps.Has(ScopeBrokerage, LevelRead) {
		t.Error("brokerage:trade should imply brokerage:read")
	}
}

func TestLoadPermissions_UnknownLevel(t *testing.T) {
	yaml := `
permissions:
  - scope: brokerage
    level: bogus
`
	_, err := LoadPermissions(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for unknown level")
	}
	if !errors.Is(err, ErrUnknownLevel) {
		t.Errorf("expected ErrUnknownLevel, got %v", err)
	}
}

func TestLoadPermissions_UnknownScope(t *testing.T) {
	yaml := `
permissions:
  - scope: bogus
    level: read
`
	_, err := LoadPermissions(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for unknown scope")
	}
	if !errors.Is(err, ErrUnknownScope) {
		t.Errorf("expected ErrUnknownScope, got %v", err)
	}
}

func TestLoadPermissions_InvalidYAML(t *testing.T) {
	yaml := `{{{not yaml`

	_, err := LoadPermissions(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
	if !errors.Is(err, ErrPermissionsDecode) {
		t.Errorf("expected ErrPermissionsDecode, got %v", err)
	}
}

func TestLoadPermissions_RoundTrip(t *testing.T) {
	yaml := `
permissions:
  - scope: brokerage
    level: read
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}

	if !ps.Has(ScopeBrokerage, LevelRead) {
		t.Error("should allow brokerage:read")
	}
	if ps.Has(ScopeBrokerage, LevelTrade) {
		t.Error("should not allow brokerage:trade")
	}
}
