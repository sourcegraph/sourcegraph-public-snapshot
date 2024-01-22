package ssc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const SSC_API_BASE_URL = "http://127.0.0.1:9982/cody"
const SSC_ADMINISTRATIVE_SECRET_TOKEN = "naman"

// FetchSubscriptionBySAMSAccountID returns the user's Cody subscription for the sams_account_id
func FetchSubscriptionBySAMSAccountID(samsAccountID string) (*SSCSubscription, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/api/rest/svc/subscription/%s", SSC_API_BASE_URL, samsAccountID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", SSC_ADMINISTRATIVE_SECRET_TOKEN))

	resp, err := httpcli.UncachedExternalDoer.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var subscription SSCSubscription
		err = json.Unmarshal(body, &subscription)
		if err != nil {
			return nil, err
		}

		subscription.SAMSAccountID = samsAccountID

		return &subscription, nil
	}

	// 204 response indicates that the user does not have a Cody Pro subscription
	if resp.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	return nil, errors.Errorf("unexpected status code %d while fetching user subscription from SSC", resp.StatusCode)
}

var febReleaseDate = time.Date(2024, 02, 15, 0, 0, 0, 0, time.Now().Location())

func ConsolidateSubscriptionDetails(ctx context.Context, user types.User, subscription *SSCSubscription) (*UserCodySubscription, error) {
	isBeforeFebReleaseDate := time.Now().Before(febReleaseDate)

	// The user has put in their cc and signed up for Cody Pro on SSC
	if subscription != nil {
		currentPeriodStart, err := time.Parse(time.RFC3339, subscription.CurrentPeriodStart)
		if err != nil {
			return nil, err
		}

		currentPeriodEnd, err := time.Parse(time.RFC3339, subscription.CurrentPeriodEnd)
		if err != nil {
			return nil, err
		}

		applyProRateLimits := subscription.Status == SubscriptionStatusActive || subscription.Status == SubscriptionStatusPastDue || subscription.Status == SubscriptionStatusTrialing

		return &UserCodySubscription{
			Status:               subscription.Status,
			IsPro:                true,
			ApplyProRateLimits:   applyProRateLimits,
			CurrentPeriodStartAt: currentPeriodStart,
			CurrentPeriodEndAt:   currentPeriodEnd,
		}, nil
	}

	currentPeriodStartAt, currentPeriodEndAt, err := preSSCCurrentPeriodDateRange(ctx, user)
	if err != nil {
		return nil, err
	}

	// The user opted for Cody Pro Trial on dotcom but have not put in their cc on SSC
	if user.CodyProOptedAt != nil {
		if currentPeriodEndAt.After(febReleaseDate) {
			currentPeriodEndAt = febReleaseDate
		}

		return &UserCodySubscription{
			Status:               SubscriptionStatusPending,
			IsPro:                true,
			ApplyProRateLimits:   isBeforeFebReleaseDate,
			CurrentPeriodStartAt: currentPeriodStartAt,
			CurrentPeriodEndAt:   currentPeriodEndAt,
		}, nil
	}

	// The user neither opted for Cody Pro on dotcom nor have put in their cc on SSC
	return &UserCodySubscription{
		Status:               SubscriptionStatusPending,
		IsPro:                false,
		ApplyProRateLimits:   false,
		CurrentPeriodStartAt: currentPeriodStartAt,
		CurrentPeriodEndAt:   currentPeriodEndAt,
	}, nil
}

func CodySubscriptionForUser(ctx context.Context, user types.User) (*UserCodySubscription, error) {
	// TODO (naman): get account_id from user_external_services table
	// based on a feature flag swtich between accounts.sourcegraph.com and accounts.sgdev.org
	sams_account_id := "018d18a8-55e8-7d2f-b902-4d29c308565f"

	var subscription *SSCSubscription
	var err error

	if featureflag.FromContext(ctx).GetBoolOr("ssc-enabled", false) {
		subscription, err = FetchSubscriptionBySAMSAccountID(sams_account_id)
		if err != nil {
			return nil, errors.Wrap(err, "error while fetching user subscription from SSC")
		}
	}

	return ConsolidateSubscriptionDetails(ctx, user, subscription)
}
