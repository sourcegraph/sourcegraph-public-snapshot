package cody

import (
	"context"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"time"
)

type currentTimeCtxKey int

const mockCurrentTimeKey currentTimeCtxKey = iota

func currentTimeFromCtx(ctx context.Context) time.Time {
	t, ok := ctx.Value(mockCurrentTimeKey).(*time.Time)
	if !ok || t == nil {
		return time.Now()
	}
	return *t
}

func withCurrentTimeMock(ctx context.Context, t time.Time) context.Context {
	return context.WithValue(ctx, mockCurrentTimeKey, &t)
}

func PreSSCReleaseCurrentPeriodDateRange(ctx context.Context, user types.User) (time.Time, time.Time) {
	// to allow mocking current time during tests
	currentDate := currentTimeFromCtx(ctx)

	subscriptionStartDate := user.CreatedAt
	gaReleaseDate := time.Date(2023, 12, 14, 0, 0, 0, 0, subscriptionStartDate.Location())

	if !currentDate.Before(gaReleaseDate) && subscriptionStartDate.Before(gaReleaseDate) {
		subscriptionStartDate = gaReleaseDate
	}

	codyProEnabledAt := user.CodyProEnabledAt
	if codyProEnabledAt != nil {
		subscriptionStartDate = *codyProEnabledAt
	}

	targetDay := subscriptionStartDate.Day()
	startDayOfTheMonth := targetDay
	endDayOfTheMonth := targetDay - 1
	startMonth := currentDate
	endMonth := currentDate

	if currentDate.Day() < targetDay {
		// Set to target day of the previous month
		startMonth = currentDate.AddDate(0, -1, 0)
	} else {
		// Set to target day of the next month
		endMonth = currentDate.AddDate(0, 1, 0)
	}

	daysInStartingMonth := time.Date(startMonth.Year(), startMonth.Month()+1, 0, 0, 0, 0, 0, startMonth.Location()).Day()
	if startDayOfTheMonth > daysInStartingMonth {
		startDayOfTheMonth = daysInStartingMonth
	}

	daysInEndingMonth := time.Date(endMonth.Year(), endMonth.Month()+1, 0, 0, 0, 0, 0, endMonth.Location()).Day()
	if endDayOfTheMonth > daysInEndingMonth {
		endDayOfTheMonth = daysInEndingMonth
	}

	startDate := time.Date(startMonth.Year(), startMonth.Month(), startDayOfTheMonth, 0, 0, 0, 0, startMonth.Location())
	endDate := time.Date(endMonth.Year(), endMonth.Month(), endDayOfTheMonth, 23, 59, 59, 59, endMonth.Location())

	return startDate, endDate
}
