package cody

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/ssc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const USE_SSC_FEATURE_FLAG = "use-ssc-for-cody-subscription"

type UserSubscriptionPlan string

const (
	UserSubscriptionPlanFree UserSubscriptionPlan = "free"
	UserSubscriptionPlanPro  UserSubscriptionPlan = "pro"
)

type UserSubscription struct {
	// Status is the current status of the subscription. "pending" means the user has no Cody Pro subscription.
	// "pending" subscription will be removed post Feb 15, 2024. It is required to support users who have opted for
	// Cody Pro Trial on dotcom.
	Status ssc.SubscriptionStatus
	// Plan is the plan the user is subscribed to. "free" or "pro".
	Plan UserSubscriptionPlan
	// ApplyProRateLimits is true if the user has a valid subscription status.
	ApplyProRateLimits bool
	// CurrentPeriodStartAt is the start date of the current billing cycle.
	CurrentPeriodStartAt time.Time
	// CurrentPeriodEndAt is the end date of the current billing cycle.
	CurrentPeriodEndAt time.Time
}

// consolidateSubscriptionDetails consolidates the subscription details from dotcom and SSC.
// Post Feb 15, 2024, this function will be removed and the logic will be simplified to us the SSC info only.
func consolidateSubscriptionDetails(ctx context.Context, user types.User, subscription *ssc.Subscription) (*UserSubscription, error) {
	now := currentTimeFromCtx(ctx)
	febReleaseDate := time.Date(2024, 02, 14, 23, 59, 59, 59, now.Location())
	isAfterFebReleaseDate := now.After(febReleaseDate)

	// The user has put in their cc and signed up for Cody Pro on SSC
	if subscription != nil {
		currentPeriodStart, err := time.ParseInLocation(time.RFC3339, subscription.CurrentPeriodStart, now.Location())
		if err != nil {
			return nil, err
		}

		currentPeriodEnd, err := time.ParseInLocation(time.RFC3339, subscription.CurrentPeriodEnd, now.Location())
		if err != nil {
			return nil, err
		}

		applyProRateLimits := subscription.Status == ssc.SubscriptionStatusActive || subscription.Status == ssc.SubscriptionStatusPastDue || subscription.Status == ssc.SubscriptionStatusTrialing

		return &UserSubscription{
			Status:               subscription.Status,
			Plan:                 UserSubscriptionPlanPro,
			ApplyProRateLimits:   applyProRateLimits,
			CurrentPeriodStartAt: currentPeriodStart,
			CurrentPeriodEndAt:   currentPeriodEnd,
		}, nil
	}

	currentPeriodStartAt, currentPeriodEndAt := PreSSCReleaseCurrentPeriodDateRange(ctx, user)

	// after the release, all the users will be on the free plan if they have no subscription on SSC
	if isAfterFebReleaseDate || user.CodyProEnabledAt == nil {
		return &UserSubscription{
			Status:               ssc.SubscriptionStatusPending,
			Plan:                 UserSubscriptionPlanFree,
			ApplyProRateLimits:   false,
			CurrentPeriodStartAt: currentPeriodStartAt,
			CurrentPeriodEndAt:   currentPeriodEndAt,
		}, nil
	}

	// NOTE: The following code is only temporary, and will be removed after Feb 15 2024.
	// After the release, the only source of truth for the subscription details will be SSC.
	// This will also allow us the remove the "pending" status from the SubscriptionStatus enum.

	// currentPeriodEndAt should not be after the release date for pro opted users without SSC subscription
	if currentPeriodEndAt.After(febReleaseDate) {
		currentPeriodEndAt = febReleaseDate
	}

	return &UserSubscription{
		Status:               ssc.SubscriptionStatusPending,
		Plan:                 UserSubscriptionPlanPro,
		ApplyProRateLimits:   true,
		CurrentPeriodStartAt: currentPeriodStartAt,
		CurrentPeriodEndAt:   currentPeriodEndAt,
	}, nil
}

// getSAMSAccountIDForUser returns the user's SAMS account ID from users_external_accounts table.
func getSAMSAccountIDForUser(ctx context.Context, db database.DB, user types.User) (string, error) {
	// TODO: make service_id configurable between accounts.sourcegraph.com and accounts.sgdev.org using a feature flag for testing.
	accounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		UserID:      user.ID,
		ServiceType: "openidconnect",
		ServiceID:   "https://accounts.sourcegraph.com",
		LimitOffset: &database.LimitOffset{
			Limit: 1,
		},
	})
	if err != nil {
		return "", errors.Wrap(err, "error while fetching user's external account")
	}

	if len(accounts) > 0 {
		return accounts[0].AccountID, nil
	}

	return "", nil
}

type SSCClient interface {
	FetchSubscriptionBySAMSAccountID(samsAccountID string) (*ssc.Subscription, error)
}

func getSubscriptionForUser(ctx context.Context, db database.DB, sscClient SSCClient, user types.User) (*UserSubscription, error) {
	samsAccountID, err := getSAMSAccountIDForUser(ctx, db, user)
	if err != nil {
		return nil, err
	}

	var subscription *ssc.Subscription

	// Fetch subscription from SSC only if the feature flag is enabled and the user has a SAMS account ID
	if samsAccountID != "" && featureflag.FromContext(ctx).GetBoolOr(USE_SSC_FEATURE_FLAG, false) {
		subscription, err = sscClient.FetchSubscriptionBySAMSAccountID(samsAccountID)
		if err != nil {
			return nil, errors.Wrap(err, "error while fetching user subscription from SSC")
		}
	}

	return consolidateSubscriptionDetails(ctx, user, subscription)
}

// SubscriptionForUser returns the user's Cody subscription details.
func SubscriptionForUser(ctx context.Context, db database.DB, user types.User) (*UserSubscription, error) {
	return getSubscriptionForUser(ctx, db, ssc.NewClient(), user)
}
