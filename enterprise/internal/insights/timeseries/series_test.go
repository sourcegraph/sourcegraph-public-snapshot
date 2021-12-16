package timeseries

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
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
