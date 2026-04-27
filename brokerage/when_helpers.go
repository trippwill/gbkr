package brokerage

import "github.com/trippwill/gbkr/when"

// parseNullDateFromJSON maps a JSON-decoded string to a NullDate.
// Empty string → NullDate{Valid: false}; non-empty → parsed date.
func parseNullDateFromJSON(s string) when.NullDate {
	if s == "" {
		return when.NullDate{}
	}
	d, err := when.ParseDate(s)
	if err != nil {
		return when.NullDate{}
	}
	return when.NullDate{Date: d, Valid: true}
}
