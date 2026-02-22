package gbkr

import (
	"errors"
	"testing"
)

func TestLevel_Hierarchy(t *testing.T) {
	// Trade level is higher than read.
	if LevelTrade <= LevelRead {
		t.Error("LevelTrade should be greater than LevelRead")
	}
	// Read level is higher than none.
	if LevelRead <= LevelNone {
		t.Error("LevelRead should be greater than LevelNone")
	}
}

func TestPermissionSet_Has(t *testing.T) {
	tests := []struct {
		name  string
		set   PermissionSet
		scope Scope
		level Level
		want  bool
	}{
		{
			name:  "empty set denies everything",
			set:   PermissionSet{},
			scope: ScopeBrokerage,
			level: LevelRead,
			want:  false,
		},
		{
			name:  "nil set denies",
			set:   nil,
			scope: ScopeBrokerage,
			level: LevelRead,
			want:  false,
		},
		{
			name:  "exact level match",
			set:   PermissionSet{ScopeBrokerage: LevelRead},
			scope: ScopeBrokerage,
			level: LevelRead,
			want:  true,
		},
		{
			name:  "trade implies read",
			set:   PermissionSet{ScopeBrokerage: LevelTrade},
			scope: ScopeBrokerage,
			level: LevelRead,
			want:  true,
		},
		{
			name:  "read does not imply trade",
			set:   PermissionSet{ScopeBrokerage: LevelRead},
			scope: ScopeBrokerage,
			level: LevelTrade,
			want:  false,
		},
		{
			name:  "wrong scope denies",
			set:   PermissionSet{ScopeBrokerage: LevelTrade},
			scope: Scope("other"),
			level: LevelRead,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.set.Has(tt.scope, tt.level)
			if got != tt.want {
				t.Errorf("PermissionSet.Has(%q, %v) = %v, want %v", tt.scope, tt.level, got, tt.want)
			}
		})
	}
}

func TestPermissionSet_Grant(t *testing.T) {
	ps := make(PermissionSet)

	// Grant read.
	ps.Grant(Permission{Scope: ScopeBrokerage, Level: LevelRead})
	if !ps.Has(ScopeBrokerage, LevelRead) {
		t.Error("should have read after grant")
	}

	// Upgrade to trade.
	ps.Grant(Permission{Scope: ScopeBrokerage, Level: LevelTrade})
	if !ps.Has(ScopeBrokerage, LevelTrade) {
		t.Error("should have trade after upgrade")
	}

	// Downgrade attempt is a no-op.
	ps.Grant(Permission{Scope: ScopeBrokerage, Level: LevelRead})
	if !ps.Has(ScopeBrokerage, LevelTrade) {
		t.Error("downgrade should be a no-op")
	}
}

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name string
		perm Permission
		want string
	}{
		{"brokerage read", Permission{ScopeBrokerage, LevelRead}, "brokerage:read"},
		{"brokerage trade", Permission{ScopeBrokerage, LevelTrade}, "brokerage:trade"},
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

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input   string
		want    Level
		wantErr bool
	}{
		{"read", LevelRead, false},
		{"trade", LevelTrade, false},
		{"unknown", LevelNone, true},
		{"", LevelNone, true},
	}

	for _, tt := range tests {
		name := tt.input
		if name == "" {
			name = "(empty)"
		}
		t.Run(name, func(t *testing.T) {
			got, err := ParseLevel(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseLevel(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestLevel_String_Fallback(t *testing.T) {
	l := Level(99)
	if got := l.String(); got != "Level(99)" {
		t.Errorf("Level(99).String() = %q, want %q", got, "Level(99)")
	}
}

func TestReadOnly_Convenience(t *testing.T) {
	ps := ReadOnly()
	if !ps.Has(ScopeBrokerage, LevelRead) {
		t.Error("ReadOnly() should grant brokerage:read")
	}
	if ps.Has(ScopeBrokerage, LevelTrade) {
		t.Error("ReadOnly() should not grant brokerage:trade")
	}
}

func TestFullAccess_Convenience(t *testing.T) {
	ps := FullAccess()
	if !ps.Has(ScopeBrokerage, LevelTrade) {
		t.Error("FullAccess() should grant brokerage:trade")
	}
	if !ps.Has(ScopeBrokerage, LevelRead) {
		t.Error("FullAccess() should imply brokerage:read")
	}
}

func TestPermissionSet_Clone(t *testing.T) {
	ps := PermissionSet{ScopeBrokerage: LevelRead}
	clone := ps.Clone()

	// Mutate clone.
	clone[ScopeBrokerage] = LevelTrade
	if ps[ScopeBrokerage] != LevelRead {
		t.Error("clone mutation should not affect original")
	}
}

func TestPermissionSet_Clone_Nil(t *testing.T) {
	var ps PermissionSet
	clone := ps.Clone()
	if clone != nil {
		t.Error("clone of nil should be nil")
	}
}

// mockPrompter implements Prompter for testing.
type mockPrompter struct {
	grant []Permission
	err   error
}

func (m *mockPrompter) Prompt(missing []Permission) ([]Permission, error) {
	return m.grant, m.err
}

func TestCheckPermissions_Granted(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{ScopeBrokerage: LevelRead}),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, ScopeBrokerage, LevelRead)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestCheckPermissions_Denied(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, ScopeBrokerage, LevelRead)
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

func TestCheckPermissions_NilPermissions_Denied(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, ScopeBrokerage, LevelRead)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestCheckPermissions_PrompterGrants(t *testing.T) {
	needed := Permission{Scope: ScopeBrokerage, Level: LevelRead}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
		WithPrompter(&mockPrompter{grant: []Permission{needed}}),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, ScopeBrokerage, LevelRead)
	if err != nil {
		t.Errorf("expected no error after prompt grant, got %v", err)
	}

	// Permission should be persisted in the client.
	if !c.Permissions().Has(ScopeBrokerage, LevelRead) {
		t.Error("granted permission not persisted")
	}
}

func TestCheckPermissions_PrompterPartialGrant(t *testing.T) {
	// Prompter grants read, but trade is needed.
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
		WithPrompter(&mockPrompter{grant: []Permission{
			{Scope: ScopeBrokerage, Level: LevelRead},
		}}),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, ScopeBrokerage, LevelTrade)
	if err == nil {
		t.Fatal("expected error for partial grant")
	}
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestCheckPermissions_PrompterNilPermissions(t *testing.T) {
	// Client has nil permissions; prompter grants what's needed.
	needed := Permission{Scope: ScopeBrokerage, Level: LevelTrade}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPrompter(&mockPrompter{grant: []Permission{needed}}),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	err = checkPermissions(c, ScopeBrokerage, LevelTrade)
	if err != nil {
		t.Errorf("expected no error after prompt grant from nil, got %v", err)
	}
}
