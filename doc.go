// Package gbkr provides a permission-gated client for the IBKR REST API.
//
// Consumers obtain capability interfaces ([SessionClient], [AccountLister], [AccountReader],
// [PositionReader], [MarketDataReader]) via constructor functions that enforce
// a three-tier permission model at both compile time and runtime.
package gbkr
