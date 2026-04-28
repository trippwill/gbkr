package when

import (
	"database/sql/driver"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Accepted datetime parse formats, tried in order.
var dateTimeFormats = []string{
	time.RFC3339,
	"2006-01-02T15:04:05",
	"2006-01-02;15:04:05",
	"20060102;150405",
	"20060102-15:04:05",
	"20060102 15:04:05",
}

// rfc3339Ms is RFC 3339 with fixed millisecond precision.
const rfc3339Ms = "2006-01-02T15:04:05.000Z07:00"

// Values at or above this are treated as milliseconds.
// 946684800000 = 2000-01-01T00:00:00Z in ms (year ~31939 in seconds).
const epochMsThreshold int64 = 946_684_800_000

// DateTime represents a point in time with full timestamp precision.
// The zero value is valid and represents the zero time (use IsZero to check).
type DateTime struct {
	t time.Time
}

// DateTimeFromTime creates a DateTime from a [time.Time].
func DateTimeFromTime(t time.Time) DateTime {
	return DateTime{t: t.UTC()}
}

// DateTimeFromEpochMs creates a DateTime from a Unix epoch in milliseconds.
func DateTimeFromEpochMs(ms int64) DateTime {
	return DateTime{t: time.UnixMilli(ms).UTC()}
}

// DateTimeFromEpochSec creates a DateTime from a Unix epoch in seconds.
func DateTimeFromEpochSec(sec int64) DateTime {
	return DateTime{t: time.Unix(sec, 0).UTC()}
}

// DateTimeFromEpoch creates a DateTime from a Unix epoch value,
// automatically detecting whether it is in seconds or milliseconds.
// Values with absolute magnitude >= epochMsThreshold are treated as
// milliseconds; smaller values are treated as seconds.
func DateTimeFromEpoch(epoch int64) DateTime {
	if epoch >= epochMsThreshold || epoch <= -epochMsThreshold {
		return DateTimeFromEpochMs(epoch)
	}
	return DateTimeFromEpochSec(epoch)
}

// ParseDateTime parses a datetime string. Accepted formats:
//   - "YYYYMMDD-HH:MM:SS" (REST JSON trade time)
//   - "YYYYMMDD HH:MM:SS" (streaming history bars)
//   - "YYYY-MM-DD;HH:MM:SS" (Flex XML timestamp)
//   - "YYYY-MM-DDTHH:MM:SSZ" (ISO 8601 / RFC 3339)
//   - "YYYY-MM-DDTHH:MM:SS" (ISO 8601 without timezone)
//
// Returns an error if the string does not match any format.
func ParseDateTime(s string) (DateTime, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return DateTime{}, fmt.Errorf("%w: empty string", ErrInvalidDateTime)
	}

	for _, layout := range dateTimeFormats {
		if t, err := time.Parse(layout, s); err == nil {
			return DateTime{t: t.UTC()}, nil
		}
	}
	return DateTime{}, fmt.Errorf("%w: %q", ErrInvalidDateTime, s)
}

// Time returns the underlying time.Time (UTC).
func (dt DateTime) Time() time.Time { return dt.t }

// Date returns a [Date] by truncating to midnight UTC.
func (dt DateTime) Date() Date { return DateFromTime(dt.t) }

// IsZero reports whether dt represents the zero time.
func (dt DateTime) IsZero() bool { return dt.t.IsZero() }

// Equal reports whether dt and other represent the same instant.
func (dt DateTime) Equal(other DateTime) bool { return dt.t.Equal(other.t) }

// Before reports whether dt is before other.
func (dt DateTime) Before(other DateTime) bool { return dt.t.Before(other.t) }

// After reports whether dt is after other.
func (dt DateTime) After(other DateTime) bool { return dt.t.After(other.t) }

// String returns the datetime in RFC 3339 format with millisecond precision.
func (dt DateTime) String() string {
	if dt.IsZero() {
		return ""
	}
	return dt.t.Format(rfc3339Ms)
}

// MarshalJSON returns the JSON encoding as an RFC 3339 quoted string.
func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + dt.String() + `"`), nil
}

// UnmarshalJSON parses a JSON value into a DateTime.
// Accepts quoted strings (all supported formats) and bare integers
// (epoch milliseconds, with heuristic to distinguish from seconds).
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		*dt = DateTime{}
		return nil
	}

	// Try bare integer (epoch timestamp).
	if len(s) > 0 && ((s[0] >= '0' && s[0] <= '9') || s[0] == '-') {
		epoch, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			if epoch >= epochMsThreshold || epoch <= -epochMsThreshold {
				*dt = DateTimeFromEpochMs(epoch)
			} else {
				*dt = DateTimeFromEpochSec(epoch)
			}
			return nil
		}
	}

	// Unquote string.
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}

	// Try epoch in string form.
	if epoch, err := strconv.ParseInt(s, 10, 64); err == nil {
		if epoch >= epochMsThreshold || epoch <= -epochMsThreshold {
			*dt = DateTimeFromEpochMs(epoch)
		} else {
			*dt = DateTimeFromEpochSec(epoch)
		}
		return nil
	}

	parsed, err := ParseDateTime(s)
	if err != nil {
		return err
	}
	*dt = parsed
	return nil
}

// MarshalText implements [encoding.TextMarshaler].
func (dt DateTime) MarshalText() ([]byte, error) {
	return []byte(dt.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (dt *DateTime) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*dt = DateTime{}
		return nil
	}
	parsed, err := ParseDateTime(string(data))
	if err != nil {
		return err
	}
	*dt = parsed
	return nil
}

// UnmarshalXMLAttr implements [xml.UnmarshalerAttr] for Flex XML attributes.
func (dt *DateTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return dt.UnmarshalText([]byte(attr.Value))
}

// Value implements [database/sql/driver.Valuer].
// Returns the datetime as an RFC 3339 string for SQLite TEXT storage.
func (dt DateTime) Value() (driver.Value, error) {
	if dt.IsZero() {
		return nil, nil
	}
	return dt.String(), nil
}

// Scan implements [database/sql.Scanner].
// Accepts string, []byte, int64, and time.Time inputs.
func (dt *DateTime) Scan(src any) error {
	if src == nil {
		*dt = DateTime{}
		return nil
	}
	switch v := src.(type) {
	case string:
		parsed, err := ParseDateTime(v)
		if err != nil {
			return err
		}
		*dt = parsed
		return nil
	case []byte:
		parsed, err := ParseDateTime(string(v))
		if err != nil {
			return err
		}
		*dt = parsed
		return nil
	case int64:
		if v >= epochMsThreshold || v <= -epochMsThreshold {
			*dt = DateTimeFromEpochMs(v)
		} else {
			*dt = DateTimeFromEpochSec(v)
		}
		return nil
	case time.Time:
		*dt = DateTimeFromTime(v)
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedType, src)
	}
}

// Ensure interface compliance at compile time.
var (
	_ json.Marshaler      = Date{}
	_ json.Unmarshaler    = (*Date)(nil)
	_ xml.UnmarshalerAttr = (*Date)(nil)
	_ driver.Valuer       = Date{}
	_ json.Marshaler      = DateTime{}
	_ json.Unmarshaler    = (*DateTime)(nil)
	_ xml.UnmarshalerAttr = (*DateTime)(nil)
	_ driver.Valuer       = DateTime{}
	_ json.Marshaler      = NullDate{}
	_ json.Unmarshaler    = (*NullDate)(nil)
	_ xml.UnmarshalerAttr = (*NullDate)(nil)
	_ driver.Valuer       = NullDate{}
	_ json.Marshaler      = NullDateTime{}
	_ json.Unmarshaler    = (*NullDateTime)(nil)
	_ xml.UnmarshalerAttr = (*NullDateTime)(nil)
	_ driver.Valuer       = NullDateTime{}
)
