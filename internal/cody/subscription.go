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

const (
	// FeatureFlagCodyProTrialStillAvailable controls whether or not we grant users who have signed up for the "Cody Pro"
	// free trial access to Pro features. The default for this is true, but we flip it to false on February 15, 2024.
	// (And in a subsequent release, remove it and all references.)
	FeatureFlagCodyProTrialStillAvailable = "ssc.cody-pro-trial-still-available"

	// FeatureFlagReadCodySubscriptionFromSSC controls whether or not we should attempt to read a user's Cody Pro
	// subscription information from the SSC backend. Defaults to false while the SSC integration is in development.
	FeatureFlagReadCodySubscriptionFromSSC = "ssc.read-cody-subscription-from-ssc"
)

type UserSubscriptionPlan string

const (
	UserSubscriptionPlanFree UserSubscriptionPlan = "FREE"
	UserSubscriptionPlanPro  UserSubscriptionPlan = "PRO"
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

// grantsProRateLimits returns whether or not a user should be given the Cody Pro limits based
// on the provided subscription status.
func grantsProRateLimits(status ssc.SubscriptionStatus) bool {
	switch status {
	case ssc.SubscriptionStatusActive:
		return true
	case ssc.SubscriptionStatusTrialing:
		return true

	case ssc.SubscriptionStatusPastDue:
		// Subscriptions not in a payment-related error state are still
		// given the Pro rate limits. However, depending on how things
		// are configured, the subscription will automatically transition
		// to a "canceled" state soon.
		return true

	default:
		// All other states. e.g. "canceled", "unpaid", "incomplete", etc.
		return false
	}
}

// consolidateSubscriptionDetails merges the (optional) subscription data available on dotcom and SCC.
// This is needed while transitioning to use SCC as the source of truth for all subscription data.
// (Which should happen ~Q1/2024.)
// TODO[sourcegraph#59785]: Update dotcom to use SSC as the source of truth for all subscription data.
func consolidateSubscriptionDetails(ctx context.Context, user types.User, subscription *ssc.Subscription) (*UserSubscription, error) {
	// If subscription information is available from SSC, we use that.
	// And just ignore what is stored in dotcom.
	if subscription != nil {
		currentPeriodStart, err := time.Parse(time.RFC3339, subscription.CurrentPeriodStart)
		if err != nil {
			return nil, err
		}
		currentPeriodEnd, err := time.Parse(time.RFC3339, subscription.CurrentPeriodEnd)
		if err != nil {
			return nil, err
		}

		return &UserSubscription{
			Status:               subscription.Status,
			Plan:                 UserSubscriptionPlanPro,
			ApplyProRateLimits:   grantsProRateLimits(subscription.Status),
			CurrentPeriodStartAt: currentPeriodStart,
			CurrentPeriodEndAt:   currentPeriodEnd,
		}, nil
	}

	userSignedUpForFreeTrial := user.CodyProEnabledAt != nil
	periodStart, periodEnd := fakeSubscriptionBillingCycle(userSignedUpForFreeTrial /* endOnFeb15, since that is when the trial will end */)

	// Without subscription information from SCC, we synthetsize one.
	fauxSubscription := UserSubscription{
		Status:               ssc.SubscriptionStatusPending,
		Plan:                 UserSubscriptionPlanFree,
		ApplyProRateLimits:   false,
		CurrentPeriodStartAt: periodStart,
		CurrentPeriodEndAt:   periodEnd,
	}

	// Check if the user has signed up for the Cody Pro free trial.
	// We'll give them the benefits of a Cody Pro subscription until we end the trial offer.
	//
	// At that time, we will be pulling all subscription information from the SSC backend.
	coryProTrialStillAvailable := !featureflag.FromContext(ctx).GetBoolOr(FeatureFlagCodyProTrialStillAvailable, true)
	if userSignedUpForFreeTrial && coryProTrialStillAvailable {
		fauxSubscription.Plan = UserSubscriptionPlanPro
		fauxSubscription.ApplyProRateLimits = true
	}

	return &fauxSubscription, nil
}

// SubscriptionForUser returns the user's Cody subscription information.
// This will return a valid UserSubscription object, even if the user is on the Cody Free
func SubscriptionForUser(ctx context.Context, db database.DB, user types.User) (*UserSubscription, error) {
	var subscription *ssc.Subscription

	tryReadSSCSubscription := featureflag.FromContext(ctx).GetBoolOr(FeatureFlagReadCodySubscriptionFromSSC, false)
	if tryReadSSCSubscription {
		// The SSC client shouldn't be created on every call.
		sscClient := ssc.NewClient()

		// Lookup the SAMS account ID for this dotcom user. This isn't part of the SSC backend, but needs
		// special care in order to have it match the targeted SSC API.
		samsAccountID, err := sscClient.LookupDotcomUserSAMSAccountID(ctx, db, user.ID)
		if err != nil {
			return nil, errors.Wrap(err, "looking up SAMS account ID")
		}
		if samsAccountID == "" {
			// Do nothing. If the user has never logged in via SAMS, they will not have a SAMS identity
			// on their dotcom user account.
		} else {
			// If they have a SAMS ID, then fetch the subscription from the SSC backend.
			// The returned subscription may be nil if the user has not subscribed yet.
			subscription, err = sscClient.FetchSubscriptionBySAMSAccountID(ctx, samsAccountID)
			if err != nil {
				return nil, errors.Wrap(err, "fetching subscription from SSC")
			}
		}
	}

	return consolidateSubscriptionDetails(ctx, user, subscription)
}
