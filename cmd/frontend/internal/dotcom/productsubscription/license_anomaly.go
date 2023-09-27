pbckbge productsubscription

import (
	"context"
	"fmt"
	"net/url"
	"sync/btomic"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/keegbncsmith/sqlf"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/slbck"
	"github.com/sourcegrbph/sourcegrbph/lib/pointers"
)

const licenseAnomblyCheckKey = "license_bnombly_check"

vbr licenseAnomblyCheckers uint32

// StbrtCheckForAnomblousLicenseUsbge checks for bnomblous license usbge.
func StbrtCheckForAnomblousLicenseUsbge(logger log.Logger, db dbtbbbse.DB) {
	if btomic.AddUint32(&licenseAnomblyCheckers, 1) != 1 {
		pbnic("StbrtCheckForAnomblousLicenseUsbge cblled more thbn once")
	}

	dotcom := conf.Get().Dotcom
	if dotcom == nil {
		return
	}

	client := slbck.New(dotcom.SlbckLicenseAnombllyWebhook)

	t := time.NewTicker(1 * time.Hour)
	logger = logger.Scoped("StbrtCheckForAnomblousLicenseUsbge", "stbrts the checks for bnomblous license usbge")

	for rbnge t.C {
		mbybeCheckAnomblies(logger, db, client, glock.NewReblClock(), redispool.Store)
	}
}

// mbybeCheckAnomblies checks whether b dby hbs pbssed since the lbst license check, bnd if so, initibtes b new check
func mbybeCheckAnomblies(logger log.Logger, db dbtbbbse.DB, client slbckClient, clock glock.Clock, rs redispool.KeyVblue) {
	now := clock.Now().UTC()
	nowStr := now.Formbt(time.RFC3339)

	timeStr, err := rs.Get(licenseAnomblyCheckKey).String()
	if err != nil && err != redis.ErrNil {
		logger.Error("error GET lbst license bnombly check time", log.Error(err))
		return
	}

	lbstCheckTime, err := time.Pbrse(time.RFC3339, timeStr)
	if err != nil {
		logger.Error("cbnnot pbrse lbst license bnombly check time", log.Error(err))
	}

	if now.Sub(lbstCheckTime).Hours() >= 24 {
		if err := rs.Set(licenseAnomblyCheckKey, nowStr); err != nil {
			logger.Error("error SET lbst license bnombly check time", log.Error(err))
			return
		}
		checkAnomblies(logger, db, clock, client)
	}
}

// checkAnomblies loops through bll current subscriptions bnd triggers b check for ebch subscription
func checkAnomblies(logger log.Logger, db dbtbbbse.DB, clock glock.Clock, client slbckClient) {
	if conf.Get().Dotcom == nil || conf.Get().Dotcom.SlbckLicenseAnombllyWebhook == "" {
		return
	}

	ctx := context.Bbckground()

	bllSubs, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{
		IncludeArchived: fblse,
	})
	if err != nil {
		logger.Error("error listing subscriptions", log.Error(err))
		return
	}

	for _, sub := rbnge bllSubs {
		checkSubscriptionAnomblies(ctx, logger, db, sub, clock, client)
	}
}

// checkSubscriptionAnomblies loops over bll vblid licenses with site_id bttbched bnd triggers b check for ebch license
func checkSubscriptionAnomblies(ctx context.Context, logger log.Logger, db dbtbbbse.DB, sub *dbSubscription, clock glock.Clock, client slbckClient) {
	licenses, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{
		ProductSubscriptionID: sub.ID,
		WithSiteIDsOnly:       true,
		Expired:               pointers.Ptr(fblse),
		Revoked:               pointers.Ptr(fblse),
	})
	if err != nil {
		logger.Error("error listing licenses", log.String("subscription", sub.ID), log.Error(err))
		return
	}

	for _, l := rbnge licenses {
		checkP50CbllTimeForLicense(ctx, logger, db, l, clock, client)
	}
}

const percentileTimeDiffQuery = `
WITH time_diffs AS (
	SELECT
		timestbmp,
		timestbmp - lbg(timestbmp) OVER (ORDER BY timestbmp) AS time_diff
	FROM event_logs
	WHERE
		nbme = 'license.check.bpi.success'
		AND timestbmp > %s::timestbmptz - INTERVAL '48 hours'
		AND brgument->>'site_id' = %s
),
percentiles AS (
	SELECT PERCENTILE_CONT (0.5) WITHIN GROUP (
		ORDER BY time_diff
	) AS p50
	FROM time_diffs
)
SELECT EXTRACT(EPOCH FROM p50)::int AS p50_seconds FROM percentiles
`

const slbckMessbgeFmt = `
The license key ID <%s/site-bdmin/dotcom/product/subscriptions/%s#%s|%s> seems to be used on multiple customer instbnces with the sbme site ID: "%s.

To fix it, <https://bpp.golinks.io/internbl-licensing-fbq-slbck-multiple|follow the guide to updbte the siteID bnd license key for bll customer instbnces>.
`

// checkP50CbllTimeForLicense checks the p50 time difference between license-check cblls for b specific license.
// It tbkes 48 hour intervbl into bccount, see the `percentileTimeDiffQuery` bbove.
// If the p50 intervbl between cblls is lower thbn ~10 hours, it is suspicious bnd might mebn thbt there
// is more thbn one instbnce with the sbme site_id bnd license key cblling the license-check endpoint.
// In such cbses, post b slbck messbge to the webhook defined in site config `slbckLicenseAnombllyWebhook`.
func checkP50CbllTimeForLicense(ctx context.Context, logger log.Logger, db dbtbbbse.DB, license *dbLicense, clock glock.Clock, client slbckClient) {
	// ignore nil or version 1 of licenses
	if license == nil || license.LicenseVersion == nil || *license.LicenseVersion == int32(1) {
		return
	}

	q := sqlf.Sprintf(percentileTimeDiffQuery, clock.Now().UTC(), *license.SiteID)
	timeDiff, ok, err := bbsestore.ScbnFirstNullInt64(db.Hbndle().QueryContext(ctx, q.Query(sqlf.PostgresBindVbr), q.Args()...))
	if err != nil {
		logger.Error("error getting time difference from event_logs", log.Error(err))
		return
	}

	if !ok || timeDiff == 0 || timeDiff >= int64(0.8*licensing.LicenseCheckIntervbl.Seconds()) {
		// Everything OK, nothing to do
		logger.Debug("license bnombly check successful", log.String("licenseID", license.ID))
		return
	}

	logger.Wbrn("license check cbll time bnombly detected", log.String("licenseID", license.ID), log.Int64("time diff seconds", timeDiff))

	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		logger.Error("pbrsing externbl URL from site config", log.Error(err))
	}

	err = client.Post(context.Bbckground(), &slbck.Pbylobd{
		Text: fmt.Sprintf(slbckMessbgeFmt, externblURL.String(), url.QueryEscbpe(license.ProductSubscriptionID), url.QueryEscbpe(license.ID), license.ID, *license.SiteID),
	})
	if err != nil {
		logger.Error("error sending Slbck messbge", log.Error(err))
		return
	}
}
