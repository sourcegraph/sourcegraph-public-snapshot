package productsubscription

import (
	"context"
	"fmt"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/slack"
)

var SourcegraphOrgWebhookURL = env.Get("SLACK_LICENSE_EXPIRATION_BOT_HOOK", "", "Webhook for announcing license expirations to a Slack channel for Sourcegraph.com.")

var started bool

// StartCheckForUpcomingLicenseExpirations checks for upcoming license expirations once per day.
func StartCheckForUpcomingLicenseExpirations() {
	if started {
		panic("already started")
	}
	started = true

	client := slack.New(SourcegraphOrgWebhookURL)

	ctx := context.Background()
	const delay = 24 * time.Hour
	for {
		allDBSubscriptions, err := dbSubscriptions{}.List(ctx, dbSubscriptionsListOptions{
			IncludeArchived: false,
		})
		if err != nil {
			log15.Warn("startCheckForUpcomingLicenseExpirations: error listing subscriptions", "error", err)
		}

		for _, dbSubscription := range allDBSubscriptions {
			// Get the active (i.e., latest created) license.
			licenses, err := dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: dbSubscription.ID, LimitOffset: &db.LimitOffset{Limit: 1}})
			if err != nil {
				log15.Warn("startCheckForUpcomingLicenseExpirations: error listing licenses", "error", err)
			}
			// Skip if subscription has no licenses.
			if len(licenses) < 1 {
				continue
			}

			user, err := db.Users.GetByID(ctx, dbSubscription.UserID)
			if err != nil {
				log15.Warn("startCheckForUpcomingLicenseExpirations: error looking up user", "error", err)
			}

			info, _, err := licensing.ParseProductLicenseKeyWithBuiltinOrGenerationKey(licenses[0].LicenseKey)
			if err != nil {
				log15.Warn("startCheckForUpcomingLicenseExpirations: error parsing license key", "error", err)
			}

			weekAway := time.Now().Add(7 * 24 * time.Hour)
			dayAway := time.Now().Add(24 * time.Hour)

			if info.ExpiresAt.After(weekAway) && info.ExpiresAt.Before(weekAway.Add(24*time.Hour)) {
				err = client.Post(&slack.Payload{
					Text: fmt.Sprintf("The license for user `%s` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s|will expire *in 7 days*>", user.Username, dbSubscription.ID),
				})
				if err != nil {
					log15.Warn("startCheckForUpcomingLicenseExpirations: error sending Slack message", "error", err)
				}
			}
			if info.ExpiresAt.After(dayAway) && info.ExpiresAt.Before(dayAway.Add(24*time.Hour)) {
				err = client.Post(&slack.Payload{
					Text: fmt.Sprintf("The license for user `%s` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s|will expire *tomorrow*> :rotating_light:", user.Username, dbSubscription.ID),
				})
				if err != nil {
					log15.Warn("startCheckForUpcomingLicenseExpirations: error sending Slack message", "error", err)
				}

			}
		}
		time.Sleep(delay)
	}
}
