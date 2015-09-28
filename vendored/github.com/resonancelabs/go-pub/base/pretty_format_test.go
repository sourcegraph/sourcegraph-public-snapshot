package base

import (
	"fmt"
	"testing"
)

// A simple test that doesn't require the gocheck fixture.
func TestPrettyFormatHumanLifetimet(t *testing.T) {
	tests := []struct {
		input    Micros
		expected string
	}{
		{0, "1 day"},
		{1, "1 day"},
		{MICROS_PER_DAY - 1, "1 day"},
		{MICROS_PER_DAY, "1 day"},
		{MICROS_PER_DAY + 1, "2 days"},
		{MICROS_PER_WEEK, "7 days"},
		{MICROS_PER_WEEK + MICROS_PER_DAY + 1, "9 days"},
		{4 * MICROS_PER_WEEK, "4 weeks"},
		{8 * MICROS_PER_WEEK, "8 weeks"},
		{91 * MICROS_PER_DAY, "4 months"},
		{100 * MICROS_PER_DAY, "4 months"},
		{365 * MICROS_PER_DAY, "1 year"},
		{366 * MICROS_PER_DAY, "1 year, 1 month"},
		{400 * MICROS_PER_DAY, "1 year, 2 months"},
		{2 * 365 * MICROS_PER_DAY, "2 years"},
		{(2*365 + 20) * MICROS_PER_DAY, "2 years, 1 month"},
		{(2*365 + 90) * MICROS_PER_DAY, "2 years, 3 months"},
		{3 * 365 * MICROS_PER_DAY, "3 years"},
		{5 * 365 * MICROS_PER_DAY, "5 years"},
	}
	for _, tc := range tests {
		result := PrettyFormatHumanLifetime(tc.input)
		if result != tc.expected {
			t.Error(fmt.Sprintf("expected '%s', got '%s' from '%s'", tc.expected, result, tc.input))
		}
	}
}

// A simple test that doesn't require the gocheck fixture.
func TestPrettyFormatInt(t *testing.T) {
	tests := []struct {
		input    int
		expected string
	}{
		{0, "0"},
		{1, "1"},
		{99, "99"},
		{999, "999"},
		{1000, "1,000"},
		{1001, "1,001"},
		{54000, "54,000"},
		{172541, "172,541"},
	}
	for _, tc := range tests {
		result := PrettyFormatInt(tc.input)
		if result != tc.expected {
			t.Error(fmt.Sprintf("expected '%v', got '%v' from '%v'", tc.expected, result, tc.input))
		}
	}
}
