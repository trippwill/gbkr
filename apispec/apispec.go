// Package apispec provides the embedded IBKR Client Portal Gateway OpenAPI specification.
package apispec

import _ "embed"

//go:embed api-docs.json
var raw []byte

// JSON returns the raw OpenAPI specification as JSON bytes.
func JSON() []byte { return raw }

// Version is the API specification version this library was built against.
const Version = "2.27.0"
