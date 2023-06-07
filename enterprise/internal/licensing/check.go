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
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var licenseCheckStarted = false

func StartLicenseCheck(logger log.Logger, globalStateStore database.GlobalStateStore) {
	if licenseCheckStarted {
		panic("already started")
	}
	licenseCheckStarted = true

	ctx := context.Background()
	const delay = 12 * time.Hour
	checker := &licenseChecker{doer: httpcli.ExternalClient}

	for {
		info, err := GetConfiguredProductLicenseInfo()
		if err != nil {
			logger.Error("error getting configured license info", log.Error(err))
		} else if checker.shouldSkip(info) {
			logger.Info("skipping license check")
		} else if globalState, err := globalStateStore.Get(ctx); err != nil {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			valid, err := checker.check(ctx, globalState, info, conf.Get().LicenseKey)
			if err != nil {
				logger.Error("error license validity", log.Error(err))
			} else {
				globalStateStore.Update(ctx, valid)
			}
			cancel()
		} else {
			logger.Error("error getting global state", log.Error(err))
		}
		time.Sleep(delay)
	}
}

type licenseChecker struct {
	doer httpcli.Doer
}

func (l *licenseChecker) shouldSkip(info *Info) bool {
	return info.Version() < 2 || info.HasTag(AirGappedTag)
}

func (l *licenseChecker) check(ctx context.Context, globalState database.GlobalState, info *Info, licenseKey string) (value bool, err error) {
	payload, err := json.Marshal(struct {
		ClientSiteID string `json:"siteID"`
	}{ClientSiteID: globalState.SiteID})

	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodPost, "https://sourcegraph.com/.api/license/check", bytes.NewBuffer(payload))
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", "Bearer "+hex.EncodeToString(GenerateHashedLicenseKeyAccessToken(licenseKey)))
	req.Header.Set("Content-Type", "application/json")

	res, err := l.doer.Do(req)
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

	if body.Error != nil {
		return false, errors.New(*body.Error)
	}

	if body.Data == nil {
		return false, errors.New("No data returned from license check")
	}

	return body.Data.IsValid, nil
}
