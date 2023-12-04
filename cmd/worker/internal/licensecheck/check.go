package licensecheck

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	licenseCheckStarted = false
	store               = redispool.Store
	baseUrl             = env.Get("SOURCEGRAPH_API_URL", "https://sourcegraph.com", "Base URL for license check API")
)

const (
	lastCalledAtStoreKey = "licensing:last_called_at"
	prevLicenseTokenKey  = "licensing:prev_license_hash"
)

type licenseChecker struct {
	siteID string
	token  string
	doer   httpcli.Doer
	logger log.Logger
}

func (l *licenseChecker) Handle(ctx context.Context) error {
	l.logger.Debug("starting license check", log.String("siteID", l.siteID))
	if err := store.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	// skip if has explicitly allowed air-gapped feature
	if err := licensing.Check(licensing.FeatureAllowAirGapped); err == nil {
		l.logger.Debug("license is air-gapped, skipping check", log.String("siteID", l.siteID))
		if err := store.Set(licensing.LicenseValidityStoreKey, true); err != nil {
			return err
		}
		return nil
	}

	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return err
	}
	if info.HasTag("dev") || info.HasTag("internal") {
		l.logger.Debug("internal or dev license, skipping license verification check")
		if err := store.Set(licensing.LicenseValidityStoreKey, true); err != nil {
			return err
		}
		return nil
	}

	payload, err := json.Marshal(struct {
		ClientSiteID string `json:"siteID"`
	}{ClientSiteID: l.siteID})

	if err != nil {
		return err
	}

	u, err := url.JoinPath(baseUrl, "/.api/license/check")
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, u, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+l.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := l.doer.Do(req)
	if err != nil {
		l.logger.Warn("error while checking license validity", log.Error(err), log.String("siteID", l.siteID))
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		l.logger.Warn("invalid http response while checking license validity", log.String("httpStatus", res.Status), log.String("siteID", l.siteID))
		return errors.Newf("Failed to check license, status code: %d", res.StatusCode)
	}

	var body licensing.LicenseCheckResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		l.logger.Warn("error while decoding license check response", log.Error(err), log.String("siteID", l.siteID))
		return err
	}

	if body.Error != "" {
		l.logger.Warn("error in license check", log.String("responseError", body.Error), log.String("siteID", l.siteID))
		return errors.New(body.Error)
	}

	if body.Data == nil {
		l.logger.Warn("no data returned from license check", log.String("siteID", l.siteID))
		return errors.New("No data returned from license check")
	}

	// best effort, ignore errors here
	_ = store.Set(licensing.LicenseInvalidReason, body.Data.Reason)

	if err := store.Set(licensing.LicenseValidityStoreKey, body.Data.IsValid); err != nil {
		return err
	}

	l.logger.Debug("finished license check", log.String("siteID", l.siteID))
	return nil
}

// calcDurationSinceLastCalled calculates the duration to wait
// before running the next license check. It returns 0 if the
// license check should be run immediately.
func calcDurationSinceLastCalled(clock glock.Clock) (time.Duration, error) {
	lastCalledAt, err := store.Get(lastCalledAtStoreKey).String()
	if err != nil {
		return 0, err
	}
	lastCalledAtTime, err := time.Parse(time.RFC3339, lastCalledAt)
	if err != nil {
		return 0, err
	}

	if lastCalledAtTime.After(clock.Now()) {
		return 0, errors.New("lastCalledAt cannot be in the future")
	}

	elapsed := clock.Since(lastCalledAtTime)

	if elapsed > licensing.LicenseCheckInterval {
		return 0, nil
	}
	return licensing.LicenseCheckInterval - elapsed, nil
}

// StartLicenseCheck starts a goroutine that periodically checks
// license validity from dotcom and stores the result in redis.
// It re-runs the check if the license key changes.
func StartLicenseCheck(originalCtx context.Context, logger log.Logger, db database.DB) {

	if licenseCheckStarted {
		logger.Info("license check already started")
		return
	}
	licenseCheckStarted = true

	ctxWithCancel, cancel := context.WithCancel(originalCtx)
	var siteID string

	// The entire logic is dependent on config so we will
	// wait for initial config to be loaded as well as
	// watch for any config changes
	conf.Watch(func() {
		// stop previously running routine
		cancel()
		ctxWithCancel, cancel = context.WithCancel(originalCtx)

		prevLicenseToken, _ := store.Get(prevLicenseTokenKey).String()
		licenseToken := license.GenerateLicenseKeyBasedAccessToken(conf.Get().LicenseKey)
		var initialWaitInterval time.Duration = 0
		if prevLicenseToken == licenseToken {
			initialWaitInterval, _ = calcDurationSinceLastCalled(glock.NewRealClock())
		}

		// continue running with new license key
		store.Set(prevLicenseTokenKey, licenseToken)

		// read site_id from global_state table if not done before
		if siteID == "" {
			gs, err := db.GlobalState().Get(ctxWithCancel)
			if err != nil {
				logger.Error("error reading global state from DB", log.Error(err))
				return
			}
			siteID = gs.SiteID
		}

		routine := goroutine.NewPeriodicGoroutine(
			ctxWithCancel,
			&licenseChecker{siteID: siteID, token: licenseToken, doer: httpcli.ExternalDoer, logger: logger.Scoped("licenseChecker")},
			goroutine.WithName("licensing.check-license-validity"),
			goroutine.WithDescription("check if license is valid from sourcegraph.com"),
			goroutine.WithInterval(licensing.LicenseCheckInterval),
			goroutine.WithInitialDelay(initialWaitInterval),
		)
		go goroutine.MonitorBackgroundRoutines(ctxWithCancel, routine)
	})
}
