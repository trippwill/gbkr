package gbkr

import "fmt"

// Area represents the top-level permission domain.
type Area int

const (
	AreaAuth      Area = iota + 1 // Authorization and session management
	AreaTrading                   // Trading operations (iserver endpoints)
	AreaPortfolio                 // Portfolio data (portfolio endpoints)
)

var areaNames = map[Area]string{
	AreaAuth:      "auth",
	AreaTrading:   "trading",
	AreaPortfolio: "portfolio",
}

var areasByName = map[string]Area{
	"auth":      AreaAuth,
	"trading":   AreaTrading,
	"portfolio": AreaPortfolio,
}

func (a Area) String() string {
	if s, ok := areaNames[a]; ok {
		return s
	}
	return fmt.Sprintf("Area(%d)", int(a))
}

// ParseArea converts a string to an Area. Returns 0 (wildcard) for "*" or "".
func ParseArea(s string) (Area, error) {
	if s == "*" || s == "" {
		return 0, nil
	}
	if a, ok := areasByName[s]; ok {
		return a, nil
	}
	return 0, ErrUnknownAreaValue(s)
}

// Resource represents the API resource within an area.
type Resource int

const (
	ResourceSession    Resource = iota + 1 // Auth sessions
	ResourceAccounts                       // Account listing and summaries
	ResourcePositions                      // Portfolio positions
	ResourceMarketData                     // Market data snapshots and history
	ResourceSummary                        // Portfolio summaries
	ResourceLedger                         // Account ledger
	ResourceAllocation                     // Account allocations
	ResourceTrades                         // Trade executions and transaction history
	ResourceContracts                      // Contract details and search
)

var resourceNames = map[Resource]string{
	ResourceSession:    "session",
	ResourceAccounts:   "accounts",
	ResourcePositions:  "positions",
	ResourceMarketData: "marketdata",
	ResourceSummary:    "summary",
	ResourceLedger:     "ledger",
	ResourceAllocation: "allocation",
	ResourceTrades:     "trades",
	ResourceContracts:  "contracts",
}

var resourcesByName = map[string]Resource{
	"session":    ResourceSession,
	"accounts":   ResourceAccounts,
	"positions":  ResourcePositions,
	"marketdata": ResourceMarketData,
	"summary":    ResourceSummary,
	"ledger":     ResourceLedger,
	"allocation": ResourceAllocation,
	"trades":     ResourceTrades,
	"contracts":  ResourceContracts,
}

func (r Resource) String() string {
	if s, ok := resourceNames[r]; ok {
		return s
	}
	return fmt.Sprintf("Resource(%d)", int(r))
}

// ParseResource converts a string to a Resource. Returns 0 (wildcard) for "*" or "".
func ParseResource(s string) (Resource, error) {
	if s == "*" || s == "" {
		return 0, nil
	}
	if r, ok := resourcesByName[s]; ok {
		return r, nil
	}
	return 0, ErrUnknownResourceValue(s)
}

// Action represents the type of operation.
type Action int

const (
	ActionRead        Action = iota + 1 // Read/query data
	ActionWrite                         // Create or modify data
	ActionInvalidate                    // Invalidate caches
	ActionUnsubscribe                   // Close data streams
)

var actionNames = map[Action]string{
	ActionRead:        "read",
	ActionWrite:       "write",
	ActionInvalidate:  "invalidate",
	ActionUnsubscribe: "unsubscribe",
}

var actionsByName = map[string]Action{
	"read":        ActionRead,
	"write":       ActionWrite,
	"invalidate":  ActionInvalidate,
	"unsubscribe": ActionUnsubscribe,
}

func (a Action) String() string {
	if s, ok := actionNames[a]; ok {
		return s
	}
	return fmt.Sprintf("Action(%d)", int(a))
}

// ParseAction converts a string to an Action. Returns 0 (wildcard) for "*" or "".
func ParseAction(s string) (Action, error) {
	if s == "*" || s == "" {
		return 0, nil
	}
	if a, ok := actionsByName[s]; ok {
		return a, nil
	}
	return 0, ErrUnknownActionValue(s)
}

// Permission is a three-tier capability grant. Zero-value fields act as wildcards.
type Permission struct {
	Area     Area
	Resource Resource
	Action   Action
}

func (p Permission) String() string {
	a, r, act := "*", "*", "*"
	if p.Area != 0 {
		a = p.Area.String()
	}
	if p.Resource != 0 {
		r = p.Resource.String()
	}
	if p.Action != 0 {
		act = p.Action.String()
	}
	return a + "." + r + "." + act
}

// Allows reports whether this permission grant covers the requested permission.
// A zero-value field in the grant matches any value in the request.
func (p Permission) Allows(requested Permission) bool {
	if p.Area != 0 && p.Area != requested.Area {
		return false
	}
	if p.Resource != 0 && p.Resource != requested.Resource {
		return false
	}
	if p.Action != 0 && p.Action != requested.Action {
		return false
	}
	return true
}

// PermissionSet is a collection of granted permissions.
type PermissionSet []Permission

// Allows reports whether any permission in the set covers the request.
func (ps PermissionSet) Allows(requested Permission) bool {
	for _, p := range ps {
		if p.Allows(requested) {
			return true
		}
	}
	return false
}

// AllowsAll reports whether the set covers all requested permissions.
func (ps PermissionSet) AllowsAll(requested ...Permission) bool {
	for _, r := range requested {
		if !ps.Allows(r) {
			return false
		}
	}
	return true
}

// ReadOnly returns a permission set granting read access to all areas and resources.
func ReadOnly() PermissionSet {
	return PermissionSet{
		{0, 0, ActionRead},
	}
}

// ReadOnlyAuth returns a permission set granting read access to all areas and resources,
// plus auth session write for login.
func ReadOnlyAuth() PermissionSet {
	return PermissionSet{
		{0, 0, ActionRead},
		{AreaAuth, ResourceSession, ActionWrite},
	}
}

// AllPermissions returns a permission set granting unrestricted access.
func AllPermissions() PermissionSet {
	return PermissionSet{
		{0, 0, 0},
	}
}

// Prompter is called by capability constructors when the static permission set
// is insufficient. It receives the missing permissions and returns the subset
// the user grants. Ungranted permissions cause [PermissionDeniedError].
type Prompter interface {
	Prompt(missing []Permission) (PermissionSet, error)
}

// checkPermissions verifies required permissions against the client's granted set.
// If permissions are missing and a [Prompter] is configured, it calls the prompter
// to JIT-grant the missing permissions. Granted permissions are added to the
// client's live set. Returns [PermissionDeniedError] if permissions remain missing.
func checkPermissions(c *Client, required ...Permission) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	missing := findMissing(c.permissions, required)
	if len(missing) == 0 {
		return nil
	}

	if c.prompter == nil {
		return ErrPermissionsDenied(required, missing)
	}

	granted, err := c.prompter.Prompt(missing)
	if err != nil {
		return err
	}
	c.permissions = append(c.permissions, granted...)

	missing = findMissing(c.permissions, required)
	if len(missing) > 0 {
		return ErrPermissionsDenied(required, missing)
	}
	return nil
}

func findMissing(granted PermissionSet, required []Permission) []Permission {
	var missing []Permission
	for _, r := range required {
		if !granted.Allows(r) {
			missing = append(missing, r)
		}
	}
	return missing
}
