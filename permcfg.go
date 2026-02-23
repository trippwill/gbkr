package gbkr

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// permissionsConfig represents the YAML permissions file format.
type permissionsConfig struct {
	Permissions []permissionEntry `yaml:"permissions"`
}

type permissionEntry struct {
	Scope string `yaml:"scope"`
	Level string `yaml:"level"`
}

// LoadPermissionsFromFile reads a YAML permissions file and returns a PermissionSet.
func LoadPermissionsFromFile(path string) (PermissionSet, error) {
	f, err := os.Open(path) //nolint:gosec // caller-provided config path
	if err != nil {
		return nil, ErrPermissionsFileOpen(err)
	}
	defer f.Close()
	return LoadPermissions(f)
}

// LoadPermissions reads YAML-formatted permissions from a reader.
func LoadPermissions(r io.Reader) (PermissionSet, error) {
	var cfg permissionsConfig
	if err := yaml.NewDecoder(r).Decode(&cfg); err != nil {
		return nil, ErrPermissionsDecoding(err)
	}

	ps := make(PermissionSet, len(cfg.Permissions))
	for _, e := range cfg.Permissions {
		scope := Scope(e.Scope)
		if scope != ScopeBrokerage {
			return nil, ErrUnknownScopeValue(e.Scope)
		}
		level, err := ParseLevel(e.Level)
		if err != nil {
			return nil, err
		}
		ps.Grant(Permission{Scope: scope, Level: level})
	}
	return ps, nil
}

// InteractivePrompter implements [Prompter] by asking the user to grant
// each missing permission individually via stderr/stdin.
type InteractivePrompter struct{}

// Prompt asks the user to grant each missing permission.
func (InteractivePrompter) Prompt(missing []Permission) ([]Permission, error) {
	reader := bufio.NewReader(os.Stdin)
	var granted []Permission

	for _, p := range missing {
		fmt.Fprintf(os.Stderr, "Grant %s:%s access? [y/N] ", p.Scope, p.Level)
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, ErrPromptReading(err)
		}
		line = strings.TrimSpace(strings.ToLower(line))
		if line == "y" || line == "yes" {
			granted = append(granted, p)
		}
	}
	return granted, nil
}
