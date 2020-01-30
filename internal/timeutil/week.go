package timeutil

import "time"

func StartOfWeek(now time.Time, weeksAgo int) time.Time {
	if weeksAgo > 0 {
		return StartOfWeek(now, 0).AddDate(0, 0, -7*weeksAgo)
	}

	// If weeksAgo == 0, start at timeNow(), and loop back by day until we hit a Sunday
	date := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	for date.Weekday() != time.Sunday {
		date = date.AddDate(0, 0, -1)
	}
	return date
}
