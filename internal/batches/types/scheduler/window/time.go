package window

import "time"

type timeOfDay struct {
	hour   int8
	minute int8

	cmp int
}

func timeOfDayFromParts(hour, minute int8) timeOfDay {
	return timeOfDay{
		hour:   hour,
		minute: minute,
		cmp:    int(hour)*60 + int(minute),
	}
}

func timeOfDayFromTime(t time.Time) timeOfDay {
	return timeOfDayFromParts(int8(t.Hour()), int8(t.Minute()))
}

func (t timeOfDay) after(other timeOfDay) bool {
	return t.cmp > other.cmp
}

func (t timeOfDay) before(other timeOfDay) bool {
	return t.cmp < other.cmp
}
