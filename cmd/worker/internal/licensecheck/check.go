package licensecheck

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"github.com/derision-test/glock"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"

	subscriptionlicensechecksv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1"
	subscriptionlicensechecksv1connect "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/subscriptionlicensechecks/v1/v1connect"
)

type ConfClient interface {
	Get() *conf.Unified
}

// newLicenseChecker returns a goroutine that periodically checks license validity
// from dotcom and stores the result in redis.
// It re-runs the check if the license key changes.
func newLicenseChecker(
	ctx context.Context,
	logger log.Logger,
	db database.DB,
	kv redispool.KeyValue,
	confClient ConfClient,
	checks subscriptionlicensechecksv1connect.SubscriptionLicenseChecksServiceClient,
) goroutine.BackgroundRoutine {
	conf.MockClient()
	return goroutine.NewPeriodicGoroutine(
		ctx,
		&licenseChecker{
			db:     db,
			logger: logger.Scoped("licenseChecker"),
			kv:     kv,
			conf:   confClient,
			checks: checks,
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
	logger log.Logger
	kv     redispool.KeyValue

	conf   ConfClient
	checks subscriptionlicensechecksv1connect.SubscriptionLicenseChecksServiceClient
}

func (l *licenseChecker) Handle(ctx context.Context) (err error) {
	gs, err := l.db.GlobalState().Get(ctx)
	if err != nil {
		return errors.Wrap(err, "error reading global state from DB")
	}
	siteID := gs.SiteID

	logger := trace.Logger(ctx, l.logger).
		With(log.String("siteID", siteID))
	logger.Debug("starting license check")

	// skip if has explicitly allowed air-gapped feature
	if err := licensing.Check(licensing.FeatureAllowAirGapped); err == nil {
		logger.Debug("license is air-gapped, skipping check", log.String("siteID", siteID))
		return l.setLicenseCheckResult(true)
	}

	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return errors.Wrap(err, "parse configured license")
	}
	if info.Plan().IsFreePlan() {
		logger.Debug("free plan, skipping license verification check")
		return l.setLicenseCheckResult(true)
	}

	// Check if the license key has changed and generate an auth token for the request.
	licenseKey := l.conf.Get().LicenseKey
	prevLicenseToken, _ := l.kv.Get(prevLicenseTokenKey).String()
	licenseToken := license.GenerateLicenseKeyBasedAccessToken(licenseKey)
	// If the key hasn't changed, let's make sure we only hit this endpoint about
	// every 12 hours.
	if prevLicenseToken == licenseToken {
		if waitDuration, _ := calcDurationSinceLastCalled(l.kv, glock.NewRealClock()); waitDuration > 0 {
			logger.Debug("license key check is not due, skipping check")
			return nil
		}
	}

	// Continue running with new license key.
	if err := l.kv.Set(prevLicenseTokenKey, licenseToken); err != nil {
		logger.Error("error storing license token in redis", log.Error(err))
	}

	resp, err := l.checks.CheckLicenseKey(ctx, connect.NewRequest(&subscriptionlicensechecksv1.CheckLicenseKeyRequest{
		LicenseKey: licenseKey,
		InstanceId: siteID,
	}))
	if err != nil {
		var connectErr *connect.Error
		if errors.As(err, &connectErr) {
			switch connectErr.Code() {
			case connect.CodeNotFound:
				return l.setLicenseCheckResult(false)
			}
		}
		logger.Warn("unexpected error while checking license validity", log.Error(err))
		return err
	}

	if resp == nil || resp.Msg == nil {
		logger.Warn("no data returned from license check")
		return errors.New("No data returned from license check")
	}

	// best effort, ignore errors here
	_ = l.kv.Set(licensing.LicenseInvalidReason, resp.Msg.Reason)

	if err := l.setLicenseCheckResult(resp.Msg.Valid); err != nil {
		logger.Warn("set license check result", log.Error(err))
		return err
	}

	logger.Debug("finished license check")
	return nil
}

// setLicenseCheckResult updates the last called timestamp and license validity
// status in the key-value store. It returns an error if either operation fails.
func (l *licenseChecker) setLicenseCheckResult(isLicenseValid bool) error {
	if err := l.kv.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339)); err != nil {
		return errors.Wrap(err, "set last license check time")
	}
	if err := l.kv.Set(licensing.LicenseValidityStoreKey, isLicenseValid); err != nil {
		return errors.Wrapf(err, "set license validity state to %v", isLicenseValid)
	}
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
