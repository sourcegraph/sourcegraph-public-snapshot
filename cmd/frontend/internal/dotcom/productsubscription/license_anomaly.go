package productsubscription

import (
	"context"
	"fmt"
	"net/url"
	"sync/atomic"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/slack"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const licenseAnomalyCheckKey = "license_anomaly_check"

var licenseAnomalyCheckers uint32

// StartCheckForAnomalousLicenseUsage checks for anomalous license usage.
func StartCheckForAnomalousLicenseUsage(logger log.Logger, db database.DB) {
	if atomic.AddUint32(&licenseAnomalyCheckers, 1) != 1 {
		panic("StartCheckForAnomalousLicenseUsage called more than once")
	}

	dotcom := conf.Get().Dotcom
	if dotcom == nil {
		return
	}

	client := slack.New(dotcom.SlackLicenseAnomallyWebhook)

	t := time.NewTicker(1 * time.Hour)
	logger = logger.Scoped("StartCheckForAnomalousLicenseUsage")

	for range t.C {
		maybeCheckAnomalies(logger, db, client, glock.NewRealClock(), redispool.Store)
	}
}

// maybeCheckAnomalies checks whether a day has passed since the last license check, and if so, initiates a new check
func maybeCheckAnomalies(logger log.Logger, db database.DB, client slackClient, clock glock.Clock, rs redispool.KeyValue) {
	now := clock.Now().UTC()
	nowStr := now.Format(time.RFC3339)

	timeStr, err := rs.Get(licenseAnomalyCheckKey).String()
	if err != nil && err != redis.ErrNil {
		logger.Error("error GET last license anomaly check time", log.Error(err))
		return
	}

	lastCheckTime, err := time.Parse(time.RFC3339, timeStr)
	if err != nil {
		logger.Error("cannot parse last license anomaly check time", log.Error(err))
	}

	if now.Sub(lastCheckTime).Hours() >= 24 {
		if err := rs.Set(licenseAnomalyCheckKey, nowStr); err != nil {
			logger.Error("error SET last license anomaly check time", log.Error(err))
			return
		}
		checkAnomalies(logger, db, clock, client)
	}
}

// checkAnomalies loops through all current subscriptions and triggers a check for each subscription
func checkAnomalies(logger log.Logger, db database.DB, clock glock.Clock, client slackClient) {
	if conf.Get().Dotcom == nil || conf.Get().Dotcom.SlackLicenseAnomallyWebhook == "" {
		return
	}

	ctx := context.Background()

	allSubs, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{
		IncludeArchived: false,
	})
	if err != nil {
		logger.Error("error listing subscriptions", log.Error(err))
		return
	}

	for _, sub := range allSubs {
		checkSubscriptionAnomalies(ctx, logger, db, sub, clock, client)
	}
}

// checkSubscriptionAnomalies loops over all valid licenses with site_id attached and triggers a check for each license
func checkSubscriptionAnomalies(ctx context.Context, logger log.Logger, db database.DB, sub *dbSubscription, clock glock.Clock, client slackClient) {
	licenses, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{
		ProductSubscriptionID: sub.ID,
		WithSiteIDsOnly:       true,
		Expired:               pointers.Ptr(false),
		Revoked:               pointers.Ptr(false),
	})
	if err != nil {
		logger.Error("error listing licenses", log.String("subscription", sub.ID), log.Error(err))
		return
	}

	for _, l := range licenses {
		checkP50CallTimeForLicense(ctx, logger, db, l, clock, client)
	}
}

const percentileTimeDiffQuery = `
WITH time_diffs AS (
	SELECT
		timestamp,
		timestamp - lag(timestamp) OVER (ORDER BY timestamp) AS time_diff
	FROM event_logs
	WHERE
		name = 'license.check.api.success'
		AND timestamp > %s::timestamptz - INTERVAL '48 hours'
		AND argument->>'site_id' = %s
),
percentiles AS (
	SELECT PERCENTILE_CONT (0.5) WITHIN GROUP (
		ORDER BY time_diff
	) AS p50
	FROM time_diffs
)
SELECT EXTRACT(EPOCH FROM p50)::int AS p50_seconds FROM percentiles
`

const slackMessageFmt = `
We are receiving irregular pings from the site ID: "%s".

This might mean that there is more than one active instance of Sourcegraph with the same site ID and license key.

At the moment, this site ID is associated with the following license key ID: <%s/site-admin/dotcom/product/subscriptions/%s#%s|%s>

To fix this, <https://app.golinks.io/internal-licensing-faq-slack-multiple|follow the guide to update the siteID and license key for all customer instances>.
`

func formatSlackMessage(externalURL *url.URL, license *dbLicense) string {
	return fmt.Sprintf(slackMessageFmt,
		*license.SiteID,
		externalURL.String(),
		url.QueryEscape(license.ProductSubscriptionID),
		url.QueryEscape(license.ID),
		license.ID)
}

// checkP50CallTimeForLicense checks the p50 time difference between license-check calls for a specific license.
// It takes 48 hour interval into account, see the `percentileTimeDiffQuery` above.
// If the p50 interval between calls is lower than ~10 hours, it is suspicious and might mean that there
// is more than one instance with the same site_id and license key calling the license-check endpoint.
// In such cases, post a slack message to the webhook defined in site config `slackLicenseAnomallyWebhook`.
func checkP50CallTimeForLicense(ctx context.Context, logger log.Logger, db database.DB, license *dbLicense, clock glock.Clock, client slackClient) {
	// ignore nil or version 1 of licenses
	if license == nil || license.LicenseVersion == nil || *license.LicenseVersion == int32(1) {
		return
	}

	q := sqlf.Sprintf(percentileTimeDiffQuery, clock.Now().UTC(), *license.SiteID)
	timeDiff, ok, err := basestore.ScanFirstNullInt64(db.Handle().QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...))
	if err != nil {
		logger.Error("error getting time difference from event_logs", log.Error(err))
		return
	}

	if !ok || timeDiff == 0 || timeDiff >= int64(0.8*licensing.LicenseCheckInterval.Seconds()) {
		// Everything OK, nothing to do
		logger.Debug("license anomaly check successful", log.String("licenseID", license.ID))
		return
	}

	logger.Warn("license check call time anomaly detected", log.String("licenseID", license.ID), log.Int64("time diff seconds", timeDiff))

	externalURL, err := url.Parse(conf.Get().ExternalURL)
	if err != nil {
		logger.Error("parsing external URL from site config", log.Error(err))
	}

	err = client.Post(context.Background(), &slack.Payload{
		Text: formatSlackMessage(externalURL, license),
	})
	if err != nil {
		logger.Error("error sending Slack message", log.Error(err))
		return
	}
}
