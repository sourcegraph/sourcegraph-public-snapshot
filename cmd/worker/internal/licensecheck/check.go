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
	baseUrl = env.Get("SOURCEGRAPH_API_URL", "https://sourcegraph.com", "Base URL for license check API")
)

// newLicenseChecker returns a goroutine that periodically checks license validity
// from dotcom and stores the result in redis.
// It re-runs the check if the license key changes.
func newLicenseChecker(ctx context.Context, logger log.Logger, db database.DB, kv redispool.KeyValue) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&licenseChecker{
			db:     db,
			doer:   httpcli.ExternalDoer,
			logger: logger.Scoped("licenseChecker"),
			kv:     kv,
		},
		goroutine.WithName("licensing.check-license-validity"),
		goroutine.WithDescription("check if license is valid from sourcegraph.com"),
		goroutine.WithInterval(15*time.Minute),
		goroutine.WithInitialDelay(1*time.Minute),
	)
}

const (
	lastCalledAtStoreKey = "licensing:last_called_at"
	prevLicenseTokenKey  = "licensing:prev_license_hash"
)

type licenseChecker struct {
	db     database.DB
	doer   httpcli.Doer
	logger log.Logger
	kv     redispool.KeyValue
}

func (l *licenseChecker) Handle(ctx context.Context) error {
	l.logger.Debug("starting license check")

	gs, err := l.db.GlobalState().Get(ctx)
	if err != nil {
		return errors.Wrap(err, "error reading global state from DB")
	}
	siteID := gs.SiteID

	// skip if has explicitly allowed air-gapped feature
	if err := licensing.Check(licensing.FeatureAllowAirGapped); err == nil {
		l.logger.Debug("license is air-gapped, skipping check", log.String("siteID", siteID))
		if err := l.kv.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339)); err != nil {
			return err
		}
		if err := l.kv.Set(licensing.LicenseValidityStoreKey, true); err != nil {
			return err
		}
		return nil
	}

	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return err
	}
	if info.HasTag("dev") || info.HasTag("internal") || info.Plan().IsFreePlan() {
		l.logger.Debug("internal, dev, or free license, skipping license verification check")
		if err := l.kv.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339)); err != nil {
			return err
		}
		if err := l.kv.Set(licensing.LicenseValidityStoreKey, true); err != nil {
			return err
		}
		return nil
	}

	// Check if the license key has changed and generate an auth token for the request.
	prevLicenseToken, _ := l.kv.Get(prevLicenseTokenKey).String()
	licenseToken := license.GenerateLicenseKeyBasedAccessToken(conf.Get().LicenseKey)
	// If the key hasn't changed, let's make sure we only hit this endpoint about
	// every 12 hours.
	if prevLicenseToken == licenseToken {
		if waitDuration, _ := calcDurationSinceLastCalled(l.kv, glock.NewRealClock()); waitDuration > 0 {
			l.logger.Debug("license key check is not due, skipping check", log.String("siteID", siteID))
			return nil
		}
	}

	// Continue running with new license key.
	if err := l.kv.Set(prevLicenseTokenKey, licenseToken); err != nil {
		l.logger.Error("error storing license token in redis", log.Error(err))
	}

	if err := l.kv.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	payload, err := json.Marshal(struct {
		ClientSiteID string `json:"siteID"`
	}{ClientSiteID: siteID})
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

	req.Header.Set("Authorization", "Bearer "+licenseToken)
	req.Header.Set("Content-Type", "application/json")

	res, err := l.doer.Do(req)
	if err != nil {
		l.logger.Warn("error while checking license validity", log.Error(err), log.String("siteID", siteID))
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		l.logger.Warn("invalid http response while checking license validity", log.String("httpStatus", res.Status), log.String("siteID", siteID))
		return errors.Newf("Failed to check license, status code: %d", res.StatusCode)
	}

	var body licensing.LicenseCheckResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		l.logger.Warn("error while decoding license check response", log.Error(err), log.String("siteID", siteID))
		return err
	}

	if body.Error != "" {
		l.logger.Warn("error in license check", log.String("responseError", body.Error), log.String("siteID", siteID))
		return errors.New(body.Error)
	}

	if body.Data == nil {
		l.logger.Warn("no data returned from license check", log.String("siteID", siteID))
		return errors.New("No data returned from license check")
	}

	// best effort, ignore errors here
	_ = l.kv.Set(licensing.LicenseInvalidReason, body.Data.Reason)

	if err := l.kv.Set(licensing.LicenseValidityStoreKey, body.Data.IsValid); err != nil {
		return err
	}

	l.logger.Debug("finished license check", log.String("siteID", siteID))
	return nil
}

// calcDurationSinceLastCalled calculates the duration to wait
// before running the next license check. It returns 0 if the
// license check should be run immediately.
func calcDurationSinceLastCalled(kv redispool.KeyValue, clock glock.Clock) (time.Duration, error) {
	lastCalledAt, err := kv.Get(lastCalledAtStoreKey).String()
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
