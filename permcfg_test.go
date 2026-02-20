package gbkr

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadPermissions_Valid(t *testing.T) {
	yaml := `
permissions:
  - area: auth
    resource: session
    action: read
  - area: trading
    resource: marketdata
    action: read
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}
	if len(ps) != 2 {
		t.Fatalf("got %d permissions, want 2", len(ps))
	}

	want0 := Permission{AreaAuth, ResourceSession, ActionRead}
	if ps[0] != want0 {
		t.Errorf("ps[0] = %v, want %v", ps[0], want0)
	}
	want1 := Permission{AreaTrading, ResourceMarketData, ActionRead}
	if ps[1] != want1 {
		t.Errorf("ps[1] = %v, want %v", ps[1], want1)
	}
}

func TestLoadPermissions_Wildcards(t *testing.T) {
	yaml := `
permissions:
  - area: "*"
    resource: "*"
    action: read
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}
	if len(ps) != 1 {
		t.Fatalf("got %d permissions, want 1", len(ps))
	}

	want := Permission{0, 0, ActionRead}
	if ps[0] != want {
		t.Errorf("ps[0] = %v, want %v", ps[0], want)
	}
}

func TestLoadPermissions_EmptyWildcard(t *testing.T) {
	yaml := `
permissions:
  - area: ""
    resource: ""
    action: ""
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}
	if len(ps) != 1 {
		t.Fatalf("got %d permissions, want 1", len(ps))
	}

	want := Permission{0, 0, 0}
	if ps[0] != want {
		t.Errorf("ps[0] = %v, want %v", ps[0], want)
	}
}

func TestLoadPermissions_UnknownArea(t *testing.T) {
	yaml := `
permissions:
  - area: bogus
    resource: session
    action: read
`
	_, err := LoadPermissions(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for unknown area")
	}
	if !errors.Is(err, ErrUnknownArea) {
		t.Errorf("expected ErrUnknownArea, got %v", err)
	}
}

func TestLoadPermissions_UnknownResource(t *testing.T) {
	yaml := `
permissions:
  - area: auth
    resource: bogus
    action: read
`
	_, err := LoadPermissions(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for unknown resource")
	}
	if !errors.Is(err, ErrUnknownResource) {
		t.Errorf("expected ErrUnknownResource, got %v", err)
	}
}

func TestLoadPermissions_UnknownAction(t *testing.T) {
	yaml := `
permissions:
  - area: auth
    resource: session
    action: bogus
`
	_, err := LoadPermissions(strings.NewReader(yaml))
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
	if !errors.Is(err, ErrUnknownAction) {
		t.Errorf("expected ErrUnknownAction, got %v", err)
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
  - area: portfolio
    resource: positions
    action: read
  - area: auth
    resource: session
    action: write
`
	ps, err := LoadPermissions(strings.NewReader(yaml))
	if err != nil {
		t.Fatalf("LoadPermissions: %v", err)
	}

	// Verify the loaded set allows the expected permissions.
	if !ps.Allows(Permission{AreaPortfolio, ResourcePositions, ActionRead}) {
		t.Error("should allow portfolio.positions.read")
	}
	if !ps.Allows(Permission{AreaAuth, ResourceSession, ActionWrite}) {
		t.Error("should allow auth.session.write")
	}
	if ps.Allows(Permission{AreaTrading, ResourceMarketData, ActionRead}) {
		t.Error("should not allow trading.marketdata.read")
	}
}
