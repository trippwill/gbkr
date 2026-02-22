package gbkr

import (
	"fmt"
	"maps"
)

// Scope identifies a capability area for permission gating.
type Scope string

const (
	ScopeBrokerage Scope = "brokerage"
)

// Level identifies the access tier within a scope.
// Higher levels imply all lower levels within the same scope.
type Level int

const (
	LevelNone  Level = 0
	LevelRead  Level = 1
	LevelTrade Level = 2
)

var levelNames = map[Level]string{
	LevelRead:  "read",
	LevelTrade: "trade",
}

var levelsByName = map[string]Level{
	"read":  LevelRead,
	"trade": LevelTrade,
}

func (l Level) String() string {
	if s, ok := levelNames[l]; ok {
		return s
	}
	return fmt.Sprintf("Level(%d)", int(l))
}

// ParseLevel converts a string to a Level.
func ParseLevel(s string) (Level, error) {
	if l, ok := levelsByName[s]; ok {
		return l, nil
	}
	return LevelNone, ErrUnknownLevelValue(s)
}

// Permission represents a Scope:Level capability grant.
type Permission struct {
	Scope Scope
	Level Level
}

func (p Permission) String() string {
	return string(p.Scope) + ":" + p.Level.String()
}

// PermissionSet maps scopes to their granted access level.
type PermissionSet map[Scope]Level

// Has reports whether the set grants at least the requested level for the given scope.
func (ps PermissionSet) Has(scope Scope, level Level) bool {
	granted, ok := ps[scope]
	return ok && granted >= level
}

// Grant adds or upgrades a permission in the set.
// If the scope is not yet granted, or the new level exceeds the current grant,
// the level is updated.
func (ps PermissionSet) Grant(p Permission) {
	if current, ok := ps[p.Scope]; !ok || p.Level > current {
		ps[p.Scope] = p.Level
	}
}

// Clone returns a shallow copy of the permission set.
func (ps PermissionSet) Clone() PermissionSet {
	if ps == nil {
		return nil
	}
	return maps.Clone(ps)
}

// ReadOnly returns a permission set granting brokerage read access.
// Gateway access is always available and does not need explicit permission.
func ReadOnly() PermissionSet {
	return PermissionSet{ScopeBrokerage: LevelRead}
}

// FullAccess grants read + trade access across all current scopes.
func FullAccess() PermissionSet {
	return PermissionSet{ScopeBrokerage: LevelTrade}
}

// Prompter is called by capability constructors when the static permission set
// is insufficient. It receives the missing permissions and returns the subset
// the user grants. Ungranted permissions cause [PermissionDeniedError].
type Prompter interface {
	Prompt(missing []Permission) ([]Permission, error)
}

// checkPermissions verifies the client has at least the requested level for the scope.
// If the permission is missing and a [Prompter] is configured, it calls the prompter
// to JIT-grant. Granted permissions are added to the client's live set.
// Returns [PermissionDeniedError] if the permission remains missing.
func checkPermissions(c *Client, scope Scope, level Level) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.permissions != nil && c.permissions.Has(scope, level) {
		return nil
	}

	missing := []Permission{{Scope: scope, Level: level}}

	if c.prompter == nil {
		return ErrPermissionsDenied(missing, missing)
	}

	granted, err := c.prompter.Prompt(missing)
	if err != nil {
		return err
	}
	if c.permissions == nil {
		c.permissions = make(PermissionSet)
	}
	for _, p := range granted {
		c.permissions.Grant(p)
	}

	if !c.permissions.Has(scope, level) {
		return ErrPermissionsDenied(missing, missing)
	}
	return nil
}
