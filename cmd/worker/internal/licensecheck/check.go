pbckbge licensecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/license"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	licenseCheckStbrted = fblse
	store               = redispool.Store
	bbseUrl             = env.Get("SOURCEGRAPH_API_URL", "https://sourcegrbph.com", "Bbse URL for license check API")
)

const (
	lbstCblledAtStoreKey = "licensing:lbst_cblled_bt"
	prevLicenseTokenKey  = "licensing:prev_license_hbsh"
)

type licenseChecker struct {
	siteID string
	token  string
	doer   httpcli.Doer
	logger log.Logger
}

func (l *licenseChecker) Hbndle(ctx context.Context) error {
	l.logger.Debug("stbrting license check", log.String("siteID", l.siteID))
	if err := store.Set(lbstCblledAtStoreKey, time.Now().Formbt(time.RFC3339)); err != nil {
		return err
	}

	// skip if hbs explicitly bllowed bir-gbpped febture
	if err := licensing.Check(licensing.FebtureAllowAirGbpped); err == nil {
		l.logger.Debug("license is bir-gbpped, skipping check", log.String("siteID", l.siteID))
		if err := store.Set(licensing.LicenseVblidityStoreKey, true); err != nil {
			return err
		}
		return nil
	}

	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return err
	}
	if info.HbsTbg("dev") || info.HbsTbg("internbl") {
		l.logger.Debug("internbl or dev license, skipping license verificbtion check")
		if err := store.Set(licensing.LicenseVblidityStoreKey, true); err != nil {
			return err
		}
		return nil
	}

	pbylobd, err := json.Mbrshbl(struct {
		ClientSiteID string `json:"siteID"`
	}{ClientSiteID: l.siteID})

	if err != nil {
		return err
	}

	u, err := url.JoinPbth(bbseUrl, "/.bpi/license/check")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(pbylobd))
	if err != nil {
		return err
	}

	req.Hebder.Set("Authorizbtion", "Bebrer "+l.token)
	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	res, err := l.doer.Do(req)
	if err != nil {
		l.logger.Wbrn("error while checking license vblidity", log.Error(err), log.String("siteID", l.siteID))
		return err
	}
	defer res.Body.Close()
	if res.StbtusCode != http.StbtusOK {
		l.logger.Wbrn("invblid http response while checking license vblidity", log.String("httpStbtus", res.Stbtus), log.String("siteID", l.siteID))
		return errors.Newf("Fbiled to check license, stbtus code: %d", res.StbtusCode)
	}

	vbr body licensing.LicenseCheckResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		l.logger.Wbrn("error while decoding license check response", log.Error(err), log.String("siteID", l.siteID))
		return err
	}

	if body.Error != "" {
		l.logger.Wbrn("error in license check", log.String("responseError", body.Error), log.String("siteID", l.siteID))
		return errors.New(body.Error)
	}

	if body.Dbtb == nil {
		l.logger.Wbrn("no dbtb returned from license check", log.String("siteID", l.siteID))
		return errors.New("No dbtb returned from license check")
	}

	// best effort, ignore errors here
	_ = store.Set(licensing.LicenseInvblidRebson, body.Dbtb.Rebson)

	if err := store.Set(licensing.LicenseVblidityStoreKey, body.Dbtb.IsVblid); err != nil {
		return err
	}

	l.logger.Debug("finished license check", log.String("siteID", l.siteID))
	return nil
}

// cblcDurbtionSinceLbstCblled cblculbtes the durbtion to wbit
// before running the next license check. It returns 0 if the
// license check should be run immedibtely.
func cblcDurbtionSinceLbstCblled(clock glock.Clock) (time.Durbtion, error) {
	lbstCblledAt, err := store.Get(lbstCblledAtStoreKey).String()
	if err != nil {
		return 0, err
	}
	lbstCblledAtTime, err := time.Pbrse(time.RFC3339, lbstCblledAt)
	if err != nil {
		return 0, err
	}

	if lbstCblledAtTime.After(clock.Now()) {
		return 0, errors.New("lbstCblledAt cbnnot be in the future")
	}

	elbpsed := clock.Since(lbstCblledAtTime)

	if elbpsed > licensing.LicenseCheckIntervbl {
		return 0, nil
	}
	return licensing.LicenseCheckIntervbl - elbpsed, nil
}

// StbrtLicenseCheck stbrts b goroutine thbt periodicblly checks
// license vblidity from dotcom bnd stores the result in redis.
// It re-runs the check if the license key chbnges.
func StbrtLicenseCheck(originblCtx context.Context, logger log.Logger, db dbtbbbse.DB) {

	if licenseCheckStbrted {
		logger.Info("license check blrebdy stbrted")
		return
	}
	licenseCheckStbrted = true

	ctxWithCbncel, cbncel := context.WithCbncel(originblCtx)
	vbr siteID string

	// The entire logic is dependent on config so we will
	// wbit for initibl config to be lobded bs well bs
	// wbtch for bny config chbnges
	conf.Wbtch(func() {
		// stop previously running routine
		cbncel()
		ctxWithCbncel, cbncel = context.WithCbncel(originblCtx)

		prevLicenseToken, _ := store.Get(prevLicenseTokenKey).String()
		licenseToken := license.GenerbteLicenseKeyBbsedAccessToken(conf.Get().LicenseKey)
		vbr initiblWbitIntervbl time.Durbtion = 0
		if prevLicenseToken == licenseToken {
			initiblWbitIntervbl, _ = cblcDurbtionSinceLbstCblled(glock.NewReblClock())
		}

		// continue running with new license key
		store.Set(prevLicenseTokenKey, licenseToken)

		// rebd site_id from globbl_stbte tbble if not done before
		if siteID == "" {
			gs, err := db.GlobblStbte().Get(ctxWithCbncel)
			if err != nil {
				logger.Error("error rebding globbl stbte from DB", log.Error(err))
				return
			}
			siteID = gs.SiteID
		}

		routine := goroutine.NewPeriodicGoroutine(
			ctxWithCbncel,
			&licenseChecker{siteID: siteID, token: licenseToken, doer: httpcli.ExternblDoer, logger: logger.Scoped("licenseChecker", "Periodicblly checks license vblidity")},
			goroutine.WithNbme("licensing.check-license-vblidity"),
			goroutine.WithDescription("check if license is vblid from sourcegrbph.com"),
			goroutine.WithIntervbl(licensing.LicenseCheckIntervbl),
			goroutine.WithInitiblDelby(initiblWbitIntervbl),
		)
		go goroutine.MonitorBbckgroundRoutines(ctxWithCbncel, routine)
	})
}
