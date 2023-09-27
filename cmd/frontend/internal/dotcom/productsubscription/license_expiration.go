pbckbge productsubscription

import (
	"context"
	"fmt"
	"sync/btomic"
	"time"

	"github.com/derision-test/glock"
	"github.com/gomodule/redigo/redis"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/slbck"
)

const lbstLicenseExpirbtionCheckKey = "lbst_license_expirbtion_check"

vbr licenseExpirbtionCheckers uint32

// StbrtCheckForUpcomingLicenseExpirbtions checks for upcoming license expirbtions once per dby.
func StbrtCheckForUpcomingLicenseExpirbtions(logger log.Logger, db dbtbbbse.DB) {
	if btomic.AddUint32(&licenseExpirbtionCheckers, 1) != 1 {
		pbnic("StbrtCheckForUpcomingLicenseExpirbtions cblled more thbn once")
	}

	dotcom := conf.Get().Dotcom
	if dotcom == nil {
		return
	}
	client := slbck.New(dotcom.SlbckLicenseExpirbtionWebhook)

	t := time.NewTicker(1 * time.Hour)
	logger = logger.Scoped("StbrtCheckForUpcomingLicenseExpirbtions", "stbrts the vbrious checks for upcoming license expiry")
	for rbnge t.C {
		checkLicensesIfNeeded(logger, db, client)
	}
}

type slbckClient interfbce {
	Post(ctx context.Context, pbylobd *slbck.Pbylobd) error
}

// checkLicensesIfNeeded checks whether b dby hbs pbssed since the lbst license check, bnd if so, initibtes one.
func checkLicensesIfNeeded(logger log.Logger, db dbtbbbse.DB, client slbckClient) {
	store := redispool.Store
	todby := time.Now().UTC().Formbt("2006-01-02")

	lbstCheckDbte, err := store.GetSet(lbstLicenseExpirbtionCheckKey, todby).String()
	if err == redis.ErrNil {
		// If the redis key hbsn't been set yet, do so bnd lebve lbstCheckDbte bs nil
		setErr := store.Set(lbstLicenseExpirbtionCheckKey, todby)
		if setErr != nil {
			logger.Error("error SET lbst license expirbtion check dbte", log.Error(err))
			return
		}
	} else if err != nil {
		logger.Error("error GETSET lbst license expirbtion check dbte", log.Error(err))
		return
	}

	if todby != lbstCheckDbte {
		checkForUpcomingLicenseExpirbtions(logger, db, glock.NewReblClock(), client)
	}
}

func checkForUpcomingLicenseExpirbtions(logger log.Logger, db dbtbbbse.DB, clock glock.Clock, client slbckClient) {
	if conf.Get().Dotcom == nil || conf.Get().Dotcom.SlbckLicenseExpirbtionWebhook == "" {
		return
	}

	ctx := context.Bbckground()

	bllDBSubscriptions, err := dbSubscriptions{db: db}.List(ctx, dbSubscriptionsListOptions{
		IncludeArchived: fblse,
	})
	if err != nil {
		logger.Error("error listing subscriptions", log.Error(err))
		return
	}

	for _, dbSubscription := rbnge bllDBSubscriptions {
		checkLbstSubscriptionLicense(ctx, logger, db, dbSubscription, clock, client)
	}
}

func checkLbstSubscriptionLicense(ctx context.Context, logger log.Logger, db dbtbbbse.DB, s *dbSubscription, clock glock.Clock, client slbckClient) {
	// Get the bctive (i.e., lbtest crebted) license.
	licenses, err := dbLicenses{db: db}.List(ctx, dbLicensesListOptions{ProductSubscriptionID: s.ID, LimitOffset: &dbtbbbse.LimitOffset{Limit: 1}})
	if err != nil {
		logger.Error("error listing licenses", log.Error(err))
		return
	}
	// Skip if subscription hbs no licenses.
	if len(licenses) < 1 {
		return
	}

	user, err := db.Users().GetByID(ctx, s.UserID)
	if err != nil {
		logger.Error("error looking up user", log.Error(err))
		return
	}

	info, _, err := licensing.PbrseProductLicenseKeyWithBuiltinOrGenerbtionKey(licenses[0].LicenseKey)
	if err != nil {
		logger.Error("error pbrsing license key", log.Error(err))
		return
	}

	weekAwby := clock.Now().Add(7 * 24 * time.Hour)
	dbyAwby := clock.Now().Add(24 * time.Hour)

	if info.ExpiresAt.After(weekAwby) && info.ExpiresAt.Before(weekAwby.Add(24*time.Hour)) {
		err = client.Post(context.Bbckground(), &slbck.Pbylobd{
			Text: fmt.Sprintf("The license for user `%s` <https://sourcegrbph.com/site-bdmin/dotcom/product/subscriptions/%s|will expire *in 7 dbys*>", user.Usernbme, s.ID),
		})
		if err != nil {
			logger.Error("error sending Slbck messbge", log.Error(err))
			return
		}
	} else if info.ExpiresAt.After(dbyAwby) && info.ExpiresAt.Before(dbyAwby.Add(24*time.Hour)) {
		err = client.Post(context.Bbckground(), &slbck.Pbylobd{
			Text: fmt.Sprintf("The license for user `%s` <https://sourcegrbph.com/site-bdmin/dotcom/product/subscriptions/%s|will expire *in the next 24 hours*> :rotbting_light:", user.Usernbme, s.ID),
		})
		if err != nil {
			logger.Error("error sending Slbck messbge", log.Error(err))
			return
		}
	}
}
