package gbkr

import (
	"net/http"
	"os"
	"path/filepath"
	"testing"
)

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithHTTPClient(custom),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.httpClient != custom {
		t.Error("WithHTTPClient did not set the client")
	}
}

func TestWithPermissionsFromFile(t *testing.T) {
	yaml := []byte(`
permissions:
  - area: auth
    resource: session
    action: read
`)
	dir := t.TempDir()
	path := filepath.Join(dir, "perms.yaml")
	if err := os.WriteFile(path, yaml, 0o600); err != nil {
		t.Fatal(err)
	}

	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissionsFromFile(path),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	want := Permission{AreaAuth, ResourceSession, ActionRead}
	if !c.Permissions().Allows(want) {
		t.Errorf("expected permission %v to be allowed", want)
	}
}

func TestWithPermissionsFromFile_Missing(t *testing.T) {
	_, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissionsFromFile("/nonexistent/perms.yaml"),
		WithRateLimit(nil),
	)
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWithPrompter(t *testing.T) {
	p := &mockPrompter{}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPrompter(p),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.prompter != p {
		t.Error("WithPrompter did not set the prompter")
	}
}

func TestClient_Permissions(t *testing.T) {
	ps := PermissionSet{{AreaAuth, ResourceSession, ActionRead}}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(ps),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	got := c.Permissions()
	if len(got) != 1 {
		t.Fatalf("Permissions() len = %d, want 1", len(got))
	}
	if got[0] != ps[0] { //nolint:gosec // length checked above
		t.Errorf("Permissions() = %v, want %v", got, ps)
	}
}
