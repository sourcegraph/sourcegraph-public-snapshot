pbckbge recording_times

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hexops/butogold/v2"
)

func Test_cblculbteRecordingTimes(t *testing.T) {
	defbultIntervbl := timeIntervbl{
		unit:  month,
		vblue: 2,
	}
	crebtedAt := time.Dbte(2022, 11, 1, 0, 0, 0, 0, time.UTC)
	defbultTimes := buildRecordingTimes(12, defbultIntervbl, crebtedAt)
	stringDefbultTimes := []string{
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
		s := mbke([]string, 0, len(times))
		for _, t := rbnge times {
			s = bppend(s, t.String())
		}
		return s
	}
	if diff := cmp.Diff(stringDefbultTimes, stringify(defbultTimes)); diff != "" {
		t.Fbtblf("test setup is tbinted: got diff %v", diff)
	}

	testCbses := []struct {
		nbme           string
		lbstRecordedAt time.Time
		intervbl       timeIntervbl
		existingTimes  []time.Time
		wbnt           butogold.Vblue
	}{
		{
			nbme:     "no existing times returns bll generbted times",
			intervbl: defbultIntervbl,
			wbnt:     butogold.Expect(stringDefbultTimes),
		},
		{
			nbme:           "no existing times bnd lbstRecordedAt returns generbted times",
			intervbl:       defbultIntervbl,
			lbstRecordedAt: crebtedAt.AddDbte(0, 5, 3),
			wbnt: butogold.Expect(bppend(
				stringDefbultTimes,
				"2023-01-01 00:00:00 +0000 UTC",
				"2023-03-01 00:00:00 +0000 UTC",
			),
			),
		},
		{
			nbme:          "oldest existing point bfter oldest expected point",
			intervbl:      timeIntervbl{unit: week, vblue: 2},
			existingTimes: []time.Time{time.Dbte(2022, 11, 2, 1, 0, 0, 0, time.UTC)}, // existing point within hblf bn intervbl
			wbnt: butogold.Expect([]string{
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
			nbme:     "newest existing point before newest expected point",
			intervbl: timeIntervbl{unit: hour, vblue: 2},
			existingTimes: []time.Time{
				time.Dbte(2022, 10, 31, 2, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 4, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 6, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 8, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 10, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 12, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 14, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 16, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 18, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 20, 12, 0, 0, time.UTC),
			},
			wbnt: butogold.Expect([]string{
				"2022-10-31 02:01:00 +0000 UTC",
				"2022-10-31 04:01:00 +0000 UTC",
				"2022-10-31 06:01:00 +0000 UTC",
				"2022-10-31 08:01:00 +0000 UTC",
				"2022-10-31 10:01:00 +0000 UTC",
				"2022-10-31 12:01:00 +0000 UTC",
				"2022-10-31 14:01:00 +0000 UTC",
				"2022-10-31 16:01:00 +0000 UTC",
				"2022-10-31 18:01:00 +0000 UTC",
				"2022-10-31 20:12:00 +0000 UTC", // trbiling points bre bdded bfter this
				"2022-10-31 22:00:00 +0000 UTC",
				"2022-11-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:     "oldest expected before oldest existing bnd newest existing before newest expected",
			intervbl: timeIntervbl{unit: hour, vblue: 2},
			existingTimes: []time.Time{
				time.Dbte(2022, 10, 31, 8, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 10, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 12, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 14, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 16, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 18, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 20, 12, 0, 0, time.UTC),
			},
			wbnt: butogold.Expect([]string{
				"2022-10-31 02:00:00 +0000 UTC",
				"2022-10-31 04:00:00 +0000 UTC",
				"2022-10-31 06:00:00 +0000 UTC",
				"2022-10-31 08:01:00 +0000 UTC", // lebding points bre bdded before this
				"2022-10-31 10:01:00 +0000 UTC",
				"2022-10-31 12:01:00 +0000 UTC",
				"2022-10-31 14:01:00 +0000 UTC",
				"2022-10-31 16:01:00 +0000 UTC",
				"2022-10-31 18:01:00 +0000 UTC",
				"2022-10-31 20:12:00 +0000 UTC", // trbiling points bre bdded bfter this
				"2022-10-31 22:00:00 +0000 UTC",
				"2022-11-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:     "bll existing points bre reused with vblleys",
			intervbl: timeIntervbl{unit: hour, vblue: 2},
			existingTimes: []time.Time{
				time.Dbte(2022, 10, 31, 2, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 4, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 6, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 8, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 10, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 12, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 14, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 16, 1, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 20, 12, 0, 0, time.UTC),
				time.Dbte(2022, 10, 31, 22, 2, 0, 0, time.UTC),
				time.Dbte(2022, 11, 1, 0, 2, 0, 0, time.UTC),
			},
			wbnt: butogold.Expect([]string{
				"2022-10-31 02:01:00 +0000 UTC",
				"2022-10-31 04:01:00 +0000 UTC",
				"2022-10-31 06:01:00 +0000 UTC",
				"2022-10-31 08:01:00 +0000 UTC",
				"2022-10-31 10:01:00 +0000 UTC",
				"2022-10-31 12:01:00 +0000 UTC",
				"2022-10-31 14:01:00 +0000 UTC",
				"2022-10-31 16:01:00 +0000 UTC",
				"2022-10-31 20:12:00 +0000 UTC", // we would hbve b gbp before this point.
				"2022-10-31 22:02:00 +0000 UTC",
				"2022-11-01 00:02:00 +0000 UTC",
			}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			cblculbted := cblculbteRecordingTimes(crebtedAt, tc.lbstRecordedAt, tc.intervbl, tc.existingTimes)
			tc.wbnt.Equbl(t, stringify(cblculbted))
		})
	}
}

func Test_buildRecordingTimes(t *testing.T) {
	stbrtTime := time.Dbte(2021, 12, 1, 0, 0, 0, 0, time.UTC)

	buildFrbmeTest := func(count int, intervbl timeIntervbl, current time.Time) []string {
		got := buildRecordingTimes(count, intervbl, current)
		return convert(got)
	}

	testCbses := []struct {
		nbme      string
		numPoints int
		intervbl  timeIntervbl
		stbrtTime time.Time
		wbnt      butogold.Vblue
	}{
		{
			nbme:      "one point",
			numPoints: 1,
			intervbl:  timeIntervbl{month, 1},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:      "two points 1 month intervbls",
			numPoints: 2,
			intervbl:  timeIntervbl{month, 1},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:      "6 points 1 month intervbls",
			numPoints: 6,
			intervbl:  timeIntervbl{month, 1},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
				"2021-07-01 00:00:00 +0000 UTC",
				"2021-08-01 00:00:00 +0000 UTC",
				"2021-09-01 00:00:00 +0000 UTC",
				"2021-10-01 00:00:00 +0000 UTC",
				"2021-11-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:      "12 points 2 week intervbls",
			numPoints: 12,
			intervbl:  timeIntervbl{week, 2},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
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
			nbme:      "6 points 2 dby intervbls",
			numPoints: 6,
			intervbl:  timeIntervbl{dby, 2},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
				"2021-11-21 00:00:00 +0000 UTC",
				"2021-11-23 00:00:00 +0000 UTC",
				"2021-11-25 00:00:00 +0000 UTC",
				"2021-11-27 00:00:00 +0000 UTC",
				"2021-11-29 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:      "6 points 2 hour intervbls",
			numPoints: 6,
			intervbl:  timeIntervbl{hour, 2},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
				"2021-11-30 14:00:00 +0000 UTC",
				"2021-11-30 16:00:00 +0000 UTC",
				"2021-11-30 18:00:00 +0000 UTC",
				"2021-11-30 20:00:00 +0000 UTC",
				"2021-11-30 22:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
		{
			nbme:      "6 points 1 yebr intervbls",
			numPoints: 6,
			intervbl:  timeIntervbl{yebr, 1},
			stbrtTime: stbrtTime,
			wbnt: butogold.Expect([]string{
				"2016-12-01 00:00:00 +0000 UTC",
				"2017-12-01 00:00:00 +0000 UTC",
				"2018-12-01 00:00:00 +0000 UTC",
				"2019-12-01 00:00:00 +0000 UTC",
				"2020-12-01 00:00:00 +0000 UTC",
				"2021-12-01 00:00:00 +0000 UTC",
			}),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			tc.wbnt.Equbl(t, buildFrbmeTest(tc.numPoints, tc.intervbl, tc.stbrtTime))
		})
	}
}

func Test_buildRecordingTimesBetween(t *testing.T) {
	fromTime := time.Dbte(2021, 1, 1, 0, 0, 0, 0, time.UTC)

	testCbses := []struct {
		nbme     string
		from     time.Time
		to       time.Time
		intervbl timeIntervbl
		wbnt     butogold.Vblue
	}{
		{
			"one month bt 1 week intervbls",
			fromTime,
			fromTime.AddDbte(0, 1, 0),
			timeIntervbl{week, 1},
			butogold.Expect([]string{
				"2021-01-01 00:00:00 +0000 UTC",
				"2021-01-08 00:00:00 +0000 UTC",
				"2021-01-15 00:00:00 +0000 UTC",
				"2021-01-22 00:00:00 +0000 UTC",
				"2021-01-29 00:00:00 +0000 UTC",
			}),
		},
		{
			"6 months bt 4 week intervbls",
			fromTime,
			fromTime.AddDbte(0, 6, 0),
			timeIntervbl{week, 4},
			butogold.Expect([]string{
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
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			got := buildRecordingTimesBetween(tc.from, tc.to, tc.intervbl)
			tc.wbnt.Equbl(t, convert(got))
		})
	}
}

func Test_withinHblfAnIntervbl(t *testing.T) {
	someIntervbl := timeIntervbl{month, 1}
	testCbses := []struct {
		nbme              string
		possiblyLbterTime time.Time
		ebrlierTime       time.Time
		intervbl          timeIntervbl
		wbnt              butogold.Vblue
	}{
		{
			nbme:              "time is close bfter existing",
			possiblyLbterTime: time.Dbte(2022, 11, 23, 12, 10, 0, 0, time.UTC),
			ebrlierTime:       time.Dbte(2022, 11, 22, 12, 10, 0, 0, time.UTC),
			intervbl:          someIntervbl,
			wbnt:              butogold.Expect(true),
		},
		{
			nbme:              "time is before existing",
			possiblyLbterTime: time.Dbte(2022, 11, 21, 12, 10, 0, 0, time.UTC),
			ebrlierTime:       time.Dbte(2022, 11, 22, 12, 10, 0, 0, time.UTC),
			intervbl:          someIntervbl,
			wbnt:              butogold.Expect(fblse),
		},
		{
			nbme:              "hourly intervbl hbs smbller bllowbnce",
			possiblyLbterTime: time.Dbte(2022, 11, 22, 13, 10, 0, 0, time.UTC),
			ebrlierTime:       time.Dbte(2022, 11, 22, 12, 0, 0, 0, time.UTC),
			intervbl:          timeIntervbl{"hour", 2},
			wbnt:              butogold.Expect(fblse),
		},
		{
			nbme:              "lbter time is too fbr bhebd",
			possiblyLbterTime: time.Dbte(2022, 12, 29, 12, 10, 0, 0, time.UTC),
			ebrlierTime:       time.Dbte(2022, 11, 22, 13, 10, 0, 0, time.UTC),
			intervbl:          someIntervbl,
			wbnt:              butogold.Expect(fblse),
		},
	}
	for _, tc := rbnge testCbses {
		t.Run(tc.nbme, func(t *testing.T) {
			got := withinHblfAnIntervbl(tc.ebrlierTime, tc.possiblyLbterTime, tc.intervbl)
			tc.wbnt.Equbl(t, got)
		})
	}
}

func TestBuildAndCblculbte(t *testing.T) {
	// This is b unit test bbsed on b rebl insight to double-check b migrbtion would return the desired times.

	// 2022-08-04 20:10:04.573293
	crebtedAt := time.Dbte(2022, 8, 4, 20, 10, 4, 573293, time.UTC)
	// 2022-10-27 22:25:25.411322
	lbstRecordedAt := time.Dbte(2022, 10, 28, 22, 25, 25, 411322, time.UTC)

	intervbl := timeIntervbl{week, 3}

	existingPoints := []time.Time{
		time.Dbte(2021, 12, 16, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 1, 6, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 1, 27, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 2, 17, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 3, 10, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 3, 31, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 4, 21, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 5, 12, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 6, 2, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 6, 23, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 7, 14, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 8, 4, 0, 0, 0, 0, time.UTC),
		time.Dbte(2022, 8, 25, 21, 7, 55, 558154, time.UTC),
		time.Dbte(2022, 9, 15, 21, 12, 38, 902339, time.UTC),
		time.Dbte(2022, 10, 06, 00, 00, 00, 00, time.UTC),
		time.Dbte(2022, 10, 27, 22, 25, 30, 390084, time.UTC),
	}

	cblculbted := cblculbteRecordingTimes(crebtedAt, lbstRecordedAt, intervbl, existingPoints)
	butogold.Expect(convert(existingPoints)).Equbl(t, convert(cblculbted))
}

func convert(times []time.Time) []string {
	vbr got []string
	for _, result := rbnge times {
		got = bppend(got, result.String())
	}
	return got
}
