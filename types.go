package gbkr

import "fmt"

// AccountID is an IBKR account identifier (e.g., "U1234567", "DU123456").
type AccountID string

func (id AccountID) String() string { return string(id) }

// ConID is an IBKR contract identifier.
type ConID int

func (id ConID) String() string { return fmt.Sprintf("%d", int(id)) }

// Currency is an ISO 4217 currency code (e.g., "USD", "EUR").
type Currency string

func (c Currency) String() string { return string(c) }

// Exchange is a market exchange identifier (e.g., "NASDAQ", "NYSE", "SMART").
type Exchange string

func (e Exchange) String() string { return string(e) }

// OrderID is an IBKR order identifier.
type OrderID string

func (id OrderID) String() string { return string(id) }

// AlertID is an IBKR alert identifier.
type AlertID string

func (id AlertID) String() string { return string(id) }

// AssetClass is a security asset classification (e.g., "STK", "OPT", "FUT").
type AssetClass string

func (a AssetClass) String() string { return string(a) }

// Known asset class values from the IBKR API.
const (
	AssetBill   AssetClass = "BILL"   // Treasury bills
	AssetBond   AssetClass = "BOND"   // Bonds
	AssetCash   AssetClass = "CASH"   // Forex / cash
	AssetCFD    AssetClass = "CFD"    // Contracts for difference
	AssetCombo  AssetClass = "COMB"   // Combination / spread orders
	AssetFOP    AssetClass = "FOP"    // Futures options
	AssetFund   AssetClass = "FUND"   // Mutual funds
	AssetFuture AssetClass = "FUT"    // Futures
	AssetOption AssetClass = "OPT"    // Options
	AssetSSF    AssetClass = "SSF"    // Single stock futures
	AssetStock  AssetClass = "STK"    // Stocks
	AssetWar    AssetClass = "WAR"    // Warrants
	AssetMargin AssetClass = "MRGN"   // Margin
	AssetCLP    AssetClass = "CLP"    // Certificate of limited partnership
	AssetCrypto AssetClass = "CRYPTO" // Cryptocurrencies
)

// AccountType is an account classification (e.g., "INDIVIDUAL", "JOINT").
type AccountType string

func (a AccountType) String() string { return string(a) }
