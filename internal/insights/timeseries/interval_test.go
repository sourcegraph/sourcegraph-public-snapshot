package timeseries

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
)

func TestStepForward(t *testing.T) {
	startTime := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		interval TimeInterval
		want     string
	}{
		{name: "1", interval: TimeInterval{Unit: types.Month, Value: 1}, want: "2022-01-01 00:00:00 +0000 UTC"},
		{name: "2", interval: TimeInterval{Unit: types.Month, Value: 13}, want: "2023-01-01 00:00:00 +0000 UTC"},
		{name: "3", interval: TimeInterval{Unit: types.Day, Value: 1}, want: "2021-12-02 00:00:00 +0000 UTC"},
		{name: "4", interval: TimeInterval{Unit: types.Hour, Value: 1}, want: "2021-12-01 01:00:00 +0000 UTC"},
		{name: "5", interval: TimeInterval{Unit: types.Year, Value: 1}, want: "2022-12-01 00:00:00 +0000 UTC"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			convert := func(input time.Time) string {
				return input.String()
			}
			got := convert(test.interval.StepForwards(startTime))
			if diff := cmp.Diff(test.want, got); diff != "" {
				t.Errorf("unexpected result (want/got): %v", diff)
			}
		})
	}
}

func TestIsValid(t *testing.T) {
	tests := []struct {
		name     string
		want     bool
		interval TimeInterval
	}{
		{
			name: "month valid",
			want: true,
			interval: TimeInterval{
				Unit:  types.Month,
				Value: 1,
			},
		},
		{
			name: "day valid",
			want: true,
			interval: TimeInterval{
				Unit:  types.Day,
				Value: 1,
			},
		},
		{
			name: "year valid",
			want: true,
			interval: TimeInterval{
				Unit:  types.Year,
				Value: 1,
			},
		},
		{
			name: "hour valid",
			want: true,
			interval: TimeInterval{
				Unit:  types.Hour,
				Value: 1,
			},
		},
		{
			name: "week valid",
			want: true,
			interval: TimeInterval{
				Unit:  types.Week,
				Value: 1,
			},
		},
		{
			name: "invalid type",
			want: false,
			interval: TimeInterval{
				Unit:  types.IntervalUnit("asdf"),
				Value: 1,
			},
		},
		{
			name: "invalid value",
			want: false,
			interval: TimeInterval{
				Unit:  types.Week,
				Value: -1,
			},
		},
		{
			name: "valid zero value",
			want: true,
			interval: TimeInterval{
				Unit:  types.Week,
				Value: 0,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := test.interval.IsValid()
			want := test.want
			if got != want {
				t.Errorf("unexpected IsValid: want: %v got: %v", want, got)
			}
		})
	}
}
