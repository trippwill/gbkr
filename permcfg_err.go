package gbkr

import "fmt"

// Sentinel errors for permission configuration.
const (
	ErrPermissionsFile   = Error("opening permissions file")
	ErrPermissionsDecode = Error("decoding permissions")
	ErrPromptRead        = Error("reading prompt response")
)

// ConfigError wraps errors from permission configuration loading.
type ConfigError struct {
	Kind Error
	Err  error
}

func (e *ConfigError) Error() string {
	return fmt.Sprintf("%s: %v", e.Kind, e.Err)
}

func (e *ConfigError) Unwrap() []error { return []error{e.Kind, e.Err} }

// ErrPermissionsFileOpen constructs a [ConfigError] for file open failures.
func ErrPermissionsFileOpen(err error) error {
	return &ConfigError{Kind: ErrPermissionsFile, Err: err}
}

// ErrPermissionsDecoding constructs a [ConfigError] for YAML decode failures.
func ErrPermissionsDecoding(err error) error {
	return &ConfigError{Kind: ErrPermissionsDecode, Err: err}
}

// ErrPromptReading constructs a [ConfigError] for stdin read failures.
func ErrPromptReading(err error) error {
	return &ConfigError{Kind: ErrPromptRead, Err: err}
}
