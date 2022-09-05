package window

import (
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
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
		for name, in := range map[string]any{
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
			in   any
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
			"valid per-second rate": {
				in:   "20/sec",
				want: rate{n: 20, unit: ratePerSecond},
			},
			"valid per-minute rate": {
				in:   "20/min",
				want: rate{n: 20, unit: ratePerMinute},
			},
			"valid per-hour rate": {
				in:   "20/hr",
				want: rate{n: 20, unit: ratePerHour},
			},
		} {
			t.Run(name, func(t *testing.T) {
				if have, err := parseRate(tc.in); err != nil {
					t.Errorf("unexpected error: %v", err)
				} else if diff := cmp.Diff(have, tc.want, cmpOptions); diff != "" {
					t.Errorf("unexpected rate (-have +want):\n%s", diff)
				}
			})
		}
	})
}

func TestRateUnit_AsDuration(t *testing.T) {
	for in, want := range map[rateUnit]time.Duration{
		ratePerSecond: time.Second,
		ratePerMinute: time.Minute,
		ratePerHour:   time.Hour,
	} {
		t.Run(strconv.Itoa(int(in)), func(t *testing.T) {
			if have := in.AsDuration(); have != want {
				t.Errorf("unexpected duration: have=%v want=%v", have, want)
			}
		})
	}

	assert.Panics(t, func() {
		ru := rateUnit(4)
		ru.AsDuration()
	})
}
