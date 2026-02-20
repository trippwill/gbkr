// Command gbkr is a CLI for testing the gbkr client library against
// a running IBKR Client Portal Gateway. It supports configurable
// permissions via file or interactive prompt.
//
// Usage:
//
//	gbkr [flags]
//
// Flags:
//
//	--base-url          IBKR API base URL (default: https://localhost:5000/v1/api)
//	--permissions-file  YAML permissions file (optional floor; JIT prompts for anything missing)
//	--insecure          Skip TLS verification for local gateway
package main
