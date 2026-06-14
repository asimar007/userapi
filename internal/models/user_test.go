package models

import (
	"testing"
	"time"
)

func mustDate(t *testing.T, s string) time.Time {
	t.Helper()
	d, err := time.Parse(DateLayout, s)
	if err != nil {
		t.Fatalf("bad test date %q: %v", s, err)
	}
	return d
}

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name string
		dob  string
		now  string
		want int
	}{
		{"birthday already passed this year", "1990-05-10", "2025-06-14", 35},
		{"birthday is today", "1990-06-14", "2025-06-14", 35},
		{"birthday not yet this year", "1990-12-25", "2025-06-14", 34},
		{"day before birthday", "1990-06-15", "2025-06-14", 34},
		{"day after birthday", "1990-06-13", "2025-06-14", 35},
		{"born this year", "2025-01-01", "2025-06-14", 0},
		{"exactly one year old", "2024-06-14", "2025-06-14", 1},
		{"one day short of one year", "2024-06-15", "2025-06-14", 0},
		{"leap day birthday, non-leap year before", "2000-02-29", "2025-02-28", 24},
		{"leap day birthday, on Mar 1", "2000-02-29", "2025-03-01", 25},
		{"future dob clamps to zero", "2030-01-01", "2025-06-14", 0},
		{"newborn today", "2025-06-14", "2025-06-14", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dob := mustDate(t, tt.dob)
			now := mustDate(t, tt.now)
			got := CalculateAge(dob, now)
			if got != tt.want {
				t.Errorf("CalculateAge(%s, %s) = %d, want %d",
					tt.dob, tt.now, got, tt.want)
			}
		})
	}
}
