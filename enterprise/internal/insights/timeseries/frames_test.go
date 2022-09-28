package timeseries

import (
	"testing"
	"time"

	"github.com/hexops/autogold"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

func TestBuildFrames(t *testing.T) {
	startTime := time.Date(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	convert := func(frames []types.Frame) [][]string {
		var got [][]string
		for _, result := range frames {
			got = append(got, []string{result.From.String(), result.To.String()})
		}
		return got
	}
	buildFrameTest := func(count int, interval TimeInterval, current time.Time) [][]string {
		got := BuildFrames(count, interval, current)
		return convert(got)
	}

	t.Run("one point", func(t *testing.T) {
		autogold.Want("one point", [][]string{{
			"2021-12-01 00:00:00 +0000 UTC",
			"2021-12-01 00:00:00 +0000 UTC",
		}}).Equal(t, buildFrameTest(1, TimeInterval{Unit: types.Month, Value: 1}, startTime))
	})

	t.Run("two points 1 month intervals", func(t *testing.T) {
		autogold.Want("two points 1 month intervals", [][]string{
			{
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
			{
				"2021-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
		}).Equal(t, buildFrameTest(2, TimeInterval{Unit: types.Month, Value: 1}, startTime))
	})

	t.Run("6 points 1 month intervals", func(t *testing.T) {
		autogold.Want("6 points 1 month intervals", [][]string{
			{
				"2021-07-01 00:00:00 +0000 UTC",
				"2021-08-01 00:00:00 +0000 UTC",
			},
			{
				"2021-08-01 00:00:00 +0000 UTC",
				"2021-09-01 00:00:00 +0000 UTC",
			},
			{
				"2021-09-01 00:00:00 +0000 UTC",
				"2021-10-01 00:00:00 +0000 UTC",
			},
			{
				"2021-10-01 00:00:00 +0000 UTC",
				"2021-11-01 00:00:00 +0000 UTC",
			},
			{
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
			{
				"2021-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
		}).Equal(t, buildFrameTest(6, TimeInterval{Unit: types.Month, Value: 1}, startTime))
	})

	t.Run("12 points 2 week intervals", func(t *testing.T) {
		autogold.Want("12 points 2 week intervals", [][]string{
			{
				"2021-06-30 00:00:00 +0000 UTC",
				"2021-07-14 00:00:00 +0000 UTC",
			},
			{
				"2021-07-14 00:00:00 +0000 UTC",
				"2021-07-28 00:00:00 +0000 UTC",
			},
			{
				"2021-07-28 00:00:00 +0000 UTC",
				"2021-08-11 00:00:00 +0000 UTC",
			},
			{
				"2021-08-11 00:00:00 +0000 UTC",
				"2021-08-25 00:00:00 +0000 UTC",
			},
			{
				"2021-08-25 00:00:00 +0000 UTC",
				"2021-09-08 00:00:00 +0000 UTC",
			},
			{
				"2021-09-08 00:00:00 +0000 UTC",
				"2021-09-22 00:00:00 +0000 UTC",
			},
			{
				"2021-09-22 00:00:00 +0000 UTC",
				"2021-10-06 00:00:00 +0000 UTC",
			},
			{
				"2021-10-06 00:00:00 +0000 UTC",
				"2021-10-20 00:00:00 +0000 UTC",
			},
			{
				"2021-10-20 00:00:00 +0000 UTC",
				"2021-11-03 00:00:00 +0000 UTC",
			},
			{
				"2021-11-03 00:00:00 +0000 UTC",
				"2021-11-17 00:00:00 +0000 UTC",
			},
			{
				"2021-11-17 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
			{
				"2021-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
		}).Equal(t, buildFrameTest(12, TimeInterval{Unit: types.Week, Value: 2}, startTime))
	})

	t.Run("6 points 2 day intervals", func(t *testing.T) {
		autogold.Want("6 points 2 day intervals", [][]string{
			{
				"2021-11-21 00:00:00 +0000 UTC",
				"2021-11-23 00:00:00 +0000 UTC",
			},
			{
				"2021-11-23 00:00:00 +0000 UTC",
				"2021-11-25 00:00:00 +0000 UTC",
			},
			{
				"2021-11-25 00:00:00 +0000 UTC",
				"2021-11-27 00:00:00 +0000 UTC",
			},
			{
				"2021-11-27 00:00:00 +0000 UTC",
				"2021-11-29 00:00:00 +0000 UTC",
			},
			{
				"2021-11-29 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
			{
				"2021-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
		}).Equal(t, buildFrameTest(6, TimeInterval{Unit: types.Day, Value: 2}, startTime))
	})

	t.Run("6 points 2 hour intervals", func(t *testing.T) {
		autogold.Want("6 points 2 hour intervals", [][]string{
			{
				"2021-11-30 14:00:00 +0000 UTC",
				"2021-11-30 16:00:00 +0000 UTC",
			},
			{
				"2021-11-30 16:00:00 +0000 UTC",
				"2021-11-30 18:00:00 +0000 UTC",
			},
			{
				"2021-11-30 18:00:00 +0000 UTC",
				"2021-11-30 20:00:00 +0000 UTC",
			},
			{
				"2021-11-30 20:00:00 +0000 UTC",
				"2021-11-30 22:00:00 +0000 UTC",
			},
			{
				"2021-11-30 22:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
			{
				"2021-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
		}).Equal(t, buildFrameTest(6, TimeInterval{Unit: types.Hour, Value: 2}, startTime))
	})

	t.Run("6 points 1 year intervals", func(t *testing.T) {
		autogold.Want("6 points 1 year intervals", [][]string{
			{
				"2016-12-01 00:00:00 +0000 UTC",
				"2017-12-01 00:00:00 +0000 UTC",
			},
			{
				"2017-12-01 00:00:00 +0000 UTC",
				"2018-12-01 00:00:00 +0000 UTC",
			},
			{
				"2018-12-01 00:00:00 +0000 UTC",
				"2019-12-01 00:00:00 +0000 UTC",
			},
			{
				"2019-12-01 00:00:00 +0000 UTC",
				"2020-12-01 00:00:00 +0000 UTC",
			},
			{
				"2020-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
			{
				"2021-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			},
		}).Equal(t, buildFrameTest(6, TimeInterval{Unit: types.Year, Value: 1}, startTime))
	})
}

func TestBuildFramesBetween(t *testing.T) {
	convert := func(frames []types.Frame) [][]string {
		var got [][]string
		for _, result := range frames {
			got = append(got, []string{result.From.String(), result.To.String()})
		}
		return got
	}

	fromTime := time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		from     time.Time
		to       time.Time
		interval TimeInterval
		want     autogold.Value
	}{
		{
			fromTime,
			fromTime.AddDate(0, 1, 0),
			TimeInterval{Unit: types.Week, Value: 1},
			autogold.Want("frames for one month at 1 week intervals", [][]string{
				{"2021-01-01 00:00:00 +0000 UTC", "2021-01-08 00:00:00 +0000 UTC"},
				{"2021-01-08 00:00:00 +0000 UTC", "2021-01-15 00:00:00 +0000 UTC"},
				{"2021-01-15 00:00:00 +0000 UTC", "2021-01-22 00:00:00 +0000 UTC"},
				{"2021-01-22 00:00:00 +0000 UTC", "2021-01-29 00:00:00 +0000 UTC"},
				{"2021-01-29 00:00:00 +0000 UTC", "2021-02-05 00:00:00 +0000 UTC"},
			}),
		},
		{
			fromTime,
			fromTime.AddDate(0, 6, 0),
			TimeInterval{Unit: types.Week, Value: 4},
			autogold.Want("frames for 6 months at 4 week intervals", [][]string{
				{"2021-01-01 00:00:00 +0000 UTC", "2021-01-29 00:00:00 +0000 UTC"},
				{"2021-01-29 00:00:00 +0000 UTC", "2021-02-26 00:00:00 +0000 UTC"},
				{"2021-02-26 00:00:00 +0000 UTC", "2021-03-26 00:00:00 +0000 UTC"},
				{"2021-03-26 00:00:00 +0000 UTC", "2021-04-23 00:00:00 +0000 UTC"},
				{"2021-04-23 00:00:00 +0000 UTC", "2021-05-21 00:00:00 +0000 UTC"},
				{"2021-05-21 00:00:00 +0000 UTC", "2021-06-18 00:00:00 +0000 UTC"},
				{"2021-06-18 00:00:00 +0000 UTC", "2021-07-16 00:00:00 +0000 UTC"},
			}),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want.Name(), func(t *testing.T) {
			got := BuildFramesBetween(tc.from, tc.to, tc.interval)
			tc.want.Equal(t, convert(got))
		})
	}
}
