package gbkr

import (
	"errors"
	"testing"
)

func TestPermission_Allows(t *testing.T) {
	tests := []struct {
		name      string
		grant     Permission
		requested Permission
		want      bool
	}{
		{
			name:      "exact match",
			grant:     Permission{AreaAuth, ResourceSession, ActionRead},
			requested: Permission{AreaAuth, ResourceSession, ActionRead},
			want:      true,
		},
		{
			name:      "full wildcard allows anything",
			grant:     Permission{0, 0, 0},
			requested: Permission{AreaTrading, ResourceMarketData, ActionWrite},
			want:      true,
		},
		{
			name:      "area wildcard",
			grant:     Permission{0, ResourceSession, ActionRead},
			requested: Permission{AreaAuth, ResourceSession, ActionRead},
			want:      true,
		},
		{
			name:      "resource wildcard",
			grant:     Permission{AreaAuth, 0, ActionRead},
			requested: Permission{AreaAuth, ResourceAccounts, ActionRead},
			want:      true,
		},
		{
			name:      "action wildcard",
			grant:     Permission{AreaAuth, ResourceSession, 0},
			requested: Permission{AreaAuth, ResourceSession, ActionWrite},
			want:      true,
		},
		{
			name:      "area mismatch",
			grant:     Permission{AreaAuth, ResourceSession, ActionRead},
			requested: Permission{AreaTrading, ResourceSession, ActionRead},
			want:      false,
		},
		{
			name:      "resource mismatch",
			grant:     Permission{AreaAuth, ResourceSession, ActionRead},
			requested: Permission{AreaAuth, ResourceAccounts, ActionRead},
			want:      false,
		},
		{
			name:      "action mismatch",
			grant:     Permission{AreaAuth, ResourceSession, ActionRead},
			requested: Permission{AreaAuth, ResourceSession, ActionWrite},
			want:      false,
		},
		{
			name:      "two wildcards",
			grant:     Permission{AreaTrading, 0, 0},
			requested: Permission{AreaTrading, ResourceMarketData, ActionRead},
			want:      true,
		},
		{
			name:      "two wildcards wrong area",
			grant:     Permission{AreaTrading, 0, 0},
			requested: Permission{AreaPortfolio, ResourcePositions, ActionRead},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.grant.Allows(tt.requested)
			if got != tt.want {
				t.Errorf("Permission%v.Allows(%v) = %v, want %v", tt.grant, tt.requested, got, tt.want)
			}
		})
	}
}

func TestPermissionSet_Allows(t *testing.T) {
	tests := []struct {
		name      string
		set       PermissionSet
		requested Permission
		want      bool
	}{
		{
			name:      "empty set denies",
			set:       PermissionSet{},
			requested: Permission{AreaAuth, ResourceSession, ActionRead},
			want:      false,
		},
		{
			name:      "single grant matches",
			set:       PermissionSet{{AreaAuth, ResourceSession, ActionRead}},
			requested: Permission{AreaAuth, ResourceSession, ActionRead},
			want:      true,
		},
		{
			name:      "single grant no match",
			set:       PermissionSet{{AreaAuth, ResourceSession, ActionRead}},
			requested: Permission{AreaTrading, ResourceMarketData, ActionRead},
			want:      false,
		},
		{
			name:      "multi grant second matches",
			set:       PermissionSet{{AreaAuth, ResourceSession, ActionRead}, {AreaTrading, ResourceMarketData, ActionRead}},
			requested: Permission{AreaTrading, ResourceMarketData, ActionRead},
			want:      true,
		},
		{
			name:      "wildcard grant in set",
			set:       PermissionSet{{0, 0, ActionRead}},
			requested: Permission{AreaPortfolio, ResourcePositions, ActionRead},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.set.Allows(tt.requested)
			if got != tt.want {
				t.Errorf("PermissionSet.Allows(%v) = %v, want %v", tt.requested, got, tt.want)
			}
		})
	}
}

func TestPermissionSet_AllowsAll(t *testing.T) {
	tests := []struct {
		name      string
		set       PermissionSet
		requested []Permission
		want      bool
	}{
		{
			name: "all met",
			set:  ReadOnlyAuth(),
			requested: []Permission{
				{AreaAuth, ResourceSession, ActionRead},
				{AreaAuth, ResourceSession, ActionWrite},
			},
			want: true,
		},
		{
			name: "partial met",
			set:  ReadOnly(),
			requested: []Permission{
				{AreaAuth, ResourceSession, ActionRead},
				{AreaAuth, ResourceSession, ActionWrite},
			},
			want: false,
		},
		{
			name:      "empty requested",
			set:       PermissionSet{},
			requested: []Permission{},
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.set.AllowsAll(tt.requested...)
			if got != tt.want {
				t.Errorf("PermissionSet.AllowsAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseArea(t *testing.T) {
	tests := []struct {
		input   string
		want    Area
		wantErr bool
	}{
		{"auth", AreaAuth, false},
		{"trading", AreaTrading, false},
		{"portfolio", AreaPortfolio, false},
		{"*", 0, false},
		{"", 0, false},
		{"unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseArea(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseArea(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseArea(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseResource(t *testing.T) {
	tests := []struct {
		input   string
		want    Resource
		wantErr bool
	}{
		{"session", ResourceSession, false},
		{"accounts", ResourceAccounts, false},
		{"positions", ResourcePositions, false},
		{"marketdata", ResourceMarketData, false},
		{"summary", ResourceSummary, false},
		{"ledger", ResourceLedger, false},
		{"allocation", ResourceAllocation, false},
		{"*", 0, false},
		{"", 0, false},
		{"unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseResource(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResource(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseResource(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseAction(t *testing.T) {
	tests := []struct {
		input   string
		want    Action
		wantErr bool
	}{
		{"read", ActionRead, false},
		{"write", ActionWrite, false},
		{"invalidate", ActionInvalidate, false},
		{"unsubscribe", ActionUnsubscribe, false},
		{"*", 0, false},
		{"", 0, false},
		{"unknown", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseAction(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAction(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseAction(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name string
		perm Permission
		want string
	}{
		{"fully specified", Permission{AreaAuth, ResourceSession, ActionRead}, "auth.session.read"},
		{"all wildcards", Permission{0, 0, 0}, "*.*.*"},
		{"area wildcard", Permission{0, ResourceSession, ActionRead}, "*.session.read"},
		{"resource wildcard", Permission{AreaAuth, 0, ActionRead}, "auth.*.read"},
		{"action wildcard", Permission{AreaAuth, ResourceSession, 0}, "auth.session.*"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.perm.String()
			if got != tt.want {
				t.Errorf("Permission.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// mockPrompter implements Prompter for testing.
type mockPrompter struct {
	grant PermissionSet
	err   error
}

func (m *mockPrompter) Prompt(missing []Permission) (PermissionSet, error) {
	return m.grant, m.err
}

func TestCheckPermissions_Granted(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{{AreaAuth, ResourceSession, ActionRead}}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, Permission{AreaAuth, ResourceSession, ActionRead})
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckPermissions_Denied(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, Permission{AreaAuth, ResourceSession, ActionRead})
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}

	var pde *PermissionDeniedError
	if !errors.As(err, &pde) {
		t.Fatal("expected *PermissionDeniedError")
	}
	if len(pde.Missing) != 1 {
		t.Errorf("Missing = %d, want 1", len(pde.Missing))
	}
}

func TestCheckPermissions_PrompterGrants(t *testing.T) {
	needed := Permission{AreaTrading, ResourceMarketData, ActionRead}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
		WithPrompter(&mockPrompter{grant: PermissionSet{needed}}),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, needed)
	if err != nil {
		t.Errorf("expected no error after prompt grant, got %v", err)
	}

	// Permission should be persisted in the client.
	if !c.Permissions().Allows(needed) {
		t.Error("granted permission not persisted")
	}
}

func TestCheckPermissions_PrompterPartialGrant(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
		WithPrompter(&mockPrompter{grant: PermissionSet{
			{AreaAuth, ResourceSession, ActionRead},
		}}),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Request two permissions but prompter only grants one.
	err = checkPermissions(c,
		Permission{AreaAuth, ResourceSession, ActionRead},
		Permission{AreaAuth, ResourceSession, ActionWrite},
	)
	if err == nil {
		t.Fatal("expected error for partial grant")
	}
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestEnumString_Fallback(t *testing.T) {
	area := Area(99)
	if got := area.String(); got != "Area(99)" {
		t.Errorf("Area(99).String() = %q, want %q", got, "Area(99)")
	}

	resource := Resource(99)
	if got := resource.String(); got != "Resource(99)" {
		t.Errorf("Resource(99).String() = %q, want %q", got, "Resource(99)")
	}

	action := Action(99)
	if got := action.String(); got != "Action(99)" {
		t.Errorf("Action(99).String() = %q, want %q", got, "Action(99)")
	}
}
