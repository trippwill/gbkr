package when

import (
	"database/sql/driver"
	"encoding/xml"
)

// NullDate represents a [Date] that may be absent.
// JSON null, empty strings, and SQL NULL map to Valid=false.
type NullDate struct {
	Date  Date
	Valid bool
}

// String implements [fmt.Stringer].
func (nd NullDate) String() string {
	if !nd.Valid {
		return ""
	}
	return nd.Date.String()
}

// MarshalJSON returns JSON null when !Valid, otherwise delegates to [Date.MarshalJSON].
func (nd NullDate) MarshalJSON() ([]byte, error) {
	if !nd.Valid {
		return []byte("null"), nil
	}
	return nd.Date.MarshalJSON()
}

// UnmarshalJSON handles JSON null and empty string → Valid=false,
// otherwise delegates to [Date.UnmarshalJSON].
func (nd *NullDate) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		nd.Valid = false
		nd.Date = Date{}
		return nil
	}
	var parsed Date
	if err := parsed.UnmarshalJSON(data); err != nil {
		nd.Valid = false
		nd.Date = Date{}
		return err
	}
	nd.Valid = true
	nd.Date = parsed
	return nil
}

// MarshalText returns an empty byte slice when !Valid.
func (nd NullDate) MarshalText() ([]byte, error) {
	if !nd.Valid {
		return []byte{}, nil
	}
	return nd.Date.MarshalText()
}

// UnmarshalText handles empty input → Valid=false.
func (nd *NullDate) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		nd.Valid = false
		nd.Date = Date{}
		return nil
	}
	if err := nd.Date.UnmarshalText(data); err != nil {
		nd.Valid = false
		nd.Date = Date{}
		return err
	}
	nd.Valid = true
	return nil
}

// UnmarshalXMLAttr implements [xml.UnmarshalerAttr].
func (nd *NullDate) UnmarshalXMLAttr(attr xml.Attr) error {
	return nd.UnmarshalText([]byte(attr.Value))
}

// Value implements [database/sql/driver.Valuer]. Returns nil when !Valid.
func (nd NullDate) Value() (driver.Value, error) {
	if !nd.Valid {
		return nil, nil
	}
	return nd.Date.Value()
}

// Scan implements [database/sql.Scanner]. Handles nil → Valid=false.
func (nd *NullDate) Scan(src any) error {
	if src == nil {
		nd.Valid = false
		nd.Date = Date{}
		return nil
	}
	if err := nd.Date.Scan(src); err != nil {
		nd.Valid = false
		nd.Date = Date{}
		return err
	}
	nd.Valid = true
	return nil
}

// NullDateTime represents a [DateTime] that may be absent.
// JSON null, empty strings, and SQL NULL map to Valid=false.
type NullDateTime struct {
	DateTime DateTime
	Valid    bool
}

// String implements [fmt.Stringer].
func (ndt NullDateTime) String() string {
	if !ndt.Valid {
		return ""
	}
	return ndt.DateTime.String()
}

// MarshalJSON returns JSON null when !Valid, otherwise delegates to [DateTime.MarshalJSON].
func (ndt NullDateTime) MarshalJSON() ([]byte, error) {
	if !ndt.Valid {
		return []byte("null"), nil
	}
	return ndt.DateTime.MarshalJSON()
}

// UnmarshalJSON handles JSON null and empty string → Valid=false,
// otherwise delegates to [DateTime.UnmarshalJSON].
func (ndt *NullDateTime) UnmarshalJSON(data []byte) error {
	s := string(data)
	if s == "null" || s == `""` {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return nil
	}
	var parsed DateTime
	if err := parsed.UnmarshalJSON(data); err != nil {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return err
	}
	ndt.Valid = true
	ndt.DateTime = parsed
	return nil
}

// MarshalText returns an empty byte slice when !Valid.
func (ndt NullDateTime) MarshalText() ([]byte, error) {
	if !ndt.Valid {
		return []byte{}, nil
	}
	return ndt.DateTime.MarshalText()
}

// UnmarshalText handles empty input → Valid=false.
func (ndt *NullDateTime) UnmarshalText(data []byte) error {
	if len(data) == 0 {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return nil
	}
	if err := ndt.DateTime.UnmarshalText(data); err != nil {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return err
	}
	ndt.Valid = true
	return nil
}

// UnmarshalXMLAttr implements [xml.UnmarshalerAttr].
func (ndt *NullDateTime) UnmarshalXMLAttr(attr xml.Attr) error {
	return ndt.UnmarshalText([]byte(attr.Value))
}

// Value implements [database/sql/driver.Valuer]. Returns nil when !Valid.
func (ndt NullDateTime) Value() (driver.Value, error) {
	if !ndt.Valid {
		return nil, nil
	}
	return ndt.DateTime.Value()
}

// Scan implements [database/sql.Scanner]. Handles nil → Valid=false.
func (ndt *NullDateTime) Scan(src any) error {
	if src == nil {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return nil
	}
	if err := ndt.DateTime.Scan(src); err != nil {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return err
	}
	ndt.Valid = true
	return nil
}
