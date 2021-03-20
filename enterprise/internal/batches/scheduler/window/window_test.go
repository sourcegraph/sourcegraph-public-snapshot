package window

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/schema"
)

func TestParseRateUnit(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for _, in := range []string{"", " ", "a"} {
			t.Run(in, func(t *testing.T) {
				if _, err := parseRateUnit(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := range []struct {
			inputs []string
			want   rateUnit
		}{
			{
				inputs: []string{"s", "S", "sec", "SEC", "secs", "SECS", "second", "SECOND", "seconds", "SECONDS"},
				want:   ratePerSecond,
			},
			{
				inputs: []string{"m", "M", "min", "MIN", "mins", "MINS", "minute", "MINUTE", "minutes", "MINUTES"},
				want:   ratePerMinute,
			},
			{
				inputs: []string{"h", "H", "hr", "HR", "hrs", "HRS", "hour", "HOUR", "hours", "HOURS"},
				want:   ratePerHour,
			},
		} {
			for _, in := range tc.inputs {
				t.Run(in, func(t *testing.T) {
					if have, err := parseRateUnit(in); err != nil {
						t.Errorf("unexpected error: %v", err)
					} else if have != tc.want {
						t.Errorf("unexpected rate: have=%v want=%v", have, tc.want)
					}
				})
			}
		}
	})
}

func TestParseRate(t *testing.T) {
	t.Run("errors", func(t *testing.T) {
		for name, in := range map[string]interface{}{
			"nil":                                nil,
			"non-zero int":                       1,
			"empty string":                       "",
			"string without slash":               "20",
			"string without a rate number":       "/min",
			"string with an invalid rate number": "n/min",
			"string with a negative rate number": "-1/min",
			"string with an invalid rate unit":   "20/year",
			"bool":                               true,
			"slice":                              []string{},
			"map":                                map[string]string{},
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := parseRate(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for name, tc := range map[string]struct {
			in   interface{}
			want rate
		}{
			"zero": {
				in:   0,
				want: rate{n: 0},
			},
			"unlimited": {
				in:   "unlimited",
				want: rate{n: -1},
			},
			"also unlimited": {
				in:   "UNLIMITED",
				want: rate{n: -1},
			},
			"valid rate": {
				in:   "20/min",
				want: rate{n: 20, unit: ratePerMinute},
			},
		} {
			t.Run(name, func(t *testing.T) {
				if have, err := parseRate(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want, cmpExport); diff != "" {
					t.Errorf("unexpected rate (-have +want):\n%s", diff)
				}
			})
		}
	})
}

func TestNewWindowTime(t *testing.T) {
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
				if _, err := newWindowTime(in); err == nil {
					t.Error("unexpected nil error")
				}
			})
		}
	})

	t.Run("success", func(t *testing.T) {
		for _, tc := range []struct {
			in   string
			want *windowTime
		}{
			{
				in:   "",
				want: nil,
			},
			{
				in:   "0:0",
				want: &windowTime{hour: 0, minute: 0},
			},
			{
				in:   "0:00",
				want: &windowTime{hour: 0, minute: 0},
			},
			{
				in:   "00:00",
				want: &windowTime{hour: 0, minute: 0},
			},
			{
				in:   "20:20",
				want: &windowTime{hour: 20, minute: 20},
			},
			{
				in:   "1:1",
				want: &windowTime{hour: 1, minute: 1},
			},
		} {
			t.Run(tc.in, func(t *testing.T) {
				if have, err := newWindowTime(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want, cmpExport); diff != "" {
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
					days: []time.Weekday{},
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
					days:  []time.Weekday{time.Monday, time.Tuesday},
					rate:  rate{n: 20, unit: ratePerMinute},
					start: &windowTime{hour: 1, minute: 15},
					end:   &windowTime{hour: 2, minute: 30},
				},
			},
		} {
			t.Run(name, func(t *testing.T) {
				if have, err := parseWindow(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want, cmpExport); diff != "" {
					t.Errorf("unexpected window (-have +want):\n%s", diff)
				}
			})
		}
	})
}
