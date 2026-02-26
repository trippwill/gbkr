package brokerage

import (
	"errors"
	"testing"
)

func TestBarUnit_String(t *testing.T) {
	tests := []struct {
		unit BarUnit
		want string
	}{
		{BarSeconds, "S"},
		{BarMinutes, "min"},
		{BarHours, "h"},
		{BarDays, "d"},
		{BarWeeks, "w"},
		{BarMonths, "m"},
		{BarUnit(0), "?"},
		{BarUnit(99), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.unit.String()
			if got != tt.want {
				t.Errorf("BarUnit(%d).String() = %q, want %q", tt.unit, got, tt.want)
			}
		})
	}
}

func TestBarUnit_Valid(t *testing.T) {
	tests := []struct {
		unit BarUnit
		want bool
	}{
		{BarUnit(0), false},
		{BarSeconds, true},
		{BarMonths, true},
		{BarUnit(7), false},
		{BarUnit(-1), false},
	}

	for _, tt := range tests {
		got := tt.unit.Valid()
		if got != tt.want {
			t.Errorf("BarUnit(%d).Valid() = %v, want %v", tt.unit, got, tt.want)
		}
	}
}

func TestPeriodUnit_String(t *testing.T) {
	tests := []struct {
		unit PeriodUnit
		want string
	}{
		{PeriodMinutes, "min"},
		{PeriodHours, "h"},
		{PeriodDays, "d"},
		{PeriodWeeks, "w"},
		{PeriodMonths, "m"},
		{PeriodYears, "y"},
		{PeriodUnit(0), "?"},
		{PeriodUnit(99), "?"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.unit.String()
			if got != tt.want {
				t.Errorf("PeriodUnit(%d).String() = %q, want %q", tt.unit, got, tt.want)
			}
		})
	}
}

func TestPeriodUnit_Valid(t *testing.T) {
	tests := []struct {
		unit PeriodUnit
		want bool
	}{
		{PeriodUnit(0), false},
		{PeriodMinutes, true},
		{PeriodYears, true},
		{PeriodUnit(7), false},
		{PeriodUnit(-1), false},
	}

	for _, tt := range tests {
		got := tt.unit.Valid()
		if got != tt.want {
			t.Errorf("PeriodUnit(%d).Valid() = %v, want %v", tt.unit, got, tt.want)
		}
	}
}

func TestBarSize_String(t *testing.T) {
	tests := []struct {
		name string
		bar  BarSize
		want string
	}{
		{"5min", BarSize{5, BarMinutes}, "5min"},
		{"1h", BarSize{1, BarHours}, "1h"},
		{"1d", BarSize{1, BarDays}, "1d"},
		{"1w", BarSize{1, BarWeeks}, "1w"},
		{"1m", BarSize{1, BarMonths}, "1m"},
		{"10S", BarSize{10, BarSeconds}, "10S"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.bar.String()
			if got != tt.want {
				t.Errorf("BarSize.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTimePeriod_String(t *testing.T) {
	tests := []struct {
		name   string
		period TimePeriod
		want   string
	}{
		{"5min", TimePeriod{5, PeriodMinutes}, "5min"},
		{"1h", TimePeriod{1, PeriodHours}, "1h"},
		{"6d", TimePeriod{6, PeriodDays}, "6d"},
		{"2w", TimePeriod{2, PeriodWeeks}, "2w"},
		{"3m", TimePeriod{3, PeriodMonths}, "3m"},
		{"1y", TimePeriod{1, PeriodYears}, "1y"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.period.String()
			if got != tt.want {
				t.Errorf("TimePeriod.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBar(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		b, err := Bar(5, BarMinutes)
		if err != nil {
			t.Fatalf("Bar(5, BarMinutes) unexpected error: %v", err)
		}
		if b.Count != 5 || b.Unit != BarMinutes {
			t.Errorf("got %+v, want Count=5 Unit=BarMinutes", b)
		}
	})

	t.Run("zero count", func(t *testing.T) {
		_, err := Bar(0, BarMinutes)
		if err == nil {
			t.Fatal("Bar(0, ...) should return error")
		}
		if !errors.Is(err, ErrInvalidCount) {
			t.Errorf("expected ErrInvalidCount, got %v", err)
		}
	})

	t.Run("negative count", func(t *testing.T) {
		_, err := Bar(-1, BarMinutes)
		if err == nil {
			t.Fatal("Bar(-1, ...) should return error")
		}
		if !errors.Is(err, ErrInvalidCount) {
			t.Errorf("expected ErrInvalidCount, got %v", err)
		}
	})

	t.Run("invalid unit zero", func(t *testing.T) {
		_, err := Bar(1, BarUnit(0))
		if err == nil {
			t.Fatal("Bar(1, 0) should return error")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Errorf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("invalid unit out of range", func(t *testing.T) {
		_, err := Bar(1, BarUnit(99))
		if err == nil {
			t.Fatal("Bar(1, 99) should return error")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Errorf("expected ErrInvalidUnit, got %v", err)
		}
	})
}

func TestPeriod(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		p, err := Period(6, PeriodDays)
		if err != nil {
			t.Fatalf("Period(6, PeriodDays) unexpected error: %v", err)
		}
		if p.Count != 6 || p.Unit != PeriodDays {
			t.Errorf("got %+v, want Count=6 Unit=PeriodDays", p)
		}
	})

	t.Run("zero count", func(t *testing.T) {
		_, err := Period(0, PeriodDays)
		if err == nil {
			t.Fatal("Period(0, ...) should return error")
		}
		if !errors.Is(err, ErrInvalidCount) {
			t.Errorf("expected ErrInvalidCount, got %v", err)
		}
	})

	t.Run("negative count", func(t *testing.T) {
		_, err := Period(-5, PeriodDays)
		if err == nil {
			t.Fatal("Period(-5, ...) should return error")
		}
		if !errors.Is(err, ErrInvalidCount) {
			t.Errorf("expected ErrInvalidCount, got %v", err)
		}
	})

	t.Run("invalid unit zero", func(t *testing.T) {
		_, err := Period(1, PeriodUnit(0))
		if err == nil {
			t.Fatal("Period(1, 0) should return error")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Errorf("expected ErrInvalidUnit, got %v", err)
		}
	})

	t.Run("invalid unit out of range", func(t *testing.T) {
		_, err := Period(1, PeriodUnit(99))
		if err == nil {
			t.Fatal("Period(1, 99) should return error")
		}
		if !errors.Is(err, ErrInvalidUnit) {
			t.Errorf("expected ErrInvalidUnit, got %v", err)
		}
	})
}

func TestValidationError_As(t *testing.T) {
	_, err := Bar(0, BarMinutes)
	if err == nil {
		t.Fatal("expected error")
	}

	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatal("errors.As should extract *ValidationError")
	}
	if ve.Kind != ErrInvalidCount {
		t.Errorf("Kind = %v, want %v", ve.Kind, ErrInvalidCount)
	}
	if ve.Value != 0 {
		t.Errorf("Value = %d, want 0", ve.Value)
	}
}

func TestSnapshotField_String(t *testing.T) {
	if FieldLast.String() != "31" {
		t.Errorf("FieldLast.String() = %q, want %q", FieldLast.String(), "31")
	}
}

func TestValidationError_String(t *testing.T) {
	err := ErrInvalidCountValue(-3)
	msg := err.Error()
	if msg != "count must be positive: -3" {
		t.Errorf("Error() = %q, want %q", msg, "count must be positive: -3")
	}

	err = ErrInvalidUnitValue(99)
	msg = err.Error()
	if msg != "invalid unit: 99" {
		t.Errorf("Error() = %q, want %q", msg, "invalid unit: 99")
	}
}
