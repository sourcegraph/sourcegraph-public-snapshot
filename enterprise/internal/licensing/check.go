package licensing

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"time"

	"net/http"

	"github.com/sourcegraph/log"

	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	licenseCheckStarted = false
	store               = redispool.Store

	routine *goroutine.PeriodicGoroutine
)

const (
	licenseCheckInterval    = 12 * time.Hour
	lastCalledAtStoreKey    = "licensing:last_called_at"
	licenseValidityStoreKey = "licensing:is_license_valid"
	prevLicenseTokenKey     = "licensing:prev_license_hash"
)

type licenseChecker struct {
	siteID string
	token  string
	doer   httpcli.Doer
}

func (l *licenseChecker) Handle(ctx context.Context) error {
	store.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339))

	// skip if has explicitly allowed air-gapped feature
	if err := Check(FeatureAllowAirGapped); err == nil {
		store.Set(licenseValidityStoreKey, true)
		return nil
	}

	payload, err := json.Marshal(struct {
		ClientSiteID string `json:"siteID"`
	}{ClientSiteID: l.siteID})

	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, "https://sourcegraph.com/.api/license/check", bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+l.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := l.doer.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.Newf("Failed to check license, status code: %d", res.StatusCode)
	}

	var body LicenseCheckResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		return err
	}

	if body.Error != "" {
		return errors.New(body.Error)
	}

	if body.Data == nil {
		return errors.New("No data returned from license check")
	}

	store.Set(licenseValidityStoreKey, body.Data.IsValid)
	return nil
}

// StartLicenseCheck starts a goroutine that periodically checks
// license validity from dotcom and stores the result in redis.
// It re-runs the check if the license key changes.
func StartLicenseCheck(ctx context.Context, logger log.Logger, siteID string) {
	if licenseCheckStarted {
		logger.Info("license check already started")
		return
	}
	licenseCheckStarted = true

	// The entire logic is dependent on config so we will
	// wait for initial config to be loaded as well as
	// watch for any config changes
	conf.Watch(func() {
		prevLicenseToken, _ := store.Get(prevLicenseTokenKey).String()
		licenseToken := hex.EncodeToString(GenerateHashedLicenseKeyAccessToken(conf.Get().LicenseKey))
		// skip if license key hasn't changed and already running
		if prevLicenseToken == licenseToken && routine != nil {
			return
		}

		// stop previously running routine
		if routine != nil {
			routine.Stop()
		}

		// continue running with new license key
		store.Set(prevLicenseTokenKey, licenseToken)

		routine := goroutine.NewPeriodicGoroutine(
			context.Background(),
			&licenseChecker{siteID: siteID, token: licenseToken, doer: httpcli.ExternalDoer},
			goroutine.WithName("licensing.check-license-validity"),
			goroutine.WithDescription("check if license is valid from sourcegraph.com"),
			goroutine.WithInterval(licenseCheckInterval),
		)
		go goroutine.MonitorBackgroundRoutines(ctx, routine)
	})
}
