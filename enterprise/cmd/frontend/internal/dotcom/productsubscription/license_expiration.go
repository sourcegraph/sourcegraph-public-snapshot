package productsubscription

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/slack"
)

var licenseExpirationCheckers uint32

// StartCheckForUpcomingLicenseExpirations checks for upcoming license expirations once per day.
func StartCheckForUpcomingLicenseExpirations() {
	if atomic.AddUint32(&licenseExpirationCheckers, 1) != 1 {
		panic("StartCheckForUpcomingLicenseExpirations called more than once")
	}

	client := slack.New(conf.Get().Dotcom.SlackLicenseExpirationWebhook)

	ctx := context.Background()

	t := time.NewTicker(24 * time.Hour)

	for range t.C {
		allDBSubscriptions, err := dbSubscriptions{}.List(ctx, dbSubscriptionsListOptions{
			IncludeArchived: false,
		})
		if err != nil {
			log15.Error("startCheckForUpcomingLicenseExpirations: error listing subscriptions", "error", err)
			continue
		}

		for _, dbSubscription := range allDBSubscriptions {
			// Get the active (i.e., latest created) license.
			licenses, err := dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: dbSubscription.ID, LimitOffset: &db.LimitOffset{Limit: 1}})
			if err != nil {
				log15.Error("startCheckForUpcomingLicenseExpirations: error listing licenses", "error", err)
				continue
			}
			// Skip if subscription has no licenses.
			if len(licenses) < 1 {
				continue
			}

			user, err := db.Users.GetByID(ctx, dbSubscription.UserID)
			if err != nil {
				log15.Error("startCheckForUpcomingLicenseExpirations: error looking up user", "error", err)
				continue
			}

			info, _, err := licensing.ParseProductLicenseKeyWithBuiltinOrGenerationKey(licenses[0].LicenseKey)
			if err != nil {
				log15.Error("startCheckForUpcomingLicenseExpirations: error parsing license key", "error", err)
				continue
			}

			weekAway := time.Now().Add(7 * 24 * time.Hour)
			dayAway := time.Now().Add(24 * time.Hour)

			if info.ExpiresAt.After(weekAway) && info.ExpiresAt.Before(weekAway.Add(24*time.Hour)) {
				err = client.Post(&slack.Payload{
					Text: fmt.Sprintf("The license for user `%s` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s|will expire *in 7 days*>", user.Username, dbSubscription.ID),
				})
				if err != nil {
					log15.Error("startCheckForUpcomingLicenseExpirations: error sending Slack message", "error", err)
					continue
				}
			} else if info.ExpiresAt.After(dayAway) && info.ExpiresAt.Before(dayAway.Add(24*time.Hour)) {
				err = client.Post(&slack.Payload{
					Text: fmt.Sprintf("The license for user `%s` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s|will expire *tomorrow*> :rotating_light:", user.Username, dbSubscription.ID),
				})
				if err != nil {
					log15.Error("startCheckForUpcomingLicenseExpirations: error sending Slack message", "error", err)
					continue
				}

			}
		}
	}
}
