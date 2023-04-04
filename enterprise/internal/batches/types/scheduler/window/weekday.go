package window

import "time"

// weekdaySet represents a set of weekdays. As a special case, if no weekdays
// are set (ie the default value), then _all_ weekdays are considered to be set;
// there's no concept of a zero weekdaySet, since a rollout window must always
// be valid for at least one weekday.
//
// In terms of the implementation, since there are only seven possible weekdays,
// we can store them as bits in an int8.
type weekdaySet int8

// newWeekdaySet instantiates a new weekdaySet and returns it. If one or more
// days are provided, they will be added to the initial state of the set.
func newWeekdaySet(days ...time.Weekday) weekdaySet {
	var ws weekdaySet
	for _, day := range days {
		ws.add(day)
	}

	return ws
}

// add adds a day to the weekdaySet.
func (ws *weekdaySet) add(day time.Weekday) {
	*ws |= weekdayToBit(day)
}

// all returns true if the weekdaySet matches all days.
func (ws weekdaySet) all() bool {
	return ws == 0 || ws == 127
}

// includes returns true if the given day is included in the weekdaySet.
func (ws weekdaySet) includes(day time.Weekday) bool {
	if ws.all() {
		return true
	}

	return (ws & weekdayToBit(day)) != 0
}

func weekdayToBit(day time.Weekday) weekdaySet {
	// We're relying on the internal representation of Go's time.Weekday type
	// here: values are in the range [0, 6], per the Go documentation. This
	// should be stable, since it's documented, but we're obviously in trouble
	// should that ever change! (There is a sanity check for this in the unit
	// tests.)
	return weekdaySet(int8(1) << day)
}
