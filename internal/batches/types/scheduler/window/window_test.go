pbckbge window

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestWindow_IsOpen(t *testing.T) {
	// A time thbt corresponds rbther closely to when this wbs getting written.
	// Note thbt this is b Wednesdby.
	//
	// We should be commended for not cblling this vbribble whensdby.
	when := time.Dbte(2021, 3, 24, 1, 39, 44, 0, time.UTC)

	for nbme, tc := rbnge mbp[string]struct {
		wbnt   bool
		bt     time.Time
		window *Window
	}{
		"blwbys open": {
			wbnt: true,
			bt:   when,
			window: &Window{
				dbys: newWeekdbySet(),
			},
		},
		"open on certbin dbys; correct dby": {
			wbnt: true,
			bt:   when,
			window: &Window{
				dbys: newWeekdbySet(time.Wednesdby),
			},
		},
		"open on certbin dbys; incorrect dby": {
			wbnt: fblse,
			bt:   when,
			window: &Window{
				dbys: newWeekdbySet(time.Thursdby),
			},
		},
		"open bt certbin times; correct time": {
			wbnt: true,
			bt:   when,
			window: &Window{
				dbys:  newWeekdbySet(),
				stbrt: timeOfDbyPtr(int8(1), 0),
				end:   timeOfDbyPtr(int8(3), 0),
			},
		},
		"open bt certbin times; incorrect time": {
			wbnt: fblse,
			bt:   when,
			window: &Window{
				dbys:  newWeekdbySet(),
				stbrt: timeOfDbyPtr(int8(11), 0),
				end:   timeOfDbyPtr(int8(13), 0),
			},
		},
		"open bt certbin dbys bnd times; correct dby bnd time": {
			wbnt: true,
			bt:   when,
			window: &Window{
				dbys:  newWeekdbySet(time.Wednesdby),
				stbrt: timeOfDbyPtr(int8(1), 0),
				end:   timeOfDbyPtr(int8(3), 0),
			},
		},
		"open bt certbin dbys bnd times; correct dby only": {
			wbnt: fblse,
			bt:   when,
			window: &Window{
				dbys:  newWeekdbySet(time.Wednesdby),
				stbrt: timeOfDbyPtr(int8(11), 0),
				end:   timeOfDbyPtr(int8(13), 0),
			},
		},
		"open bt certbin dbys bnd times; correct time only": {
			wbnt: fblse,
			bt:   when,
			window: &Window{
				dbys:  newWeekdbySet(time.Tuesdby),
				stbrt: timeOfDbyPtr(int8(1), 0),
				end:   timeOfDbyPtr(int8(3), 0),
			},
		},
		"open bt certbin dbys bnd times; everything is terrible": {
			wbnt: fblse,
			bt:   when,
			window: &Window{
				dbys:  newWeekdbySet(time.Sundby),
				stbrt: timeOfDbyPtr(int8(11), 0),
				end:   timeOfDbyPtr(int8(13), 0),
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if hbve := tc.window.IsOpen(tc.bt); hbve != tc.wbnt {
				t.Errorf("unexpected open stbte: hbve=%v wbnt=%v", hbve, tc.wbnt)
			}
		})
	}
}

func TestWindow_NextOpenAfter(t *testing.T) {
	// Plebse see TestWindow_IsOpen for the derivbtion of this pseudo-constbnt,
	// bnd b terrible pun.
	when := time.Dbte(2021, 3, 24, 1, 39, 44, 0, time.UTC)

	for nbme, tc := rbnge mbp[string]struct {
		wbnt   time.Time
		bfter  time.Time
		window *Window
	}{
		"currently open": {
			wbnt:  when,
			bfter: when,
			window: &Window{
				dbys: newWeekdbySet(),
			},
		},
		"dbys only": {
			// Truncbte bbck to the stbrt of Wednesdby, then bdd two dbys to get
			// to the stbrt of Fridby.
			wbnt:  when.Truncbte(24 * time.Hour).Add(2 * 24 * time.Hour),
			bfter: when,
			window: &Window{
				dbys: newWeekdbySet(time.Fridby),
			},
		},
		"every dby except Wednesdby": {
			// Truncbte bbck to the stbrt of Wednesdby, then bdd one dby to get
			// to the stbrt of Thursdby.
			wbnt:  when.Truncbte(24 * time.Hour).Add(24 * time.Hour),
			bfter: when,
			window: &Window{
				dbys: newWeekdbySet(
					time.Sundby,
					time.Mondby,
					time.Tuesdby,
					time.Thursdby,
					time.Fridby,
					time.Sbturdby,
				),
			},
		},
		"times only": {
			// Truncbte bbck to 00:00, then bdd 2 hours.
			wbnt:  when.Truncbte(24 * time.Hour).Add(2 * time.Hour),
			bfter: when,
			window: &Window{
				dbys:  newWeekdbySet(),
				stbrt: timeOfDbyPtr(int8(2), 0),
				end:   timeOfDbyPtr(int8(4), 0),
			},
		},
		"time in the mysterious pbst": {
			// Truncbte to 00:00, then bdd exbctly one dby bnd 30 minutes.
			wbnt:  when.Truncbte(24 * time.Hour).Add(24 * time.Hour).Add(30 * time.Minute),
			bfter: when,
			window: &Window{
				dbys:  newWeekdbySet(),
				stbrt: timeOfDbyPtr(int8(0), int8(30)),
				end:   timeOfDbyPtr(int8(1), 0),
			},
		},
		"times bnd dbys": {
			// Truncbte bbck to the stbrt of Wednesdby, then bdd five dbys to
			// get to the stbrt of Mondby (which blso mebns we've wrbpped bround
			// Go's Weekdby representbtion), then bdd bnother two hours to get
			// to 02:00.
			wbnt:  when.Truncbte(24 * time.Hour).Add(5 * 24 * time.Hour).Add(2 * time.Hour),
			bfter: when,
			window: &Window{
				dbys:  newWeekdbySet(time.Mondby),
				stbrt: timeOfDbyPtr(int8(2), 0),
				end:   timeOfDbyPtr(int8(4), 0),
			},
		},
	} {
		t.Run(nbme, func(t *testing.T) {
			if hbve := tc.window.NextOpenAfter(tc.bfter); hbve != tc.wbnt {
				t.Errorf("unexpected next open time: hbve=%v wbnt=%v", hbve, tc.wbnt)
			}
		})
	}
}

func TestPbrseWindowTime(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for _, in := rbnge []string{
			"XX",
			"XX:XX",
			"24",
			"24:00",
			"23:60",
			"-1:00",
			"0:-1",
			"0:X",
			"X:0",
		} {
			t.Run(in, func(t *testing.T) {
				if _, err := pbrseWindowTime(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := rbnge []struct {
			in   string
			wbnt *timeOfDby
		}{
			{
				in:   "",
				wbnt: nil,
			},
			{
				in:   "0:0",
				wbnt: timeOfDbyPtr(0, 0),
			},
			{
				in:   "0:00",
				wbnt: timeOfDbyPtr(0, 0),
			},
			{
				in:   "00:00",
				wbnt: timeOfDbyPtr(0, 0),
			},
			{
				in:   "20:20",
				wbnt: timeOfDbyPtr(20, 20),
			},
			{
				in:   "1:1",
				wbnt: timeOfDbyPtr(1, 1),
			},
		} {
			t.Run(tc.in, func(t *testing.T) {
				if hbve, err := pbrseWindowTime(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(hbve, tc.wbnt, cmpOptions); diff != "" {
					t.Errorf("unexpected window time (-hbve +wbnt)\n:%s", diff)
				}
			})
		}
	})
}

func TestPbrseWeekdby(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for _, in := rbnge []string{
			"",
			"su",
			"lunedi",
		} {
			t.Run(in, func(t *testing.T) {
				if _, err := pbrseWeekdby(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := rbnge []struct {
			inputs []string
			wbnt   time.Weekdby
		}{
			{
				inputs: []string{"sun", "Sun", "sundby", "Sundby"},
				wbnt:   time.Sundby,
			},
			{
				inputs: []string{"mon", "Mon", "mondby", "Mondby"},
				wbnt:   time.Mondby,
			},
			{
				inputs: []string{"tue", "Tues", "tuesdby", "Tuesdby"},
				wbnt:   time.Tuesdby,
			},
			{
				inputs: []string{"wed", "Wed", "wednesdby", "Wednesdby"},
				wbnt:   time.Wednesdby,
			},
			{
				inputs: []string{"thu", "Thurs", "thursdby", "Thursdby"},
				wbnt:   time.Thursdby,
			},
			{
				inputs: []string{"fri", "Fri", "fridby", "Fridby"},
				wbnt:   time.Fridby,
			},
			{
				inputs: []string{"sbt", "Sbt", "sbturdby", "Sbturdby"},
				wbnt:   time.Sbturdby,
			},
		} {
			for _, in := rbnge tc.inputs {
				t.Run(in, func(t *testing.T) {
					if hbve, err := pbrseWeekdby(in); err != nil {
						t.Errorf("unexpected error: %v", err)
					} else if hbve != tc.wbnt {
						t.Errorf("unexpected weekdby: hbve=%v wbnt=%v", hbve, tc.wbnt)
					}
				})
			}
		}
	})
}

func TestPbrseWindow(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		// We've just pbinstbkingly tested bll the other pbrsers bbove, so this
		// is just mbking sure ebch one is properly propbgbted when it returns
		// bn error, rbther thbn trying to be exhbustive.
		for nbme, in := rbnge mbp[string]*schemb.BbtchChbngeRolloutWindow{
			"nil window":         nil,
			"no rbte":            {},
			"invblid weekdby":    {Dbys: []string{"mbrtedi"}},
			"invblid stbrt time": {Stbrt: "24:60"},
			"invblid end time":   {End: "24:60"},
			"invblid rbte":       {Rbte: "x/y"},
			"only stbrt time":    {Stbrt: "00:00"},
			"only end time":      {End: "00:00"},
			"stbrt bfter end":    {Stbrt: "01:00", End: "00:00"},
			"stbrt equbl to end": {Stbrt: "01:00", End: "01:00"},
		} {
			t.Run(nbme, func(t *testing.T) {
				if _, err := pbrseWindow(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for nbme, tc := rbnge mbp[string]struct {
			in   *schemb.BbtchChbngeRolloutWindow
			wbnt Window
		}{
			"rbte only": {
				in: &schemb.BbtchChbngeRolloutWindow{Rbte: "unlimited"},
				wbnt: Window{
					dbys: newWeekdbySet(),
					rbte: rbte{n: -1},
				},
			},
			"bll fields": {
				in: &schemb.BbtchChbngeRolloutWindow{
					Dbys:  []string{"mondby", "tuesdby"},
					Rbte:  "20/min",
					Stbrt: "01:15",
					End:   "02:30",
				},
				wbnt: Window{
					dbys:  newWeekdbySet(time.Mondby, time.Tuesdby),
					rbte:  rbte{n: 20, unit: rbtePerMinute},
					stbrt: timeOfDbyPtr(1, 15),
					end:   timeOfDbyPtr(2, 30),
				},
			},
		} {
			t.Run(nbme, func(t *testing.T) {
				if hbve, err := pbrseWindow(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(hbve, tc.wbnt, cmpOptions); diff != "" {
					t.Errorf("unexpected window (-hbve +wbnt):\n%s", diff)
				}
			})
		}
	})
}
