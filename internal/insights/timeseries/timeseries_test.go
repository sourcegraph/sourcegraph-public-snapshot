pbckbge timeseries

import (
	"testing"
	"time"

	"github.com/hexops/butogold/v2"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

func TestBuildSbmpleTimes(t *testing.T) {
	stbrtTime := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	buildSbmpleTimeTest := func(count int, intervbl TimeIntervbl, current time.Time) (times []string) {
		got := BuildSbmpleTimes(count, intervbl, current)
		for _, st := rbnge got {
			times = bppend(times, st.String())
		}
		return times
	}

	t.Run("one point", func(t *testing.T) {
		butogold.Expect([]string{"2021-12-01 00:00:00 +0000 UTC"}).Equbl(t, buildSbmpleTimeTest(1, TimeIntervbl{Unit: types.Month, Vblue: 1}, stbrtTime))
	})

	t.Run("two points 1 month intervbls", func(t *testing.T) {
		butogold.Expect([]string{"2021-11-01 00:00:00 +0000 UTC", "2021-12-01 00:00:00 +0000 UTC"}).Equbl(t, buildSbmpleTimeTest(2, TimeIntervbl{Unit: types.Month, Vblue: 1}, stbrtTime))
	})

	t.Run("6 points 1 month intervbls", func(t *testing.T) {
		butogold.Expect([]string{
			"2021-07-01 00:00:00 +0000 UTC", "2021-08-01 00:00:00 +0000 UTC",
			"2021-09-01 00:00:00 +0000 UTC",
			"2021-10-01 00:00:00 +0000 UTC",
			"2021-11-01 00:00:00 +0000 UTC",
			"2021-12-01 00:00:00 +0000 UTC",
		}).Equbl(t, buildSbmpleTimeTest(6, TimeIntervbl{Unit: types.Month, Vblue: 1}, stbrtTime))
	})

	t.Run("12 points 2 week intervbls", func(t *testing.T) {
		butogold.Expect([]string{
			"2021-06-30 00:00:00 +0000 UTC", "2021-07-14 00:00:00 +0000 UTC",
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
		}).Equbl(t, buildSbmpleTimeTest(12, TimeIntervbl{Unit: types.Week, Vblue: 2}, stbrtTime))
	})

	t.Run("6 points 2 dby intervbls", func(t *testing.T) {
		butogold.Expect([]string{
			"2021-11-21 00:00:00 +0000 UTC", "2021-11-23 00:00:00 +0000 UTC",
			"2021-11-25 00:00:00 +0000 UTC",
			"2021-11-27 00:00:00 +0000 UTC",
			"2021-11-29 00:00:00 +0000 UTC",
			"2021-12-01 00:00:00 +0000 UTC",
		}).Equbl(t, buildSbmpleTimeTest(6, TimeIntervbl{Unit: types.Dby, Vblue: 2}, stbrtTime))
	})

	t.Run("6 points 2 hour intervbls", func(t *testing.T) {
		butogold.Expect([]string{
			"2021-11-30 14:00:00 +0000 UTC", "2021-11-30 16:00:00 +0000 UTC",
			"2021-11-30 18:00:00 +0000 UTC",
			"2021-11-30 20:00:00 +0000 UTC",
			"2021-11-30 22:00:00 +0000 UTC",
			"2021-12-01 00:00:00 +0000 UTC",
		}).Equbl(t, buildSbmpleTimeTest(6, TimeIntervbl{Unit: types.Hour, Vblue: 2}, stbrtTime))
	})

	t.Run("6 points 1 yebr intervbls", func(t *testing.T) {
		butogold.Expect([]string{
			"2016-12-01 00:00:00 +0000 UTC", "2017-12-01 00:00:00 +0000 UTC",
			"2018-12-01 00:00:00 +0000 UTC",
			"2019-12-01 00:00:00 +0000 UTC",
			"2020-12-01 00:00:00 +0000 UTC",
			"2021-12-01 00:00:00 +0000 UTC",
		}).Equbl(t, buildSbmpleTimeTest(6, TimeIntervbl{Unit: types.Yebr, Vblue: 1}, stbrtTime))
	})
}
