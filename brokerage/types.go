package brokerage

import "fmt"

// BarUnit represents the time unit for a history bar size.
type BarUnit int

const (
	BarSeconds BarUnit = iota + 1
	BarMinutes
	BarHours
	BarDays
	BarWeeks
	BarMonths
)

// Valid reports whether u is a known BarUnit constant.
func (u BarUnit) Valid() bool {
	return u >= BarSeconds && u <= BarMonths
}

func (u BarUnit) String() string {
	switch u {
	case BarSeconds:
		return "S"
	case BarMinutes:
		return "min"
	case BarHours:
		return "h"
	case BarDays:
		return "d"
	case BarWeeks:
		return "w"
	case BarMonths:
		return "m"
	default:
		return "?"
	}
}

// BarSize represents a history bar width (e.g., 5 minutes, 1 hour).
type BarSize struct {
	Count int
	Unit  BarUnit
}

// Bar creates a BarSize with the given count and unit.
// Returns an error if count is not positive or unit is not a known value.
func Bar(count int, unit BarUnit) (BarSize, error) {
	if count <= 0 {
		return BarSize{}, ErrInvalidCountValue(count)
	}
	if !unit.Valid() {
		return BarSize{}, ErrInvalidUnitValue(int(unit))
	}
	return BarSize{Count: count, Unit: unit}, nil
}

func (b BarSize) String() string { return fmt.Sprintf("%d%s", b.Count, b.Unit) }

// PeriodUnit represents the time unit for a history time period.
type PeriodUnit int

const (
	PeriodMinutes PeriodUnit = iota + 1
	PeriodHours
	PeriodDays
	PeriodWeeks
	PeriodMonths
	PeriodYears
)

// Valid reports whether u is a known PeriodUnit constant.
func (u PeriodUnit) Valid() bool {
	return u >= PeriodMinutes && u <= PeriodYears
}

func (u PeriodUnit) String() string {
	switch u {
	case PeriodMinutes:
		return "min"
	case PeriodHours:
		return "h"
	case PeriodDays:
		return "d"
	case PeriodWeeks:
		return "w"
	case PeriodMonths:
		return "m"
	case PeriodYears:
		return "y"
	default:
		return "?"
	}
}

// TimePeriod represents a history time range (e.g., 6 days, 1 year).
type TimePeriod struct {
	Count int
	Unit  PeriodUnit
}

// Period creates a TimePeriod with the given count and unit.
// Returns an error if count is not positive or unit is not a known value.
func Period(count int, unit PeriodUnit) (TimePeriod, error) {
	if count <= 0 {
		return TimePeriod{}, ErrInvalidCountValue(count)
	}
	if !unit.Valid() {
		return TimePeriod{}, ErrInvalidUnitValue(int(unit))
	}
	return TimePeriod{Count: count, Unit: unit}, nil
}

func (p TimePeriod) String() string { return fmt.Sprintf("%d%s", p.Count, p.Unit) }

// SnapshotField identifies a market data field in a snapshot request.
// Use the predefined Field* constants, or cast any numeric code via SnapshotField("code").
type SnapshotField string

func (f SnapshotField) String() string { return string(f) }

// Price fields.
const (
	FieldLast       SnapshotField = "31"   // Last traded price
	FieldBid        SnapshotField = "84"   // Highest bid
	FieldAsk        SnapshotField = "86"   // Lowest offer
	FieldHigh       SnapshotField = "70"   // Current day high
	FieldLow        SnapshotField = "71"   // Current day low
	FieldOpen       SnapshotField = "7295" // Today's opening price
	FieldClose      SnapshotField = "7296" // Today's closing price
	FieldPriorClose SnapshotField = "7741" // Yesterday's closing price
	FieldMark       SnapshotField = "7635" // Mark price
)

// Size fields.
const (
	FieldBidSize    SnapshotField = "88"   // Contracts/shares at bid
	FieldAskSize    SnapshotField = "85"   // Contracts/shares at ask
	FieldLastSize   SnapshotField = "7059" // Units traded at last price
	FieldVolume     SnapshotField = "87"   // Day volume (formatted)
	FieldVolumeLong SnapshotField = "7762" // Day volume (high precision)
)

// Identity fields.
const (
	FieldSymbol        SnapshotField = "55"   // Ticker symbol
	FieldText          SnapshotField = "58"   // Text
	FieldCompanyName   SnapshotField = "7051" // Company name
	FieldConID         SnapshotField = "6008" // Contract identifier
	FieldContractDesc  SnapshotField = "7219" // Contract description
	FieldContractDesc2 SnapshotField = "7220" // Contract description (alt)
	FieldExchange      SnapshotField = "6004" // Exchange
	FieldListingExch   SnapshotField = "7221" // Listing exchange
	FieldSecType       SnapshotField = "6070" // Security/asset type
	FieldMonths        SnapshotField = "6072" // Months
	FieldRegularExpiry SnapshotField = "6073" // Regular expiry
	FieldConIDExchange SnapshotField = "7094" // ConID + Exchange
)

// Change fields.
const (
	FieldChange          SnapshotField = "82"   // Change from prior close
	FieldChangePct       SnapshotField = "83"   // Change % from prior close
	FieldChangeSinceOpen SnapshotField = "7682" // Change since open
)

// Position and PnL fields.
const (
	FieldMarketValue        SnapshotField = "73"   // Current position market value
	FieldAvgPrice           SnapshotField = "74"   // Average position price
	FieldUnrealizedPnL      SnapshotField = "75"   // Unrealized PnL
	FieldFormattedPosition  SnapshotField = "76"   // Formatted position
	FieldFormattedUnrlzdPnL SnapshotField = "77"   // Formatted unrealized PnL
	FieldDailyPnL           SnapshotField = "78"   // Daily PnL since prior close
	FieldRealizedPnL        SnapshotField = "79"   // Realized PnL
	FieldUnrealizedPnLPct   SnapshotField = "80"   // Unrealized PnL %
	FieldCostBasis          SnapshotField = "7292" // Position × avg price × multiplier
	FieldPctOfMarkValue     SnapshotField = "7639" // % of account mark value
	FieldDailyPnLRaw        SnapshotField = "7920" // Daily PnL (raw)
	FieldCostBasisRaw       SnapshotField = "7921" // Cost basis (raw)
)

// Options and Greeks fields.
const (
	FieldDelta             SnapshotField = "7308" // Option delta
	FieldGamma             SnapshotField = "7309" // Option gamma
	FieldTheta             SnapshotField = "7310" // Option theta
	FieldVega              SnapshotField = "7311" // Option vega
	FieldImpliedVol        SnapshotField = "7633" // Implied volatility % (specific strike)
	FieldOptImpliedVol     SnapshotField = "7283" // Implied vol % (from underlying, 30-day)
	FieldOptOpenInterest   SnapshotField = "7638" // Option open interest
	FieldOptVolume         SnapshotField = "7089" // Option volume
	FieldOptVolChangePct   SnapshotField = "7607" // Option volume change %
	FieldPutCallInterest   SnapshotField = "7085" // Put/call open interest ratio
	FieldPutCallVolume     SnapshotField = "7086" // Put/call volume ratio
	FieldPutCallRatio      SnapshotField = "7285" // Put/call ratio
	FieldImpliedHistRatio  SnapshotField = "7084" // Implied vol / historical vol %
	FieldHistVol           SnapshotField = "7087" // 30-day historical volatility %
	FieldHistVolClose      SnapshotField = "7088" // Historical vol % (close-based)
	FieldHistVolDeprecated SnapshotField = "7284" // Historical vol % (deprecated)
)

// Analytics fields.
const (
	FieldWeek52High   SnapshotField = "7293" // 52-week high
	FieldWeek52Low    SnapshotField = "7294" // 52-week low
	FieldAvgVolume    SnapshotField = "7282" // 90-day average daily volume
	FieldEMA200       SnapshotField = "7674" // Exponential moving average (200)
	FieldEMA100       SnapshotField = "7675" // Exponential moving average (100)
	FieldEMA50        SnapshotField = "7676" // Exponential moving average (50)
	FieldEMA20        SnapshotField = "7677" // Exponential moving average (20)
	FieldPriceEMA200  SnapshotField = "7678" // Price/EMA(200) ratio − 1
	FieldPriceEMA100  SnapshotField = "7679" // Price/EMA(100) ratio − 1
	FieldPriceEMA50   SnapshotField = "7724" // Price/EMA(50) ratio − 1
	FieldPriceEMA20   SnapshotField = "7681" // Price/EMA(20) ratio − 1
	FieldMorningstar  SnapshotField = "7655" // Morningstar rating
	FieldDividends    SnapshotField = "7671" // Expected dividends (next 12 months)
	FieldDividendsTTM SnapshotField = "7672" // Dividends (trailing 12 months)
	FieldIndustry     SnapshotField = "7280" // Industry classification
	FieldCategory     SnapshotField = "7281" // Industry sub-category
)

// Exchange routing fields.
const (
	FieldAskExch  SnapshotField = "7057" // Ask exchange(s)
	FieldBidExch  SnapshotField = "7068" // Bid exchange(s)
	FieldLastExch SnapshotField = "7058" // Last trade exchange(s)
)

// Trading permission and shortability fields.
const (
	FieldCanBeTraded     SnapshotField = "7184" // 1 if tradeable, 0 otherwise
	FieldHasTradingPerms SnapshotField = "7768" // 1 if user has trading permissions
	FieldShortable       SnapshotField = "7644" // Short-sell difficulty level
	FieldShortableShares SnapshotField = "7636" // Shares available to short
	FieldFeeRate         SnapshotField = "7637" // Borrow fee rate
	FieldRebateRate      SnapshotField = "7943" // Rebate rate
)

// Bond fields.
const (
	FieldLastYield       SnapshotField = "7698" // Last yield
	FieldBidYield        SnapshotField = "7699" // Bid yield
	FieldAskYield        SnapshotField = "7720" // Ask yield
	FieldOrgType         SnapshotField = "7704" // Organization type
	FieldDebtClass       SnapshotField = "7705" // Debt class
	FieldRatings         SnapshotField = "7706" // Bond ratings
	FieldBondStateCode   SnapshotField = "7707" // Bond state code
	FieldBondType        SnapshotField = "7708" // Bond type
	FieldLastTradingDate SnapshotField = "7714" // Last trading date
	FieldIssueDate       SnapshotField = "7715" // Issue date
)

// Upcoming/recent event fields.
const (
	FieldUpcomingEvent     SnapshotField = "7683" // Next major event
	FieldUpcomingEventDate SnapshotField = "7684" // Next major event date
	FieldUpcomingAnalyst   SnapshotField = "7685" // Next analyst meeting
	FieldUpcomingEarnings  SnapshotField = "7686" // Next earnings
	FieldUpcomingMisc      SnapshotField = "7687" // Next misc event
	FieldRecentAnalyst     SnapshotField = "7688" // Most recent analyst meeting
	FieldRecentEarnings    SnapshotField = "7689" // Most recent earnings
	FieldRecentMisc        SnapshotField = "7690" // Most recent misc event
)

// Probability fields.
const (
	FieldProbMaxReturn  SnapshotField = "7694" // Probability of max return
	FieldBreakEven      SnapshotField = "7695" // Break-even points
	FieldSPXDelta       SnapshotField = "7696" // Beta-weighted delta (SPX)
	FieldProbMaxReturn2 SnapshotField = "7700" // Probability of max return (alt)
	FieldProbMaxLoss    SnapshotField = "7702" // Probability of max loss
	FieldProfitProb     SnapshotField = "7703" // Probability of any gain
)

// Futures fields.
const (
	FieldFuturesOpenInterest SnapshotField = "7697" // Futures open interest
)

// Metadata fields.
const (
	FieldMarker          SnapshotField = "6119" // Market data delivery marker
	FieldUnderlyingConID SnapshotField = "6457" // Underlying contract ID
	FieldServiceParams   SnapshotField = "6508" // Service parameters
	FieldMDAvailability  SnapshotField = "6509" // Market data availability
)

// Convenience field sets for common snapshot requests.
var (
	// FieldsQuote is a standard quote: symbol, last, bid/ask, high/low, volume, change.
	FieldsQuote = []SnapshotField{
		FieldSymbol, FieldCompanyName,
		FieldLast, FieldBid, FieldAsk,
		FieldHigh, FieldLow, FieldOpen, FieldClose, FieldPriorClose,
		FieldVolume, FieldChange, FieldChangePct,
	}

	// FieldsGreeks contains option greeks and implied volatility.
	FieldsGreeks = []SnapshotField{
		FieldDelta, FieldGamma, FieldTheta, FieldVega,
		FieldImpliedVol, FieldOptImpliedVol,
	}

	// FieldsPnL contains position-level profit and loss fields.
	FieldsPnL = []SnapshotField{
		FieldMarketValue, FieldAvgPrice,
		FieldUnrealizedPnL, FieldUnrealizedPnLPct,
		FieldRealizedPnL, FieldDailyPnL,
		FieldCostBasis,
	}

	// FieldsBond contains bond-specific fields.
	FieldsBond = []SnapshotField{
		FieldLastYield, FieldBidYield, FieldAskYield,
		FieldRatings, FieldBondType, FieldDebtClass,
		FieldIssueDate, FieldLastTradingDate,
	}
)
