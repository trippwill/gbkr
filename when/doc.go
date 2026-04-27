// Package when provides [Date], [NullDate], [DateTime], and [NullDateTime]
// types for strongly typed temporal values in IBKR API responses.
//
// IBKR uses at least five different date/time wire formats across its REST,
// WebSocket, and Flex XML APIs. These types absorb all formats at the
// wire→domain boundary so consumers get standard [time.Time] values via
// the Time() accessor without knowing the source format.
//
// # Date
//
// [Date] represents a calendar date with no time component. Internally it
// stores a [time.Time] normalized to midnight UTC. It accepts multiple
// input formats:
//
//	d := when.ParseDate("2026-01-15")   // Flex XML format
//	d := when.ParseDate("20260115")     // REST JSON format
//	d := when.NewDate(2026, 1, 15)      // explicit
//
// # DateTime
//
// [DateTime] represents a point in time with full timestamp precision.
// It accepts string formats and epoch timestamps:
//
//	dt := when.ParseDateTime("20231211-18:00:49")        // REST trade time
//	dt := when.DateTimeFromEpochMs(1702317649000)        // millisecond epoch
//	dt := when.DateTimeFromEpochSec(1702317649)          // second epoch
//
// # Nullable Variants
//
// [NullDate] and [NullDateTime] follow the [github.com/trippwill/gbkr/num.NullNum]
// pattern: a Valid flag distinguishes absent values from zero. JSON null,
// empty strings, and SQL NULL all map to Valid=false.
package when
