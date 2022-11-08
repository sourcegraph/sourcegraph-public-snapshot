package insights

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold"
)

func Test_calculateRecordingTimes(t *testing.T) {
	defaultInterval := timeInterval{
		unit:  month,
		value: 2,
	}
	createdAt := time.Date(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	defaultTimes := buildRecordingTimes(12, defaultInterval, createdAt)
	stringDefaultTimes := []string{
		"2021-01-01 00:00:00 +0000 UTC",
		"2021-03-01 00:00:00 +0000 UTC",
		"2021-05-01 00:00:00 +0000 UTC",
		"2021-07-01 00:00:00 +0000 UTC",
		"2021-09-01 00:00:00 +0000 UTC",
		"2021-11-01 00:00:00 +0000 UTC",
		"2022-01-01 00:00:00 +0000 UTC",
		"2022-03-01 00:00:00 +0000 UTC",
		"2022-05-01 00:00:00 +0000 UTC",
		"2022-07-01 00:00:00 +0000 UTC",
		"2022-09-01 00:00:00 +0000 UTC",
		"2022-11-01 00:00:00 +0000 UTC",
	}

	stringify := func(times []time.Time) []string {
		s := make([]string, 0, len(times))
		for _, t := range times {
			s = append(s, t.String())
		}
		return s
	}
	if diff := cmp.Diff(stringDefaultTimes, stringify(defaultTimes)); diff != "" {
		t.Fatalf("test setup is tainted: got diff %v", diff)
	}

	testCases := []struct {
		lastRecordedAt time.Time
		interval       timeInterval
		existingTimes  []time.Time
		want           autogold.Value
	}{
		{
			interval: defaultInterval,
			want:     autogold.Want("no existing times returns all generated times", stringDefaultTimes),
		},
		{
			interval:       defaultInterval,
			lastRecordedAt: createdAt.AddDate(0, 5, 3),
			want: autogold.Want("no existing times and lastRecordedAt returns generated times",
				append(
					stringDefaultTimes,
					"2023-01-01 00:00:00 +0000 UTC",
					"2023-03-01 00:00:00 +0000 UTC",
				),
			),
		},
		{
			interval:      timeInterval{unit: week, value: 2},
			existingTimes: []time.Time{time.Date(2022, 10, 21, 0, 0, 0, 0, time.UTC)}, // existing point within half an interval
			want: autogold.Want("existing time returns modified list", []string{
				"2022-05-31 00:00:00 +0000 UTC",
				"2022-06-14 00:00:00 +0000 UTC",
				"2022-06-28 00:00:00 +0000 UTC",
				"2022-07-12 00:00:00 +0000 UTC",
				"2022-07-26 00:00:00 +0000 UTC",
				"2022-08-09 00:00:00 +0000 UTC",
				"2022-08-23 00:00:00 +0000 UTC",
				"2022-09-06 00:00:00 +0000 UTC",
				"2022-09-20 00:00:00 +0000 UTC",
				"2022-10-04 00:00:00 +0000 UTC",
				"2022-10-21 00:00:00 +0000 UTC", // this is the modified rough point
				"2022-11-01 00:00:00 +0000 UTC",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			calculated := calculateRecordingTimes(createdAt, tc.lastRecordedAt, tc.interval, tc.existingTimes)
			tc.want.Equal(t, stringify(calculated))
		})
	}
}
