// Package flex provides a Go client for the IBKR Flex Web Service, a standalone
// HTTP API for retrieving Activity Statement and Trade Confirmation reports.
//
// The Flex Web Service is completely independent of the IBKR Client Portal
// Gateway that [github.com/trippwill/gbkr] wraps. It uses a separate
// authentication mechanism (long-lived API token) and communicates with
// IBKR's cloud servers at ndcdyn.interactivebrokers.com.
//
// # Two-Step Retrieval Protocol
//
// Flex reports are retrieved in two steps:
//
//  1. [Client.SendRequest] submits a query ID and receives a reference code.
//  2. [Client.GetStatement] polls with the reference code until the report
//     is ready, then returns the parsed [Statement].
//
// The convenience method [Client.FetchReport] combines both steps with
// configurable retry/backoff.
//
// # Rate Limits
//
// The Flex Web Service enforces a limit of 1 request per second and
// 10 requests per minute per token. Activity Statements update once daily
// after market close; Trade Confirmations update within 5–10 minutes of
// execution.
//
// # Data Sections
//
// This package parses the following Activity Flex Query sections:
//
//   - Trades — execution data with commissions, cost basis, and realized PnL
//   - Cash Transactions — dividends, margin interest, fees, withholding tax
//   - Option Exercises, Assignments & Expirations — covered call lifecycle events
//   - Commission Details — granular per-trade fee breakdown
//
// Flex queries must be pre-configured in the IBKR Client Portal. The query
// template determines which sections and fields are included in the report.
package flex
