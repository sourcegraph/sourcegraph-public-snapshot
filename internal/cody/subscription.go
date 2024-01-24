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
const CODY_PRO_TRIAL_ENDED = "cody-pro-trial-ended"

const SAMS_SERVICE_ID = "https://accounts.sgdev.org"
const SAMS_SERVICE_TYPE = "openidconnect"

type UserSubscriptionPlan string

const (
	UserSubscriptionPlanFree UserSubscriptionPlan = "free"
	UserSubscriptionPlanPro  UserSubscriptionPlan = "pro"
)

type UserSubscription struct {
	// Status is the current status of the subscription. "pending" means the user has no Cody Pro subscription.
	// "pending" subscription will be removed post Feb 15, 2024. It is required to support users who have opted for
	// Cody Pro Trial on dotcom, but have not entered payment information on SSC.
	// (So they don't have an SSC backed subscription, but we need to act like they do, until 2/15.)
	Status ssc.SubscriptionStatus
	// Plan is the plan the user is subscribed to.
	Plan UserSubscriptionPlan
	// ApplyProRateLimits indicates the user should be given higher rate limits
	// for Cody and related functionality. (Use this value instead of checking
	// the subscription status for simplicity.)
	ApplyProRateLimits bool
	// CurrentPeriodStartAt is the start date of the current billing cycle.
	CurrentPeriodStartAt time.Time
	// CurrentPeriodEndAt is the end date of the current billing cycle.
	// IMPORTANT: This may be IN THE PAST. e.g. if the subscription was
	// canceled, this will be when the subscription ended.
	CurrentPeriodEndAt time.Time
}

// consolidateSubscriptionDetails merges the subscription data available on dotcom and SCC.
// This is needed while transitioning to use SCC as the source of truth for all subscription data. (Which should happen ~Q1/2024.)
// TODO[sourcegraph#59785]: Update dotcom to use SSC as the source of truth for all subscription data.
func consolidateSubscriptionDetails(ctx context.Context, user types.User, subscription *ssc.Subscription) (*UserSubscription, error) {
	// If subscription information is available from SSC, we use that.
	// And just ignore what is stored in dotcom. (Since they've already
	// been migrated so to speak.)
	if subscription != nil {
		currentPeriodStart, err := time.Parse(time.RFC3339, subscription.CurrentPeriodStart)
		if err != nil {
			return nil, err
		}

		currentPeriodEnd, err := time.Parse(time.RFC3339, subscription.CurrentPeriodEnd)
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

	// If the user doesn't have a subscription in the SSC backend, then we need
	// synthesize one using the data available on dotcom.
	currentPeriodStartAt, currentPeriodEndAt := PreSSCReleaseCurrentPeriodDateRange(ctx, user)

	if user.CodyProEnabledAt != nil {
		return &UserSubscription{
			Status:               ssc.SubscriptionStatusPending,
			Plan:                 UserSubscriptionPlanPro,
			ApplyProRateLimits:   !featureflag.FromContext(ctx).GetBoolOr(CODY_PRO_TRIAL_ENDED, false),
			CurrentPeriodStartAt: currentPeriodStartAt,
			CurrentPeriodEndAt:   currentPeriodEndAt,
		}, nil
	}

	return &UserSubscription{
		Status:               ssc.SubscriptionStatusPending,
		Plan:                 UserSubscriptionPlanFree,
		ApplyProRateLimits:   false,
		CurrentPeriodStartAt: currentPeriodStartAt,
		CurrentPeriodEndAt:   currentPeriodEndAt,
	}, nil
}

// getSAMSAccountIDForUser returns the user's SAMS account ID from users_external_accounts table.
func getSAMSAccountIDForUser(ctx context.Context, db database.DB, user types.User) (string, error) {
	// TODO(sourcegraph#59786): make service_id configurable between accounts.sourcegraph.com and accounts.sgdev.org using a feature flag for testing.
	accounts, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
		UserID:      user.ID,
		ServiceType: SAMS_SERVICE_TYPE,
		ServiceID:   SAMS_SERVICE_ID,
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

func getSubscriptionForUser(ctx context.Context, db database.DB, sscClient ssc.Client, user types.User) (*UserSubscription, error) {
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
