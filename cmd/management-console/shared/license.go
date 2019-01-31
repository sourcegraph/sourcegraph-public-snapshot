package shared

import (
	"context"
	"time"
)

var GetProductNameWithBrand = func(hasLicense bool, licenseTags []string) string {
	return "Sourcegraph OSS"
}

var GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds have no license
}

var GetLicenseUserCount = func() (uint, error) {
	return 0, nil // OSS builds have no license
}

var GetLicenseExpiresAt = func() (time.Time, error) {
	return time.Time{}, nil // OSS builds have no license
}

var GetLicenseTags = func() ([]string, error) {
	return []string{}, nil // OSS builds have no license
}

func ProductNameWithBrand() (string, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}
	hasLicense := info != nil
	var licenseTags []string
	if hasLicense {
		licenseTags = info.TagsValue
	}
	return GetProductNameWithBrand(hasLicense, licenseTags), nil
}

// ActualUserCount is called to obtain the actual maximum number of user accounts that have been active
// on this Sourcegraph instance for the current license.
var ActualUserCount = func(ctx context.Context) (int32, error) {
	return 0, nil
}

// ActualUserCountDate is called to obtain the timestamp when the actual maximum number of user accounts
// that have been active on this Sourcegraph instance for the current license was reached.
var ActualUserCountDate = func(ctx context.Context) (string, error) {
	return "", nil
}
