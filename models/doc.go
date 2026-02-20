// Package models defines request and response types for the IBKR REST API.
//
// Types are hand-crafted to match the IBKR Client Portal Gateway endpoints.
// Strongly-typed aliases ([AccountID], [ConID], [Currency]) prevent parameter
// confusion, and structured types ([BarSize], [TimePeriod]) enforce valid
// market data query parameters.
package models
