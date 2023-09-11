package window

import (
	"fmt"
	"testing"
	"time"
)

var allWeekdays = []time.Weekday{
	time.Sunday,
	time.Monday,
	time.Tuesday,
	time.Wednesday,
	time.Thursday,
	time.Friday,
	time.Saturday,
}

// Equal is needed for test purposes, but not in normal use.
func (ws *weekdaySet) Equal(other *weekdaySet) bool {
	if ws != nil && other != nil {
		return *ws == *other
	}
	return false
}

func TestWeekday_All(t *testing.T) {
	for name, ws := range map[string]weekdaySet{
		"zero": newWeekdaySet(),
		"all":  newWeekdaySet(allWeekdays...),
	} {
		t.Run(name, func(t *testing.T) {
			if !ws.all() {
				t.Error("unexpected false return from all")
			}

			for _, day := range allWeekdays {
				if !ws.includes(day) {
					t.Errorf("day not included: %v", day)
				}
			}
		})
	}

	for i := 1; i < len(allWeekdays); i++ {
		t.Run(fmt.Sprintf("%d weekday(s)", i), func(t *testing.T) {
			ws := newWeekdaySet(allWeekdays[0:i]...)

			if ws.all() {
				t.Error("unexpected true return from all")
			}
		})
	}
}

func TestWeekday_Includes(t *testing.T) {
	for _, day := range allWeekdays {
		t.Run(day.String(), func(t *testing.T) {
			ws := newWeekdaySet(day)

			for _, check := range allWeekdays {
				if check == day {
					if !ws.includes(check) {
						t.Errorf("expected %v to be in set; it was not", check)
					}
				} else {
					if ws.includes(check) {
						t.Errorf("did not expect %v to be in set; it was", check)
					}
				}
			}
		})
	}
}

func TestWeekdayBitSanity(t *testing.T) {
	// This test exists solely as a safeguard in case Go ever changes the
	// internal representation of time.Weekday: it _should_ be covered by
	// semver, since it's documented, but there's no harm in being paranoid,
	// right?
	for _, day := range allWeekdays {
		if int(day) < 0 || int(day) > 6 {
			t.Errorf("unexpected Weekday value: %v", day)
		}
	}
}
