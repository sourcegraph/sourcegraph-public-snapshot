package recording_times

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/autogold/v2"
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
		name           string
		lastRecordedAt time.Time
		interval       timeInterval
		existingTimes  []time.Time
		want           autogold.Value
	}{
		{
			name:     "no existing times returns all generated times",
			interval: defaultInterval,
			want:     autogold.Expect(stringDefaultTimes),
		},
		{
			name:           "no existing times and lastRecordedAt returns generated times",
			interval:       defaultInterval,
			lastRecordedAt: createdAt.AddDate(0, 5, 3),
			want: autogold.Expect(append(
				stringDefaultTimes,
				"2023-01-01 00:00:00 +0000 UTC",
				"2023-03-01 00:00:00 +0000 UTC",
			),
			),
		},
		{
			name:          "oldest existing point after oldest expected point",
			interval:      timeInterval{unit: week, value: 2},
			existingTimes: []time.Time{time.Date(2022, 11, 2, 1, 0, 0, 0, time.UTC)}, // existing point within half an interval
			want: autogold.Expect([]string{
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
				"2022-10-18 00:00:00 +0000 UTC",
				"2022-11-02 01:00:00 +0000 UTC", // this is the existing point
			}),
		},
		{
			name:     "newest existing point before newest expected point",
			interval: timeInterval{unit: hour, value: 2},
			existingTimes: []time.Time{
				time.Date(2022, 10, 31, 2, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 4, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 6, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 8, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 10, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 12, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 14, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 16, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 18, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 20, 12, 0, 0, time.UTC),
			},
			want: autogold.Expect([]string{
				"2022-10-31 02:01:00 +0000 UTC",
				"2022-10-31 04:01:00 +0000 UTC",
				"2022-10-31 06:01:00 +0000 UTC",
				"2022-10-31 08:01:00 +0000 UTC",
				"2022-10-31 10:01:00 +0000 UTC",
				"2022-10-31 12:01:00 +0000 UTC",
				"2022-10-31 14:01:00 +0000 UTC",
				"2022-10-31 16:01:00 +0000 UTC",
				"2022-10-31 18:01:00 +0000 UTC",
				"2022-10-31 20:12:00 +0000 UTC", // trailing points are added after this
				"2022-10-31 22:00:00 +0000 UTC",
				"2022-11-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:     "oldest expected before oldest existing and newest existing before newest expected",
			interval: timeInterval{unit: hour, value: 2},
			existingTimes: []time.Time{
				time.Date(2022, 10, 31, 8, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 10, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 12, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 14, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 16, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 18, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 20, 12, 0, 0, time.UTC),
			},
			want: autogold.Expect([]string{
				"2022-10-31 02:00:00 +0000 UTC",
				"2022-10-31 04:00:00 +0000 UTC",
				"2022-10-31 06:00:00 +0000 UTC",
				"2022-10-31 08:01:00 +0000 UTC", // leading points are added before this
				"2022-10-31 10:01:00 +0000 UTC",
				"2022-10-31 12:01:00 +0000 UTC",
				"2022-10-31 14:01:00 +0000 UTC",
				"2022-10-31 16:01:00 +0000 UTC",
				"2022-10-31 18:01:00 +0000 UTC",
				"2022-10-31 20:12:00 +0000 UTC", // trailing points are added after this
				"2022-10-31 22:00:00 +0000 UTC",
				"2022-11-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:     "all existing points are reused with valleys",
			interval: timeInterval{unit: hour, value: 2},
			existingTimes: []time.Time{
				time.Date(2022, 10, 31, 2, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 4, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 6, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 8, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 10, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 12, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 14, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 16, 1, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 20, 12, 0, 0, time.UTC),
				time.Date(2022, 10, 31, 22, 2, 0, 0, time.UTC),
				time.Date(2022, 11, 1, 0, 2, 0, 0, time.UTC),
			},
			want: autogold.Expect([]string{
				"2022-10-31 02:01:00 +0000 UTC",
				"2022-10-31 04:01:00 +0000 UTC",
				"2022-10-31 06:01:00 +0000 UTC",
				"2022-10-31 08:01:00 +0000 UTC",
				"2022-10-31 10:01:00 +0000 UTC",
				"2022-10-31 12:01:00 +0000 UTC",
				"2022-10-31 14:01:00 +0000 UTC",
				"2022-10-31 16:01:00 +0000 UTC",
				"2022-10-31 20:12:00 +0000 UTC", // we would have a gap before this point.
				"2022-10-31 22:02:00 +0000 UTC",
				"2022-11-01 00:02:00 +0000 UTC",
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
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
		name      string
		numPoints int
		interval  timeInterval
		startTime time.Time
		want      autogold.Value
	}{
		{
			name:      "one point",
			numPoints: 1,
			interval:  timeInterval{month, 1},
			startTime: startTime,
			want: autogold.Expect([]string{
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:      "two points 1 month intervals",
			numPoints: 2,
			interval:  timeInterval{month, 1},
			startTime: startTime,
			want: autogold.Expect([]string{
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:      "6 points 1 month intervals",
			numPoints: 6,
			interval:  timeInterval{month, 1},
			startTime: startTime,
			want: autogold.Expect([]string{
				"2021-07-01 00:00:00 +0000 UTC",
				"2021-08-01 00:00:00 +0000 UTC",
				"2021-09-01 00:00:00 +0000 UTC",
				"2021-10-01 00:00:00 +0000 UTC",
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:      "12 points 2 week intervals",
			numPoints: 12,
			interval:  timeInterval{week, 2},
			startTime: startTime,
			want: autogold.Expect([]string{
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
			name:      "6 points 2 day intervals",
			numPoints: 6,
			interval:  timeInterval{day, 2},
			startTime: startTime,
			want: autogold.Expect([]string{
				"2021-11-21 00:00:00 +0000 UTC",
				"2021-11-23 00:00:00 +0000 UTC",
				"2021-11-25 00:00:00 +0000 UTC",
				"2021-11-27 00:00:00 +0000 UTC",
				"2021-11-29 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:      "6 points 2 hour intervals",
			numPoints: 6,
			interval:  timeInterval{hour, 2},
			startTime: startTime,
			want: autogold.Expect([]string{
				"2021-11-30 14:00:00 +0000 UTC",
				"2021-11-30 16:00:00 +0000 UTC",
				"2021-11-30 18:00:00 +0000 UTC",
				"2021-11-30 20:00:00 +0000 UTC",
				"2021-11-30 22:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			name:      "6 points 1 year intervals",
			numPoints: 6,
			interval:  timeInterval{year, 1},
			startTime: startTime,
			want: autogold.Expect([]string{
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
		t.Run(tc.name, func(t *testing.T) {
			tc.want.Equal(t, buildFrameTest(tc.numPoints, tc.interval, tc.startTime))
		})
	}
}

func Test_buildRecordingTimesBetween(t *testing.T) {
	fromTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name     string
		from     time.Time
		to       time.Time
		interval timeInterval
		want     autogold.Value
	}{
		{
			"one month at 1 week intervals",
			fromTime,
			fromTime.AddDate(0, 1, 0),
			timeInterval{week, 1},
			autogold.Expect([]string{
				"2021-01-01 00:00:00 +0000 UTC",
				"2021-01-08 00:00:00 +0000 UTC",
				"2021-01-15 00:00:00 +0000 UTC",
				"2021-01-22 00:00:00 +0000 UTC",
				"2021-01-29 00:00:00 +0000 UTC",
			}),
		},
		{
			"6 months at 4 week intervals",
			fromTime,
			fromTime.AddDate(0, 6, 0),
			timeInterval{week, 4},
			autogold.Expect([]string{
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
		t.Run(tc.name, func(t *testing.T) {
			got := buildRecordingTimesBetween(tc.from, tc.to, tc.interval)
			tc.want.Equal(t, convert(got))
		})
	}
}

func Test_withinHalfAnInterval(t *testing.T) {
	someInterval := timeInterval{month, 1}
	testCases := []struct {
		name              string
		possiblyLaterTime time.Time
		earlierTime       time.Time
		interval          timeInterval
		want              autogold.Value
	}{
		{
			name:              "time is close after existing",
			possiblyLaterTime: time.Date(2022, 11, 23, 12, 10, 0, 0, time.UTC),
			earlierTime:       time.Date(2022, 11, 22, 12, 10, 0, 0, time.UTC),
			interval:          someInterval,
			want:              autogold.Expect(true),
		},
		{
			name:              "time is before existing",
			possiblyLaterTime: time.Date(2022, 11, 21, 12, 10, 0, 0, time.UTC),
			earlierTime:       time.Date(2022, 11, 22, 12, 10, 0, 0, time.UTC),
			interval:          someInterval,
			want:              autogold.Expect(false),
		},
		{
			name:              "hourly interval has smaller allowance",
			possiblyLaterTime: time.Date(2022, 11, 22, 13, 10, 0, 0, time.UTC),
			earlierTime:       time.Date(2022, 11, 22, 12, 0, 0, 0, time.UTC),
			interval:          timeInterval{"hour", 2},
			want:              autogold.Expect(false),
		},
		{
			name:              "later time is too far ahead",
			possiblyLaterTime: time.Date(2022, 12, 29, 12, 10, 0, 0, time.UTC),
			earlierTime:       time.Date(2022, 11, 22, 13, 10, 0, 0, time.UTC),
			interval:          someInterval,
			want:              autogold.Expect(false),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := withinHalfAnInterval(tc.earlierTime, tc.possiblyLaterTime, tc.interval)
			tc.want.Equal(t, got)
		})
	}
}

func TestBuildAndCalculate(t *testing.T) {
	// This is a unit test based on a real insight to double-check a migration would return the desired times.

	// 2022-08-04 20:10:04.573293
	createdAt := time.Date(2022, 8, 4, 20, 10, 4, 573293, time.UTC)
	// 2022-10-27 22:25:25.411322
	lastRecordedAt := time.Date(2022, 10, 28, 22, 25, 25, 411322, time.UTC)

	interval := timeInterval{week, 3}

	existingPoints := []time.Time{
		time.Date(2021, 12, 16, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 1, 6, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 1, 27, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 2, 17, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 3, 31, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 4, 21, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 5, 12, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 6, 2, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 6, 23, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 7, 14, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 8, 4, 0, 0, 0, 0, time.UTC),
		time.Date(2022, 8, 25, 21, 7, 55, 558154, time.UTC),
		time.Date(2022, 9, 15, 21, 12, 38, 902339, time.UTC),
		time.Date(2022, 10, 06, 00, 00, 00, 00, time.UTC),
		time.Date(2022, 10, 27, 22, 25, 30, 390084, time.UTC),
	}

	calculated := calculateRecordingTimes(createdAt, lastRecordedAt, interval, existingPoints)
	autogold.Expect(convert(existingPoints)).Equal(t, convert(calculated))
}

func convert(times []time.Time) []string {
	var got []string
	for _, result := range times {
		got = append(got, result.String())
	}
	return got
}
