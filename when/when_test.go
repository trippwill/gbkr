package when

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"testing"
	"time"
)

// --- Date ---

func TestNewDate(t *testing.T) {
	d := NewDate(2026, time.January, 15)
	if d.Year() != 2026 || d.Month() != time.January || d.Day() != 15 {
		t.Errorf("NewDate = %v, want 2026-01-15", d)
	}
	if d.IsZero() {
		t.Error("non-zero date reports IsZero")
	}
}

func TestDateFromTime(t *testing.T) {
	tm := time.Date(2026, 3, 20, 14, 30, 0, 0, time.FixedZone("EST", -5*3600))
	d := DateFromTime(tm)
	if d.Year() != 2026 || d.Month() != time.March || d.Day() != 20 {
		t.Errorf("DateFromTime = %v, want 2026-03-20", d)
	}
	if d.Time().Hour() != 0 || d.Time().Minute() != 0 {
		t.Errorf("DateFromTime should be midnight UTC, got %v", d.Time())
	}
}

func TestParseDateFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Date
	}{
		{"ISO 8601", "2026-01-15", NewDate(2026, 1, 15)},
		{"compact", "20260115", NewDate(2026, 1, 15)},
		{"flex timestamp", "2026-03-20;19:30:00", NewDate(2026, 3, 20)},
		{"REST timestamp", "20231211-18:00:49", NewDate(2023, 12, 11)},
		{"whitespace", "  2026-01-15  ", NewDate(2026, 1, 15)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDate(tt.input)
			if err != nil {
				t.Fatalf("ParseDate(%q) error: %v", tt.input, err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseDate(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseDateErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"whitespace only", "   "},
		{"garbage", "not-a-date"},
		{"incomplete", "2026-01"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDate(tt.input)
			if err == nil {
				t.Fatalf("ParseDate(%q) should fail", tt.input)
			}
			if !errors.Is(err, ErrInvalidDate) {
				t.Errorf("error = %v, want ErrInvalidDate", err)
			}
		})
	}
}

func TestDateZeroValue(t *testing.T) {
	var d Date
	if !d.IsZero() {
		t.Error("zero value should be IsZero")
	}
	if d.String() != "" {
		t.Errorf("zero String() = %q, want empty", d.String())
	}
}

func TestDateComparisons(t *testing.T) {
	a := NewDate(2026, 1, 15)
	b := NewDate(2026, 3, 20)
	same := NewDate(2026, 1, 15)

	if !a.Before(b) {
		t.Error("a should be before b")
	}
	if !b.After(a) {
		t.Error("b should be after a")
	}
	if !a.Equal(same) {
		t.Error("a should equal same")
	}
}

func TestDateString(t *testing.T) {
	d := NewDate(2026, 1, 5)
	if got := d.String(); got != "2026-01-05" {
		t.Errorf("String() = %q, want %q", got, "2026-01-05")
	}
}

func TestDateJSONRoundtrip(t *testing.T) {
	type rec struct {
		D Date `json:"d"`
	}
	orig := rec{D: NewDate(2026, 1, 15)}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `{"d":"2026-01-15"}` {
		t.Errorf("Marshal = %s", data)
	}

	var restored rec
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !restored.D.Equal(orig.D) {
		t.Errorf("roundtrip: got %v, want %v", restored.D, orig.D)
	}
}

func TestDateUnmarshalJSON_Formats(t *testing.T) {
	tests := []struct {
		name string
		json string
		want Date
	}{
		{"ISO", `"2026-01-15"`, NewDate(2026, 1, 15)},
		{"compact", `"20260115"`, NewDate(2026, 1, 15)},
		{"flex timestamp", `"2026-03-20;19:30:00"`, NewDate(2026, 3, 20)},
		{"REST timestamp", `"20231211-18:00:49"`, NewDate(2023, 12, 11)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d Date
			if err := json.Unmarshal([]byte(tt.json), &d); err != nil {
				t.Fatalf("Unmarshal(%s) error: %v", tt.json, err)
			}
			if !d.Equal(tt.want) {
				t.Errorf("Unmarshal(%s) = %v, want %v", tt.json, d, tt.want)
			}
		})
	}
}

func TestDateUnmarshalJSON_NullAndEmpty(t *testing.T) {
	var d Date

	if err := json.Unmarshal([]byte("null"), &d); err != nil {
		t.Fatal(err)
	}
	if !d.IsZero() {
		t.Error("null should produce zero date")
	}

	d = NewDate(2026, 1, 1)
	if err := json.Unmarshal([]byte(`""`), &d); err != nil {
		t.Fatal(err)
	}
	if !d.IsZero() {
		t.Error("empty string should produce zero date")
	}
}

func TestDateMarshalJSON_Zero(t *testing.T) {
	var d Date
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `""` {
		t.Errorf("Marshal(zero) = %s, want empty string", data)
	}
}

func TestDateTextRoundtrip(t *testing.T) {
	d := NewDate(2026, 1, 15)
	data, err := d.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "2026-01-15" {
		t.Errorf("MarshalText = %s", data)
	}

	var restored Date
	if err := restored.UnmarshalText(data); err != nil {
		t.Fatal(err)
	}
	if !restored.Equal(d) {
		t.Errorf("text roundtrip: got %v, want %v", restored, d)
	}
}

func TestDateTextEmpty(t *testing.T) {
	var d Date
	if err := d.UnmarshalText([]byte{}); err != nil {
		t.Fatal(err)
	}
	if !d.IsZero() {
		t.Error("empty text should produce zero date")
	}
}

func TestDateMapKey(t *testing.T) {
	m := map[Date]string{
		NewDate(2026, 1, 15): "MLK Day",
		NewDate(2026, 7, 4):  "Independence Day",
	}
	data, err := json.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	var restored map[Date]string
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored[NewDate(2026, 1, 15)] != "MLK Day" {
		t.Errorf("map key roundtrip failed: %v", restored)
	}
	if restored[NewDate(2026, 7, 4)] != "Independence Day" {
		t.Errorf("map key roundtrip failed: %v", restored)
	}
}

func TestDateXMLAttr(t *testing.T) {
	type elem struct {
		XMLName   xml.Name `xml:"trade"`
		TradeDate Date     `xml:"tradeDate,attr"`
	}
	input := `<trade tradeDate="2026-01-15"/>`
	var e elem
	if err := xml.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("XML unmarshal: %v", err)
	}
	if !e.TradeDate.Equal(NewDate(2026, 1, 15)) {
		t.Errorf("XML attr = %v, want 2026-01-15", e.TradeDate)
	}
}

func TestDateXMLAttr_Compact(t *testing.T) {
	type elem struct {
		XMLName xml.Name `xml:"trade"`
		Expiry  Date     `xml:"expiry,attr"`
	}
	input := `<trade expiry="20260115"/>`
	var e elem
	if err := xml.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("XML unmarshal: %v", err)
	}
	if !e.Expiry.Equal(NewDate(2026, 1, 15)) {
		t.Errorf("XML attr = %v, want 2026-01-15", e.Expiry)
	}
}

func TestDateSQLRoundtrip(t *testing.T) {
	d := NewDate(2026, 1, 15)
	v, err := d.Value()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("Value type = %T, want string", v)
	}
	if s != "2026-01-15" {
		t.Errorf("Value() = %q", s)
	}

	var restored Date
	if err := restored.Scan(s); err != nil {
		t.Fatal(err)
	}
	if !restored.Equal(d) {
		t.Errorf("SQL roundtrip: got %v, want %v", restored, d)
	}
}

func TestDateScanNil(t *testing.T) {
	var d Date
	if err := d.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if !d.IsZero() {
		t.Error("Scan(nil) should produce zero date")
	}
}

func TestDateScanBytes(t *testing.T) {
	var d Date
	if err := d.Scan([]byte("2026-01-15")); err != nil {
		t.Fatal(err)
	}
	if !d.Equal(NewDate(2026, 1, 15)) {
		t.Errorf("Scan(bytes) = %v", d)
	}
}

func TestDateScanTime(t *testing.T) {
	var d Date
	tm := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
	if err := d.Scan(tm); err != nil {
		t.Fatal(err)
	}
	if !d.Equal(NewDate(2026, 1, 15)) {
		t.Errorf("Scan(time) = %v", d)
	}
}

func TestDateScanUnsupported(t *testing.T) {
	var d Date
	err := d.Scan(42)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("error = %v, want ErrUnsupportedType", err)
	}
}

func TestDateValueZero(t *testing.T) {
	var d Date
	v, err := d.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("Value(zero) = %v, want nil", v)
	}
}

// --- DateTime ---

func TestDateTimeFromTime(t *testing.T) {
	tm := time.Date(2026, 3, 20, 14, 30, 45, 0, time.FixedZone("EST", -5*3600))
	dt := DateTimeFromTime(tm)
	if dt.Time().Hour() != 19 || dt.Time().Minute() != 30 {
		t.Errorf("DateTimeFromTime should normalize to UTC, got %v", dt.Time())
	}
}

func TestDateTimeFromEpoch(t *testing.T) {
	ms := int64(1702317649000)
	dt := DateTimeFromEpochMs(ms)
	if dt.IsZero() {
		t.Error("epoch ms should not be zero")
	}
	if dt.Time().UnixMilli() != ms {
		t.Errorf("epoch ms roundtrip: got %d, want %d", dt.Time().UnixMilli(), ms)
	}

	sec := int64(1702317649)
	dt2 := DateTimeFromEpochSec(sec)
	if dt2.Time().Unix() != sec {
		t.Errorf("epoch sec roundtrip: got %d, want %d", dt2.Time().Unix(), sec)
	}
}

func TestParseDateTimeFormats(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Time
	}{
		{"REST trade", "20231211-18:00:49", time.Date(2023, 12, 11, 18, 0, 49, 0, time.UTC)},
		{"Flex timestamp", "2026-03-20;19:30:00", time.Date(2026, 3, 20, 19, 30, 0, 0, time.UTC)},
		{"Flex compact", "20260428;153347", time.Date(2026, 4, 28, 15, 33, 47, 0, time.UTC)},
		{"RFC 3339", "2026-03-20T19:30:00Z", time.Date(2026, 3, 20, 19, 30, 0, 0, time.UTC)},
		{"ISO no TZ", "2026-03-20T19:30:00", time.Date(2026, 3, 20, 19, 30, 0, 0, time.UTC)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDateTime(tt.input)
			if err != nil {
				t.Fatalf("ParseDateTime(%q) error: %v", tt.input, err)
			}
			if !got.Time().Equal(tt.want) {
				t.Errorf("ParseDateTime(%q) = %v, want %v", tt.input, got.Time(), tt.want)
			}
		})
	}
}

func TestParseDateTimeErrors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"whitespace", "   "},
		{"garbage", "not-a-datetime"},
		{"date only", "2026-01-15"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseDateTime(tt.input)
			if err == nil {
				t.Fatalf("ParseDateTime(%q) should fail", tt.input)
			}
			if !errors.Is(err, ErrInvalidDateTime) {
				t.Errorf("error = %v, want ErrInvalidDateTime", err)
			}
		})
	}
}

func TestDateTimeDate(t *testing.T) {
	dt := DateTimeFromTime(time.Date(2026, 3, 20, 14, 30, 0, 0, time.UTC))
	d := dt.Date()
	if !d.Equal(NewDate(2026, 3, 20)) {
		t.Errorf("Date() = %v, want 2026-03-20", d)
	}
}

func TestDateTimeZeroValue(t *testing.T) {
	var dt DateTime
	if !dt.IsZero() {
		t.Error("zero value should be IsZero")
	}
	if dt.String() != "" {
		t.Errorf("zero String() = %q, want empty", dt.String())
	}
}

func TestDateTimeComparisons(t *testing.T) {
	a := DateTimeFromEpochSec(1000)
	b := DateTimeFromEpochSec(2000)
	same := DateTimeFromEpochSec(1000)

	if !a.Before(b) {
		t.Error("a should be before b")
	}
	if !b.After(a) {
		t.Error("b should be after a")
	}
	if !a.Equal(same) {
		t.Error("a should equal same")
	}
}

func TestDateTimeString(t *testing.T) {
	dt := DateTimeFromTime(time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC))
	got := dt.String()
	if got != "2026-01-15T14:30:00.000Z" {
		t.Errorf("String() = %q, want %q", got, "2026-01-15T14:30:00.000Z")
	}
}

func TestDateTimeString_WithMs(t *testing.T) {
	dt := DateTimeFromEpochMs(1702317649123)
	got := dt.String()
	if got != "2023-12-11T18:00:49.123Z" {
		t.Errorf("String() = %q, want ms preserved", got)
	}
}

func TestDateTimeJSONRoundtrip(t *testing.T) {
	type rec struct {
		T DateTime `json:"t"`
	}
	orig := rec{T: DateTimeFromTime(time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC))}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}

	var restored rec
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !restored.T.Equal(orig.T) {
		t.Errorf("roundtrip: got %v, want %v", restored.T, orig.T)
	}
}

func TestDateTimeUnmarshalJSON_Formats(t *testing.T) {
	tests := []struct {
		name string
		json string
		want time.Time
	}{
		{"RFC 3339", `"2026-01-15T14:30:00Z"`, time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)},
		{"Flex", `"2026-03-20;19:30:00"`, time.Date(2026, 3, 20, 19, 30, 0, 0, time.UTC)},
		{"REST", `"20231211-18:00:49"`, time.Date(2023, 12, 11, 18, 0, 49, 0, time.UTC)},
		{"epoch ms", `1702317649000`, time.UnixMilli(1702317649000).UTC()},
		{"epoch sec", `1702317649`, time.Unix(1702317649, 0).UTC()},
		{"quoted epoch ms", `"1702317649000"`, time.UnixMilli(1702317649000).UTC()},
		{"quoted epoch sec", `"1702317649"`, time.Unix(1702317649, 0).UTC()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var dt DateTime
			if err := json.Unmarshal([]byte(tt.json), &dt); err != nil {
				t.Fatalf("Unmarshal(%s) error: %v", tt.json, err)
			}
			if !dt.Time().Equal(tt.want) {
				t.Errorf("Unmarshal(%s) = %v, want %v", tt.json, dt.Time(), tt.want)
			}
		})
	}
}

func TestDateTimeUnmarshalJSON_NullAndEmpty(t *testing.T) {
	var dt DateTime
	if err := json.Unmarshal([]byte("null"), &dt); err != nil {
		t.Fatal(err)
	}
	if !dt.IsZero() {
		t.Error("null should produce zero datetime")
	}

	dt = DateTimeFromEpochSec(1000)
	if err := json.Unmarshal([]byte(`""`), &dt); err != nil {
		t.Fatal(err)
	}
	if !dt.IsZero() {
		t.Error("empty string should produce zero datetime")
	}
}

func TestDateTimeMarshalJSON_Zero(t *testing.T) {
	var dt DateTime
	data, err := json.Marshal(dt)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != `""` {
		t.Errorf("Marshal(zero) = %s, want empty string", data)
	}
}

func TestDateTimeTextRoundtrip(t *testing.T) {
	dt := DateTimeFromTime(time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC))
	data, err := dt.MarshalText()
	if err != nil {
		t.Fatal(err)
	}

	var restored DateTime
	if err := restored.UnmarshalText(data); err != nil {
		t.Fatal(err)
	}
	if !restored.Equal(dt) {
		t.Errorf("text roundtrip: got %v, want %v", restored, dt)
	}
}

func TestDateTimeTextEmpty(t *testing.T) {
	var dt DateTime
	if err := dt.UnmarshalText([]byte{}); err != nil {
		t.Fatal(err)
	}
	if !dt.IsZero() {
		t.Error("empty text should produce zero datetime")
	}
}

func TestDateTimeXMLAttr(t *testing.T) {
	type elem struct {
		XMLName   xml.Name `xml:"statement"`
		Generated DateTime `xml:"whenGenerated,attr"`
	}
	input := `<statement whenGenerated="2026-03-20;19:30:00"/>`
	var e elem
	if err := xml.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("XML unmarshal: %v", err)
	}
	want := time.Date(2026, 3, 20, 19, 30, 0, 0, time.UTC)
	if !e.Generated.Time().Equal(want) {
		t.Errorf("XML attr = %v, want %v", e.Generated, want)
	}
}

func TestDateTimeSQLRoundtrip(t *testing.T) {
	dt := DateTimeFromTime(time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC))
	v, err := dt.Value()
	if err != nil {
		t.Fatal(err)
	}
	s, ok := v.(string)
	if !ok {
		t.Fatalf("Value type = %T, want string", v)
	}

	var restored DateTime
	if err := restored.Scan(s); err != nil {
		t.Fatal(err)
	}
	if !restored.Equal(dt) {
		t.Errorf("SQL roundtrip: got %v, want %v", restored, dt)
	}
}

func TestDateTimeScanNil(t *testing.T) {
	var dt DateTime
	if err := dt.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if !dt.IsZero() {
		t.Error("Scan(nil) should produce zero datetime")
	}
}

func TestDateTimeScanInt64_Ms(t *testing.T) {
	var dt DateTime
	ms := int64(1702317649000)
	if err := dt.Scan(ms); err != nil {
		t.Fatal(err)
	}
	if dt.Time().UnixMilli() != ms {
		t.Errorf("Scan(ms) = %d, want %d", dt.Time().UnixMilli(), ms)
	}
}

func TestDateTimeScanInt64_Sec(t *testing.T) {
	var dt DateTime
	sec := int64(1702317649)
	if err := dt.Scan(sec); err != nil {
		t.Fatal(err)
	}
	if dt.Time().Unix() != sec {
		t.Errorf("Scan(sec) = %d, want %d", dt.Time().Unix(), sec)
	}
}

func TestDateTimeScanTime(t *testing.T) {
	var dt DateTime
	tm := time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)
	if err := dt.Scan(tm); err != nil {
		t.Fatal(err)
	}
	if !dt.Time().Equal(tm) {
		t.Errorf("Scan(time) = %v, want %v", dt.Time(), tm)
	}
}

func TestDateTimeScanUnsupported(t *testing.T) {
	var dt DateTime
	err := dt.Scan(3.14)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrUnsupportedType) {
		t.Errorf("error = %v, want ErrUnsupportedType", err)
	}
}

func TestDateTimeValueZero(t *testing.T) {
	var dt DateTime
	v, err := dt.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("Value(zero) = %v, want nil", v)
	}
}

func TestDateTimeEpochHeuristic(t *testing.T) {
	var dt DateTime

	// Below threshold: treated as seconds.
	if err := json.Unmarshal([]byte("946684799999"), &dt); err != nil {
		t.Fatal(err)
	}
	if dt.Time().Unix() != 946684799999 {
		t.Errorf("below threshold: got unix %d", dt.Time().Unix())
	}

	// At threshold (Jan 1 2000 in ms): treated as milliseconds (uses >=).
	if err := json.Unmarshal([]byte("946684800000"), &dt); err != nil {
		t.Fatal(err)
	}
	if dt.Time().UnixMilli() != 946684800000 {
		t.Errorf("at threshold: got unix_ms %d", dt.Time().UnixMilli())
	}

	// Above threshold: treated as milliseconds.
	if err := json.Unmarshal([]byte("946684800001"), &dt); err != nil {
		t.Fatal(err)
	}
	if dt.Time().UnixMilli() != 946684800001 {
		t.Errorf("above threshold: got unix_ms %d", dt.Time().UnixMilli())
	}
}

// --- NullDate ---

func TestNullDateJSON_Null(t *testing.T) {
	nd := NullDate{Valid: false}
	data, err := json.Marshal(nd)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("Marshal(!Valid) = %s, want null", data)
	}

	var restored NullDate
	if err := json.Unmarshal([]byte("null"), &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Error("Unmarshal(null) should set Valid=false")
	}
}

func TestNullDateJSON_EmptyString(t *testing.T) {
	var nd NullDate
	if err := json.Unmarshal([]byte(`""`), &nd); err != nil {
		t.Fatal(err)
	}
	if nd.Valid {
		t.Error("empty string should set Valid=false")
	}
}

func TestNullDateJSON_Valid(t *testing.T) {
	nd := NullDate{Date: NewDate(2026, 1, 15), Valid: true}
	data, err := json.Marshal(nd)
	if err != nil {
		t.Fatal(err)
	}

	var restored NullDate
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid {
		t.Error("should be valid")
	}
	if !restored.Date.Equal(nd.Date) {
		t.Errorf("roundtrip: got %v, want %v", restored.Date, nd.Date)
	}
}

func TestNullDateJSON_Roundtrip(t *testing.T) {
	type rec struct {
		Expiry NullDate `json:"expiry"`
		Trade  NullDate `json:"trade"`
	}
	orig := rec{
		Expiry: NullDate{Date: NewDate(2026, 1, 15), Valid: true},
		Trade:  NullDate{Valid: false},
	}
	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatal(err)
	}
	var restored rec
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Expiry.Valid || !restored.Expiry.Date.Equal(orig.Expiry.Date) {
		t.Errorf("Expiry roundtrip failed: %+v", restored.Expiry)
	}
	if restored.Trade.Valid {
		t.Errorf("Trade should be null: %+v", restored.Trade)
	}
}

func TestNullDateText(t *testing.T) {
	nd := NullDate{Valid: false}
	data, err := nd.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("MarshalText(!Valid) = %q, want empty", data)
	}

	var restored NullDate
	if err := restored.UnmarshalText([]byte{}); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Error("empty text should set Valid=false")
	}
}

func TestNullDateText_Valid(t *testing.T) {
	nd := NullDate{Date: NewDate(2026, 1, 15), Valid: true}
	data, err := nd.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	var restored NullDate
	if err := restored.UnmarshalText(data); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid || !restored.Date.Equal(nd.Date) {
		t.Errorf("text roundtrip failed: %+v", restored)
	}
}

func TestNullDateXMLAttr(t *testing.T) {
	type elem struct {
		XMLName xml.Name `xml:"trade"`
		Expiry  NullDate `xml:"expiry,attr"`
	}

	// Valid expiry
	input := `<trade expiry="2026-01-15"/>`
	var e elem
	if err := xml.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("XML unmarshal: %v", err)
	}
	if !e.Expiry.Valid || !e.Expiry.Date.Equal(NewDate(2026, 1, 15)) {
		t.Errorf("XML attr = %+v, want valid 2026-01-15", e.Expiry)
	}

	// Empty expiry
	input = `<trade expiry=""/>`
	e = elem{}
	if err := xml.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("XML unmarshal: %v", err)
	}
	if e.Expiry.Valid {
		t.Errorf("empty XML attr should be invalid: %+v", e.Expiry)
	}
}

func TestNullDateSQL(t *testing.T) {
	nd := NullDate{Date: NewDate(2026, 1, 15), Valid: true}
	v, err := nd.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v == nil {
		t.Fatal("valid NullDate Value() should not be nil")
	}

	var restored NullDate
	if err := restored.Scan(v); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid || !restored.Date.Equal(nd.Date) {
		t.Errorf("SQL roundtrip failed: %+v", restored)
	}
}

func TestNullDateSQL_Nil(t *testing.T) {
	nd := NullDate{Valid: false}
	v, err := nd.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("invalid NullDate Value() = %v, want nil", v)
	}

	var restored NullDate
	if err := restored.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Error("Scan(nil) should set Valid=false")
	}
}

func TestNullDateString(t *testing.T) {
	valid := NullDate{Date: NewDate(2026, 1, 15), Valid: true}
	if got := valid.String(); got != "2026-01-15" {
		t.Errorf("valid String() = %q, want %q", got, "2026-01-15")
	}
	null := NullDate{Valid: false}
	if got := null.String(); got != "" {
		t.Errorf("null String() = %q, want empty", got)
	}
}

// --- NullDateTime ---

func TestNullDateTimeJSON_Null(t *testing.T) {
	ndt := NullDateTime{Valid: false}
	data, err := json.Marshal(ndt)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "null" {
		t.Errorf("Marshal(!Valid) = %s, want null", data)
	}

	var restored NullDateTime
	if err := json.Unmarshal([]byte("null"), &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Error("Unmarshal(null) should set Valid=false")
	}
}

func TestNullDateTimeJSON_EmptyString(t *testing.T) {
	var ndt NullDateTime
	if err := json.Unmarshal([]byte(`""`), &ndt); err != nil {
		t.Fatal(err)
	}
	if ndt.Valid {
		t.Error("empty string should set Valid=false")
	}
}

func TestNullDateTimeJSON_Valid(t *testing.T) {
	ndt := NullDateTime{DateTime: DateTimeFromEpochSec(1702317649), Valid: true}
	data, err := json.Marshal(ndt)
	if err != nil {
		t.Fatal(err)
	}

	var restored NullDateTime
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid {
		t.Error("should be valid")
	}
	if !restored.DateTime.Equal(ndt.DateTime) {
		t.Errorf("roundtrip: got %v, want %v", restored.DateTime, ndt.DateTime)
	}
}

func TestNullDateTimeJSON_EpochMs(t *testing.T) {
	var ndt NullDateTime
	if err := json.Unmarshal([]byte(`1702317649000`), &ndt); err != nil {
		t.Fatal(err)
	}
	if !ndt.Valid {
		t.Error("bare epoch should be valid")
	}
	if ndt.DateTime.Time().UnixMilli() != 1702317649000 {
		t.Errorf("epoch ms: got %d", ndt.DateTime.Time().UnixMilli())
	}
}

func TestNullDateTimeText(t *testing.T) {
	ndt := NullDateTime{Valid: false}
	data, err := ndt.MarshalText()
	if err != nil {
		t.Fatal(err)
	}
	if len(data) != 0 {
		t.Errorf("MarshalText(!Valid) = %q, want empty", data)
	}

	var restored NullDateTime
	if err := restored.UnmarshalText([]byte{}); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Error("empty text should set Valid=false")
	}
}

func TestNullDateTimeXMLAttr(t *testing.T) {
	type elem struct {
		XMLName   xml.Name     `xml:"statement"`
		Generated NullDateTime `xml:"whenGenerated,attr"`
	}
	input := `<statement whenGenerated="2026-03-20;19:30:00"/>`
	var e elem
	if err := xml.Unmarshal([]byte(input), &e); err != nil {
		t.Fatalf("XML unmarshal: %v", err)
	}
	if !e.Generated.Valid {
		t.Fatal("should be valid")
	}
	want := time.Date(2026, 3, 20, 19, 30, 0, 0, time.UTC)
	if !e.Generated.DateTime.Time().Equal(want) {
		t.Errorf("XML attr = %v, want %v", e.Generated.DateTime, want)
	}
}

func TestNullDateTimeSQL(t *testing.T) {
	ndt := NullDateTime{
		DateTime: DateTimeFromTime(time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)),
		Valid:    true,
	}
	v, err := ndt.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v == nil {
		t.Fatal("valid NullDateTime Value() should not be nil")
	}

	var restored NullDateTime
	if err := restored.Scan(v); err != nil {
		t.Fatal(err)
	}
	if !restored.Valid || !restored.DateTime.Equal(ndt.DateTime) {
		t.Errorf("SQL roundtrip failed: %+v", restored)
	}
}

func TestNullDateTimeSQL_Nil(t *testing.T) {
	ndt := NullDateTime{Valid: false}
	v, err := ndt.Value()
	if err != nil {
		t.Fatal(err)
	}
	if v != nil {
		t.Errorf("invalid NullDateTime Value() = %v, want nil", v)
	}

	var restored NullDateTime
	if err := restored.Scan(nil); err != nil {
		t.Fatal(err)
	}
	if restored.Valid {
		t.Error("Scan(nil) should set Valid=false")
	}
}

func TestNullDateTimeString(t *testing.T) {
	valid := NullDateTime{
		DateTime: DateTimeFromTime(time.Date(2026, 1, 15, 14, 30, 0, 0, time.UTC)),
		Valid:    true,
	}
	if got := valid.String(); got != "2026-01-15T14:30:00.000Z" {
		t.Errorf("valid String() = %q", got)
	}
	null := NullDateTime{Valid: false}
	if got := null.String(); got != "" {
		t.Errorf("null String() = %q, want empty", got)
	}
}

// --- Errors ---

func TestErrorSentinels(t *testing.T) {
	_, err := ParseDate("garbage")
	if !errors.Is(err, ErrInvalidDate) {
		t.Errorf("ParseDate error = %v, want ErrInvalidDate", err)
	}

	_, err = ParseDateTime("garbage")
	if !errors.Is(err, ErrInvalidDateTime) {
		t.Errorf("ParseDateTime error = %v, want ErrInvalidDateTime", err)
	}
}

// --- Millisecond precision ---

func TestDateTimeMsRoundtrip(t *testing.T) {
	dt := DateTimeFromEpochMs(1702317649123)
	data, err := json.Marshal(dt)
	if err != nil {
		t.Fatal(err)
	}
	var restored DateTime
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.Time().UnixMilli() != 1702317649123 {
		t.Errorf("ms roundtrip: got %d, want 1702317649123", restored.Time().UnixMilli())
	}
}

func TestDateTimeMsSQLRoundtrip(t *testing.T) {
	dt := DateTimeFromEpochMs(1702317649456)
	v, err := dt.Value()
	if err != nil {
		t.Fatal(err)
	}
	var restored DateTime
	if err := restored.Scan(v); err != nil {
		t.Fatal(err)
	}
	if restored.Time().UnixMilli() != 1702317649456 {
		t.Errorf("SQL ms roundtrip: got %d, want 1702317649456", restored.Time().UnixMilli())
	}
}

// --- Invalid Null unmarshal preserves Valid=false ---

func TestNullDateUnmarshalJSON_InvalidKeepsInvalid(t *testing.T) {
	nd := NullDate{Date: NewDate(2026, 1, 15), Valid: true}
	err := json.Unmarshal([]byte(`"not-a-date"`), &nd)
	if err == nil {
		t.Fatal("expected error from invalid date")
	}
	if nd.Valid {
		t.Error("NullDate should be invalid after failed unmarshal")
	}
}

func TestNullDateTimeUnmarshalJSON_InvalidKeepsInvalid(t *testing.T) {
	ndt := NullDateTime{
		DateTime: DateTimeFromEpochSec(1000),
		Valid:    true,
	}
	err := json.Unmarshal([]byte(`"not-a-datetime"`), &ndt)
	if err == nil {
		t.Fatal("expected error from invalid datetime")
	}
	if ndt.Valid {
		t.Error("NullDateTime should be invalid after failed unmarshal")
	}
}

func TestNullDateScan_FailureKeepsInvalid(t *testing.T) {
	nd := NullDate{Date: NewDate(2026, 1, 15), Valid: true}
	if err := nd.Scan(struct{}{}); err == nil {
		t.Fatal("expected error from unsupported scan type")
	}
	if nd.Valid {
		t.Error("NullDate should be invalid after failed Scan")
	}
}

func TestNullDateTimeScan_FailureKeepsInvalid(t *testing.T) {
	ndt := NullDateTime{
		DateTime: DateTimeFromEpochSec(1000),
		Valid:    true,
	}
	if err := ndt.Scan(3.14); err == nil {
		t.Fatal("expected error from unsupported scan type")
	}
	if ndt.Valid {
		t.Error("NullDateTime should be invalid after failed Scan")
	}
}
