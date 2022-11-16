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
			want: autogold.Want("2 week intervals with existing time returns modified list", []string{
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
		{
			interval:      timeInterval{unit: hour, value: 2},
			existingTimes: []time.Time{time.Date(2022, 9, 31, 03, 50, 0, 0, time.UTC), time.Date(2022, 10, 31, 18, 26, 0, 0, time.UTC)},
			want: autogold.Want("2 hour intervals with existing time returns modified list", []string{
				"2022-10-31 02:00:00 +0000 UTC",
				"2022-10-31 04:00:00 +0000 UTC",
				"2022-10-31 06:00:00 +0000 UTC",
				"2022-10-31 08:00:00 +0000 UTC",
				"2022-10-31 10:00:00 +0000 UTC",
				"2022-10-31 12:00:00 +0000 UTC",
				"2022-10-31 14:00:00 +0000 UTC",
				"2022-10-31 16:00:00 +0000 UTC",
				"2022-10-31 18:26:00 +0000 UTC", // this is the modified rough point
				"2022-10-31 20:00:00 +0000 UTC",
				"2022-10-31 22:00:00 +0000 UTC",
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

func Test_buildRecordingTimes(t *testing.T) {
	startTime := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	buildFrameTest := func(count int, interval timeInterval, current time.Time) []string {
		got := buildRecordingTimes(count, interval, current)
		return convert(got)
	}

	testCases := []struct {
		numPoints int
		interval  timeInterval
		startTime time.Time
		want      autogold.Value
	}{
		{
			numPoints: 1,
			interval:  timeInterval{month, 1},
			startTime: startTime,
			want: autogold.Want("one point", []string{
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			numPoints: 2,
			interval:  timeInterval{month, 1},
			startTime: startTime,
			want: autogold.Want("two points 1 month intervals", []string{
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			numPoints: 6,
			interval:  timeInterval{month, 1},
			startTime: startTime,
			want: autogold.Want("6 points 1 month intervals", []string{
				"2021-07-01 00:00:00 +0000 UTC",
				"2021-08-01 00:00:00 +0000 UTC",
				"2021-09-01 00:00:00 +0000 UTC",
				"2021-10-01 00:00:00 +0000 UTC",
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			numPoints: 12,
			interval:  timeInterval{week, 2},
			startTime: startTime,
			want: autogold.Want("12 points 2 week intervals", []string{
				"2021-06-30 00:00:00 +0000 UTC",
				"2021-07-14 00:00:00 +0000 UTC",
				"2021-07-28 00:00:00 +0000 UTC",
				"2021-08-11 00:00:00 +0000 UTC",
				"2021-08-25 00:00:00 +0000 UTC",
				"2021-09-08 00:00:00 +0000 UTC",
				"2021-09-22 00:00:00 +0000 UTC",
				"2021-10-06 00:00:00 +0000 UTC",
				"2021-10-20 00:00:00 +0000 UTC",
				"2021-11-03 00:00:00 +0000 UTC",
				"2021-11-17 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			numPoints: 6,
			interval:  timeInterval{day, 2},
			startTime: startTime,
			want: autogold.Want("6 points 2 day intervals", []string{
				"2021-11-21 00:00:00 +0000 UTC",
				"2021-11-23 00:00:00 +0000 UTC",
				"2021-11-25 00:00:00 +0000 UTC",
				"2021-11-27 00:00:00 +0000 UTC",
				"2021-11-29 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			numPoints: 6,
			interval:  timeInterval{hour, 2},
			startTime: startTime,
			want: autogold.Want("6 points 2 hour intervals", []string{
				"2021-11-30 14:00:00 +0000 UTC",
				"2021-11-30 16:00:00 +0000 UTC",
				"2021-11-30 18:00:00 +0000 UTC",
				"2021-11-30 20:00:00 +0000 UTC",
				"2021-11-30 22:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			numPoints: 6,
			interval:  timeInterval{year, 1},
			startTime: startTime,
			want: autogold.Want("6 points 1 year intervals", []string{
				"2016-12-01 00:00:00 +0000 UTC",
				"2017-12-01 00:00:00 +0000 UTC",
				"2018-12-01 00:00:00 +0000 UTC",
				"2019-12-01 00:00:00 +0000 UTC",
				"2020-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			tc.want.Equal(t, buildFrameTest(tc.numPoints, tc.interval, tc.startTime))
		})
	}
}

func Test_buildRecordingTimesBetween(t *testing.T) {
	fromTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		from     time.Time
		to       time.Time
		interval timeInterval
		want     autogold.Value
	}{
		{
			fromTime,
			fromTime.AddDate(0, 1, 0),
			timeInterval{week, 1},
			autogold.Want("one month at 1 week intervals", []string{
				"2021-01-01 00:00:00 +0000 UTC",
				"2021-01-08 00:00:00 +0000 UTC",
				"2021-01-15 00:00:00 +0000 UTC",
				"2021-01-22 00:00:00 +0000 UTC",
				"2021-01-29 00:00:00 +0000 UTC",
			}),
		},
		{
			fromTime,
			fromTime.AddDate(0, 6, 0),
			timeInterval{week, 4},
			autogold.Want("6 months at 4 week intervals", []string{
				"2021-01-01 00:00:00 +0000 UTC",
				"2021-01-29 00:00:00 +0000 UTC",
				"2021-02-26 00:00:00 +0000 UTC",
				"2021-03-26 00:00:00 +0000 UTC",
				"2021-04-23 00:00:00 +0000 UTC",
				"2021-05-21 00:00:00 +0000 UTC",
				"2021-06-18 00:00:00 +0000 UTC",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := buildRecordingTimesBetween(tc.from, tc.to, tc.interval)
			tc.want.Equal(t, convert(got))
		})
	}
}

func convert(times []time.Time) []string {
	var got []string
	for _, result := range times {
		got = append(got, result.String())
	}
	return got
}
