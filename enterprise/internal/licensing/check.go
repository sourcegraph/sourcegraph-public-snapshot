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
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	licenseCheckStarted     = false
	store                   = redispool.Store
	delay                   = 12 * time.Hour
	lastCalledAtStoreKey    = "licensing:last_called_at"
	licenseValidityStoreKey = "licensing:is_license_valid"
)

func StartLicenseCheck(logger log.Logger, globalStateStore database.GlobalStateStore) {
	if licenseCheckStarted {
		panic("already started")
	}
	licenseCheckStarted = true

	// initial sleep on server restarts
	var durationToWait time.Duration = 0
	lastCalledStr, err := store.Get(lastCalledAtStoreKey).String()
	if err != nil {
		logger.Error("error getting last-called-at from cache", log.Error(err))
	}
	if lastCalledAt, err := time.Parse(time.RFC3339, lastCalledStr); err == nil {
		durationToWait = maxDelayOrZero(lastCalledAt, time.Now(), delay)
	}
	time.Sleep(durationToWait)

	ctx := context.Background()
	for {
		info, err := GetConfiguredProductLicenseInfo()
		if err != nil {
			logger.Error("error getting configured license info", log.Error(err))
		} else if info.HasTag(AllowAirGappedTag) { // todo: discuss what will happen with existing air-gapped v1 license instances
			logger.Info("skipping license check")
		} else if globalState, err := globalStateStore.Get(ctx); err != nil {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			valid, err := checkLicenseValidity(ctx, httpcli.ExternalClient, globalState.SiteID, conf.Get().LicenseKey)
			cancel()
			if err != nil {
				logger.Error("error while checking license validity", log.Error(err))
			} else {
				store.Set(licenseValidityStoreKey, valid)
			}
			store.Set(lastCalledAtStoreKey, time.Now().Format(time.RFC3339))
		} else {
			logger.Error("error getting global state", log.Error(err))
		}
		time.Sleep(delay)
	}
}

func maxDelayOrZero(before time.Time, after time.Time, maxDelay time.Duration) time.Duration {
	if before.IsZero() || after.IsZero() || before.After(after) {
		return 0
	}

	timePassed := time.Since(before)

	if timePassed < delay {
		return delay - timePassed
	}

	return 0
}

func checkLicenseValidity(ctx context.Context, doer httpcli.Doer, siteID, licenseKey string) (bool, error) {
	payload, err := json.Marshal(struct {
		ClientSiteID string `json:"siteID"`
	}{ClientSiteID: siteID})

	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://sourcegraph.com/.api/license/check", bytes.NewBuffer(payload))
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+hex.EncodeToString(GenerateHashedLicenseKeyAccessToken(licenseKey)))
	req.Header.Set("Content-Type", "application/json")

	res, err := doer.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return false, errors.Newf("Failed to check license, status code: %d", res.StatusCode)
	}

	var body LicenseCheckResponse
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return false, err
	}

	json.Unmarshal([]byte(resBody), &body)

	if body.Error != "" {
		return false, errors.New(body.Error)
	}

	if body.Data == nil {
		return false, errors.New("No data returned from license check")
	}

	return body.Data.IsValid, nil
}
