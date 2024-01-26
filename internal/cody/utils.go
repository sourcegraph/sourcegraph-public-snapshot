package cody

import (
	"context"
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

// fakeSubscriptionDateRange returns a fake date range for a Cody subscription.
// Since this isn't backed by Stripe or anything, the exact dates don't matter.
//
// If set, we peg the end date to February 15. (To align with the expected Cody
// Pro release, when we expect existing free trials to end.) Otherwise the end
// of the month is chosen.
func fakeSubscriptionBillingCycle(endOnFeb15 bool) (time.Time, time.Time) {
	now := time.Now()
	startOfTheMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Peg the end date to 2/15 unless it has already elapsed. (And hopefully
	// by then we'll be able to delete this function entirely.)
	feb15th2024 := time.Date(2024, time.February, 15, 12, 0, 0, 0, time.UTC)
	if endOnFeb15 && now.Before(feb15th2024) {
		return startOfTheMonth, feb15th2024
	}

	// This won't be pedantically correct in some situations, but for a stable
	// fake subscription billing period, it's sufficient.
	endOfTheMonth := time.Now().AddDate(0, 1, 0).UTC()
	return startOfTheMonth, endOfTheMonth
}
