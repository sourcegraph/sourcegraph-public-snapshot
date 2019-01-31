package shared

import (
	"context"
	"time"
)

// GetProductNameWithBrand is called to obtain the full product name (e.g., "Sourcegraph OSS") from a
// product license.
//
// It is overridden in non-OSS builds to return the actual name of the product subscription in
// use.
var GetProductNameWithBrand = func(hasLicense bool, licenseTags []string) string {
	return "Sourcegraph OSS"
}

// GetConfiguredProductLicenseInfo is called to obtain a site's product subscription info.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return information about the actual product subscription in
// use.
var GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds have no license
}

// GetLicenseUserCount is called to obtain a product subscription's total allowed user accounts.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return the actual product subscription user count.
var GetLicenseUserCount = func() (uint, error) {
	return 0, nil // OSS builds have no license
}

// GetLicenseExpiresAt is called to obtain a product subscription's expiry date.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return the actual product subscription expiry date.
var GetLicenseExpiresAt = func() (time.Time, error) {
	return time.Time{}, nil // OSS builds have no license
}

// GetLicenseTags is called to obtain a product subscription's tags.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return the actual product subscription's tags.
var GetLicenseTags = func() ([]string, error) {
	return []string{}, nil // OSS builds have no license
}

// ProductNameWithBrand is called to obtain the full product name (e.g., "Sourcegraph OSS") from a
// product license.
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
