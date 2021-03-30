package window

import "time"

// weekdaySet represents a set of weekdays. If no weekdays are set, then _all_
// weekdays are considered to be set.
type weekdaySet struct {
	// d is used to encode the set of weekdays: since there are only seven
	// possible weekdays, we can store them as bits in an int8.
	d int8
}

// newWeekdaySet instantiates a new weekdaySet and returns it. If one or more
// days are provided, they will be added to the initial state of the set.
func newWeekdaySet(days ...time.Weekday) *weekdaySet {
	ws := &weekdaySet{}
	for _, day := range days {
		ws.add(day)
	}

	return ws
}

// add adds a day to the weekdaySet.
func (ws *weekdaySet) add(day time.Weekday) {
	ws.d |= weekdayToBit(day)
}

// all returns true if the weekdaySet matches all days.
func (ws *weekdaySet) all() bool {
	return ws.d == 0 || ws.d == 127
}

// includes returns true if the given day is included in the weekdaySet.
func (ws *weekdaySet) includes(day time.Weekday) bool {
	if ws.all() {
		return true
	}

	return (ws.d & weekdayToBit(day)) != 0
}

func (ws *weekdaySet) Equal(other *weekdaySet) bool {
	if ws != nil && other != nil {
		return ws.d == other.d
	}
	return false
}

func weekdayToBit(day time.Weekday) int8 {
	// We're relying on the internal representation of Go's time.Weekday type
	// here: values are in the range [0, 6], per the Go documentation. This
	// should be stable, since it's documented, but we're obviously in trouble
	// should that ever change! (There is a sanity check for this in the unit
	// tests.)
	return int8(1) << int(day)
}
