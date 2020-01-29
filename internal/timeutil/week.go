package timeutil

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

func startOfPeriod(periodType db.PeriodType, periodsAgo int) (time.Time, error) {
	switch periodType {
	case db.Daily:
		now := timeNow().UTC()
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -periodsAgo), nil
	case db.Weekly:
		return startOfWeek(periodsAgo), nil
	case db.Monthly:
		now := timeNow().UTC()
		return time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC).AddDate(0, -periodsAgo, 0), nil
	}
	return time.Time{}, fmt.Errorf("periodType must be \"daily\", \"weekly\", or \"monthly\". Got %s", periodType)
}

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
