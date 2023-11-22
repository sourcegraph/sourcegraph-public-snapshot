package window

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestWindow_IsOpen(t *testing.T) {
	// A time that corresponds rather closely to when this was getting written.
	// Note that this is a Wednesday.
	//
	// We should be commended for not calling this variable whensday.
	when := time.Date(2021, 3, 24, 1, 39, 44, 0, time.UTC)

	for name, tc := range map[string]struct {
		want   bool
		at     time.Time
		window *Window
	}{
		"always open": {
			want: true,
			at:   when,
			window: &Window{
				days: newWeekdaySet(),
			},
		},
		"open on certain days; correct day": {
			want: true,
			at:   when,
			window: &Window{
				days: newWeekdaySet(time.Wednesday),
			},
		},
		"open on certain days; incorrect day": {
			want: false,
			at:   when,
			window: &Window{
				days: newWeekdaySet(time.Thursday),
			},
		},
		"open at certain times; correct time": {
			want: true,
			at:   when,
			window: &Window{
				days:  newWeekdaySet(),
				start: timeOfDayPtr(int8(1), 0),
				end:   timeOfDayPtr(int8(3), 0),
			},
		},
		"open at certain times; incorrect time": {
			want: false,
			at:   when,
			window: &Window{
				days:  newWeekdaySet(),
				start: timeOfDayPtr(int8(11), 0),
				end:   timeOfDayPtr(int8(13), 0),
			},
		},
		"open at certain days and times; correct day and time": {
			want: true,
			at:   when,
			window: &Window{
				days:  newWeekdaySet(time.Wednesday),
				start: timeOfDayPtr(int8(1), 0),
				end:   timeOfDayPtr(int8(3), 0),
			},
		},
		"open at certain days and times; correct day only": {
			want: false,
			at:   when,
			window: &Window{
				days:  newWeekdaySet(time.Wednesday),
				start: timeOfDayPtr(int8(11), 0),
				end:   timeOfDayPtr(int8(13), 0),
			},
		},
		"open at certain days and times; correct time only": {
			want: false,
			at:   when,
			window: &Window{
				days:  newWeekdaySet(time.Tuesday),
				start: timeOfDayPtr(int8(1), 0),
				end:   timeOfDayPtr(int8(3), 0),
			},
		},
		"open at certain days and times; everything is terrible": {
			want: false,
			at:   when,
			window: &Window{
				days:  newWeekdaySet(time.Sunday),
				start: timeOfDayPtr(int8(11), 0),
				end:   timeOfDayPtr(int8(13), 0),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if have := tc.window.IsOpen(tc.at); have != tc.want {
				t.Errorf("unexpected open state: have=%v want=%v", have, tc.want)
			}
		})
	}
}

func TestWindow_NextOpenAfter(t *testing.T) {
	// Please see TestWindow_IsOpen for the derivation of this pseudo-constant,
	// and a terrible pun.
	when := time.Date(2021, 3, 24, 1, 39, 44, 0, time.UTC)

	for name, tc := range map[string]struct {
		want   time.Time
		after  time.Time
		window *Window
	}{
		"currently open": {
			want:  when,
			after: when,
			window: &Window{
				days: newWeekdaySet(),
			},
		},
		"days only": {
			// Truncate back to the start of Wednesday, then add two days to get
			// to the start of Friday.
			want:  when.Truncate(24 * time.Hour).Add(2 * 24 * time.Hour),
			after: when,
			window: &Window{
				days: newWeekdaySet(time.Friday),
			},
		},
		"every day except Wednesday": {
			// Truncate back to the start of Wednesday, then add one day to get
			// to the start of Thursday.
			want:  when.Truncate(24 * time.Hour).Add(24 * time.Hour),
			after: when,
			window: &Window{
				days: newWeekdaySet(
					time.Sunday,
					time.Monday,
					time.Tuesday,
					time.Thursday,
					time.Friday,
					time.Saturday,
				),
			},
		},
		"times only": {
			// Truncate back to 00:00, then add 2 hours.
			want:  when.Truncate(24 * time.Hour).Add(2 * time.Hour),
			after: when,
			window: &Window{
				days:  newWeekdaySet(),
				start: timeOfDayPtr(int8(2), 0),
				end:   timeOfDayPtr(int8(4), 0),
			},
		},
		"time in the mysterious past": {
			// Truncate to 00:00, then add exactly one day and 30 minutes.
			want:  when.Truncate(24 * time.Hour).Add(24 * time.Hour).Add(30 * time.Minute),
			after: when,
			window: &Window{
				days:  newWeekdaySet(),
				start: timeOfDayPtr(int8(0), int8(30)),
				end:   timeOfDayPtr(int8(1), 0),
			},
		},
		"times and days": {
			// Truncate back to the start of Wednesday, then add five days to
			// get to the start of Monday (which also means we've wrapped around
			// Go's Weekday representation), then add another two hours to get
			// to 02:00.
			want:  when.Truncate(24 * time.Hour).Add(5 * 24 * time.Hour).Add(2 * time.Hour),
			after: when,
			window: &Window{
				days:  newWeekdaySet(time.Monday),
				start: timeOfDayPtr(int8(2), 0),
				end:   timeOfDayPtr(int8(4), 0),
			},
		},
	} {
		t.Run(name, func(t *testing.T) {
			if have := tc.window.NextOpenAfter(tc.after); have != tc.want {
				t.Errorf("unexpected next open time: have=%v want=%v", have, tc.want)
			}
		})
	}
}

func TestParseWindowTime(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for _, in := range []string{
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
				if _, err := parseWindowTime(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := range []struct {
			in   string
			want *timeOfDay
		}{
			{
				in:   "",
				want: nil,
			},
			{
				in:   "0:0",
				want: timeOfDayPtr(0, 0),
			},
			{
				in:   "0:00",
				want: timeOfDayPtr(0, 0),
			},
			{
				in:   "00:00",
				want: timeOfDayPtr(0, 0),
			},
			{
				in:   "20:20",
				want: timeOfDayPtr(20, 20),
			},
			{
				in:   "1:1",
				want: timeOfDayPtr(1, 1),
			},
		} {
			t.Run(tc.in, func(t *testing.T) {
				if have, err := parseWindowTime(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want, cmpOptions); diff != "" {
					t.Errorf("unexpected window time (-have +want)\n:%s", diff)
				}
			})
		}
	})
}

func TestParseWeekday(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for _, in := range []string{
			"",
			"su",
			"lunedi",
		} {
			t.Run(in, func(t *testing.T) {
				if _, err := parseWeekday(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := range []struct {
			inputs []string
			want   time.Weekday
		}{
			{
				inputs: []string{"sun", "Sun", "sunday", "Sunday"},
				want:   time.Sunday,
			},
			{
				inputs: []string{"mon", "Mon", "monday", "Monday"},
				want:   time.Monday,
			},
			{
				inputs: []string{"tue", "Tues", "tuesday", "Tuesday"},
				want:   time.Tuesday,
			},
			{
				inputs: []string{"wed", "Wed", "wednesday", "Wednesday"},
				want:   time.Wednesday,
			},
			{
				inputs: []string{"thu", "Thurs", "thursday", "Thursday"},
				want:   time.Thursday,
			},
			{
				inputs: []string{"fri", "Fri", "friday", "Friday"},
				want:   time.Friday,
			},
			{
				inputs: []string{"sat", "Sat", "saturday", "Saturday"},
				want:   time.Saturday,
			},
		} {
			for _, in := range tc.inputs {
				t.Run(in, func(t *testing.T) {
					if have, err := parseWeekday(in); err != nil {
						t.Errorf("unexpected error: %v", err)
					} else if have != tc.want {
						t.Errorf("unexpected weekday: have=%v want=%v", have, tc.want)
					}
				})
			}
		}
	})
}

func TestParseWindow(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		// We've just painstakingly tested all the other parsers above, so this
		// is just making sure each one is properly propagated when it returns
		// an error, rather than trying to be exhaustive.
		for name, in := range map[string]*schema.BatchChangeRolloutWindow{
			"nil window":         nil,
			"no rate":            {},
			"invalid weekday":    {Days: []string{"martedi"}},
			"invalid start time": {Start: "24:60"},
			"invalid end time":   {End: "24:60"},
			"invalid rate":       {Rate: "x/y"},
			"only start time":    {Start: "00:00"},
			"only end time":      {End: "00:00"},
			"start after end":    {Start: "01:00", End: "00:00"},
			"start equal to end": {Start: "01:00", End: "01:00"},
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := parseWindow(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   *schema.BatchChangeRolloutWindow
			want Window
		}{
			"rate only": {
				in: &schema.BatchChangeRolloutWindow{Rate: "unlimited"},
				want: Window{
					days: newWeekdaySet(),
					rate: rate{n: -1},
				},
			},
			"all fields": {
				in: &schema.BatchChangeRolloutWindow{
					Days:  []string{"monday", "tuesday"},
					Rate:  "20/min",
					Start: "01:15",
					End:   "02:30",
				},
				want: Window{
					days:  newWeekdaySet(time.Monday, time.Tuesday),
					rate:  rate{n: 20, unit: ratePerMinute},
					start: timeOfDayPtr(1, 15),
					end:   timeOfDayPtr(2, 30),
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				if have, err := parseWindow(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want, cmpOptions); diff != "" {
					t.Errorf("unexpected window (-have +want):\n%s", diff)
				}
			})
		}
	})
}
