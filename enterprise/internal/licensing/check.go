package licensing

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io"
	"time"

	"net/http"

	"github.com/sourcegraph/log"

	"context"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	licenseCheckStarted = false
	store               = redispool.Store

	lastCalledAtStoreKey    = "licensing:last_called_at"
	licenseValidityStoreKey = "licensing:is_license_valid"
	prevLicenseTokenKey     = "licensing:prev_license_hash"

	routine *goroutine.PeriodicGoroutine
)

type licenseChecker struct {
	siteID string
	token  string
	info   *Info
	doer   httpcli.Doer
}

func (l *licenseChecker) Handle(ctx context.Context) error {
	store.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339))

	if l.info.HasTag(AllowAirGappedTag) || l.info.SalesforceSubscriptionID != nil {
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
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	json.Unmarshal([]byte(resBody), &body)

	if body.Error != "" {
		return errors.New(body.Error)
	}

	if body.Data == nil {
		return errors.New("No data returned from license check")
	}

	store.Set(licenseValidityStoreKey, body.Data.IsValid)
	return nil
}

func StartLicenseCheck(ctx context.Context, interval time.Duration, logger log.Logger, globalStateStore database.GlobalStateStore) {
	if licenseCheckStarted {
		logger.Info("license check already started")
		return
	}
	licenseCheckStarted = true

	// The entire logic is dependent on config so we will
	// wait for initial config to be loaded as well as
	// watch for any config changes
	conf.Watch(func() {
		prevLicenseToken, err := store.Get(prevLicenseTokenKey).String()
		if err != nil {
			logger.Error("error getting previous license hash", log.Error(err))
			return
		}
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

		globalState, err := globalStateStore.Get(ctx)
		if err != nil {
			logger.Error("error getting global state", log.Error(err))
			return
		}
		info, err := GetConfiguredProductLicenseInfo()
		if err != nil {
			logger.Error("error getting configured license info", log.Error(err))
			return
		}
		routine := goroutine.NewPeriodicGoroutine(
			context.Background(),
			&licenseChecker{siteID: globalState.SiteID, token: licenseToken, info: info, doer: httpcli.ExternalDoer},
			goroutine.WithName("licensing.check-license-validity"),
			goroutine.WithDescription("check if license is valid from sourcegraph.com"),
			goroutine.WithInterval(interval),
		)
		go goroutine.MonitorBackgroundRoutines(ctx, routine)
	})
}
