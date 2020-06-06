package repos

import (
	"testing"
	"time"
)

func TestIsSaturdayNight(t *testing.T) {
	cases := map[string]bool{
		"2012-11-01T22:08:41+00:00": false,
		"2012-11-03T22:08:41+00:00": true,

		// Boundary conditions
		"2012-11-03T21:59:59+00:00": false,
		"2012-11-03T22:00:00+00:00": true,
		"2012-11-03T22:59:59+00:00": true,
		"2012-11-03T23:00:00+00:00": false,

		// Not 10am
		"2012-11-03T10:05:00+00:00": false,

		// Time zone matters
		"2012-11-03T21:59:59+02:00": false,
		"2012-11-03T22:00:00+02:00": true,
	}
	for ts, want := range cases {
		tm, err := time.Parse(time.RFC3339, ts)
		if err != nil {
			t.Fatal(err)
		}
		if got := isSaturdayNight(tm); want != got {
			if got {
				t.Errorf("%s (%s) should not be saturday night", ts, tm.Format("Mon 15:04"))
			} else {
				t.Errorf("%s (%s) should be saturday night", ts, tm.Format("Mon 15:04"))
			}
		}
	}
}
