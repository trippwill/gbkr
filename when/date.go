package when

import (
	"database/sql/driver"
	"encoding/xml"
	"fmt"
	"strings"
	"time"
)

// Accepted date parse formats, tried in order.
var dateFormats = []string{
	"2006-01-02",
	"20060102",
}

// Date represents a calendar date with no time component.
// Internally it stores a [time.Time] normalized to midnight UTC.
// The zero value is valid and represents the zero time (use IsZero to check).
type Date struct {
	t time.Time
}

// NewDate creates a Date from year, month, and day components.
func NewDate(year int, month time.Month, day int) Date {
	return Date{t: time.Date(year, month, day, 0, 0, 0, 0, time.UTC)}
}

// DateFromTime creates a Date by truncating t to midnight UTC.
func DateFromTime(t time.Time) Date {
	u := t.UTC()
	return Date{t: time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)}
}

// ParseDate parses a date string. Accepted formats:
//   - "YYYY-MM-DD" (Flex XML, ISO 8601 date)
//   - "YYYYMMDD" (REST JSON)
//   - "YYYY-MM-DD;HH:MM:SS" (Flex timestamp — time component is dropped)
//   - "YYYYMMDD-HH:MM:SS" (REST timestamp — time component is dropped)
//
// Returns an error if the string does not match any format.
func ParseDate(s string) (Date, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return Date{}, fmt.Errorf("%w: empty string", ErrInvalidDate)
	}

	// Strip time component if present.
	dateStr := stripTime(s)

	for _, layout := range dateFormats {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return Date{t: t}, nil
		}
	}
	return Date{}, fmt.Errorf("%w: %q", ErrInvalidDate, s)
}

// stripTime removes the time portion from compound date-time strings.
func stripTime(s string) string {
	if before, _, found := strings.Cut(s, ";"); found {
		return before
	}
	if i := strings.IndexByte(s, '-'); i >= 0 {
		// "YYYYMMDD-HH:MM:SS" — the dash is after 8 digits.
		// vs "YYYY-MM-DD" — the dash is at position 4.
		// Only strip if the part before the dash is 8 digits (YYYYMMDD).
		if i == 8 {
			return s[:8]
		}
	}
	return s
}

// Time returns the underlying time.Time (midnight UTC).
func (d Date) Time() time.Time { return d.t }

// Year returns the year.
func (d Date) Year() int { return d.t.Year() }

// Month returns the month.
func (d Date) Month() time.Month { return d.t.Month() }

// Day returns the day of the month.
func (d Date) Day() int { return d.t.Day() }

// IsZero reports whether d represents the zero time.
func (d Date) IsZero() bool { return d.t.IsZero() }

// Equal reports whether d and other represent the same date.
func (d Date) Equal(other Date) bool { return d.t.Equal(other.t) }

// Before reports whether d is before other.
func (d Date) Before(other Date) bool { return d.t.Before(other.t) }

// After reports whether d is after other.
func (d Date) After(other Date) bool { return d.t.After(other.t) }

// String returns the date in "YYYY-MM-DD" format.
func (d Date) String() string {
	if d.IsZero() {
		return ""
	}
	return d.t.Format("2006-01-02")
}

// MarshalJSON returns the JSON encoding as a "YYYY-MM-DD" quoted string.
func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte(`""`), nil
	}
	return []byte(`"` + d.String() + `"`), nil
}

// UnmarshalJSON parses a JSON string into a Date.
func (d *Date) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		*d = Date{}
		return nil
	}
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	parsed, err := ParseDate(s)
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// MarshalText implements [encoding.TextMarshaler].
// Used for XML attributes and map keys.
func (d Date) MarshalText() ([]byte, error) {
	return []byte(d.String()), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
// Used for XML attributes and map keys.
func (d *Date) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		*d = Date{}
		return nil
	}
	parsed, err := ParseDate(string(data))
	if err != nil {
		return err
	}
	*d = parsed
	return nil
}

// UnmarshalXMLAttr implements [xml.UnmarshalerAttr] for Flex XML attributes.
func (d *Date) UnmarshalXMLAttr(attr xml.Attr) error {
	return d.UnmarshalText([]byte(attr.Value))
}

// Value implements [database/sql/driver.Valuer].
// Returns the date as a "YYYY-MM-DD" string for SQLite TEXT storage.
func (d Date) Value() (driver.Value, error) {
	if d.IsZero() {
		return nil, nil
	}
	return d.String(), nil
}

// Scan implements [database/sql.Scanner].
// Accepts string, []byte, and time.Time inputs.
func (d *Date) Scan(src any) error {
	if src == nil {
		*d = Date{}
		return nil
	}
	switch v := src.(type) {
	case string:
		parsed, err := ParseDate(v)
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case []byte:
		parsed, err := ParseDate(string(v))
		if err != nil {
			return err
		}
		*d = parsed
		return nil
	case time.Time:
		*d = DateFromTime(v)
		return nil
	default:
		return fmt.Errorf("%w: %T", ErrUnsupportedType, src)
	}
}
