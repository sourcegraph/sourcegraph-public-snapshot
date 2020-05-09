package productsubscription

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/slack"
)

const key = "last_license_expiration_check"

var (
	licenseExpirationCheckers uint32

	pool = redispool.Store
)

// StartCheckForUpcomingLicenseExpirations checks for upcoming license expirations once per day.
func StartCheckForUpcomingLicenseExpirations() {
	if atomic.AddUint32(&licenseExpirationCheckers, 1) != 1 {
		panic("StartCheckForUpcomingLicenseExpirations called more than once")
	}

	t := time.NewTicker(1 * time.Hour)
	c := pool.Get()
	defer c.Close()

	for range t.C {
		now := time.Now()
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).Format("2006-01-02")

		last_check_day, err := redis.String(c.Do("GETSET", key, today))
		if err != nil {
			log15.Error("startCheckForUpcomingLicenseExpirations: error getting/setting last license expiration check day", "error", err)
			continue
		}

		if today != last_check_day {
			check()
		}
	}
}

func check() {
	if conf.Get().Dotcom == nil || conf.Get().Dotcom.SlackLicenseExpirationWebhook == "" {
		return
	}

	client := slack.New(conf.Get().Dotcom.SlackLicenseExpirationWebhook)

	ctx := context.Background()

	allDBSubscriptions, err := dbSubscriptions{}.List(ctx, dbSubscriptionsListOptions{
		IncludeArchived: false,
	})
	if err != nil {
		log15.Error("startCheckForUpcomingLicenseExpirations: error listing subscriptions", "error", err)
		return
	}

	for _, dbSubscription := range allDBSubscriptions {
		// Get the active (i.e., latest created) license.
		licenses, err := dbLicenses{}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: dbSubscription.ID, LimitOffset: &db.LimitOffset{Limit: 1}})
		if err != nil {
			log15.Error("startCheckForUpcomingLicenseExpirations: error listing licenses", "error", err)
			return
		}
		// Skip if subscription has no licenses.
		if len(licenses) < 1 {
			return
		}

		user, err := db.Users.GetByID(ctx, dbSubscription.UserID)
		if err != nil {
			log15.Error("startCheckForUpcomingLicenseExpirations: error looking up user", "error", err)
			return
		}

		info, _, err := licensing.ParseProductLicenseKeyWithBuiltinOrGenerationKey(licenses[0].LicenseKey)
		if err != nil {
			log15.Error("startCheckForUpcomingLicenseExpirations: error parsing license key", "error", err)
			return
		}

		weekAway := time.Now().Add(7 * 24 * time.Hour)
		dayAway := time.Now().Add(24 * time.Hour)

		if info.ExpiresAt.After(weekAway) && info.ExpiresAt.Before(weekAway.Add(24*time.Hour)) {
			err = client.Post(&slack.Payload{
				Text: fmt.Sprintf("The license for user `%s` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s|will expire *in 7 days*>", user.Username, dbSubscription.ID),
			})
			if err != nil {
				log15.Error("startCheckForUpcomingLicenseExpirations: error sending Slack message", "error", err)
				return
			}
		} else if info.ExpiresAt.After(dayAway) && info.ExpiresAt.Before(dayAway.Add(24*time.Hour)) {
			err = client.Post(&slack.Payload{
				Text: fmt.Sprintf("The license for user `%s` <https://sourcegraph.com/site-admin/dotcom/product/subscriptions/%s|will expire *in the next 24 hours*> :rotating_light:", user.Username, dbSubscription.ID),
			})
			if err != nil {
				log15.Error("startCheckForUpcomingLicenseExpirations: error sending Slack message", "error", err)
				return
			}
		}
	}
}
