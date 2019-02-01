package licensing

import "context"

// GetProductNameWithBrand is called to obtain the full product name (e.g., "Sourcegraph OSS") from a
// product license.
var GetProductNameWithBrand = func(hasLicense bool, licenseTags []string) string {
	return "Sourcegraph OSS"
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

// NoLicenseMaximumAllowedUserCount is the maximum allowed user count when there is no license, or
// nil if there is no limit.
var NoLicenseMaximumAllowedUserCount uint

func ProductNameWithBrand() (string, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}
	hasLicense := info != nil
	var licenseTags []string
	if hasLicense {
		licenseTags = info.Tags
	}
	return GetProductNameWithBrand(hasLicense, licenseTags), nil
}

func MaximumAllowedUserCount(ctx context.Context) (uint, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return 0, err
	}
	if info != nil {
		return info.UserCount, nil
	}
	return NoLicenseMaximumAllowedUserCount, nil
}
