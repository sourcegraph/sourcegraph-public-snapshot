package graphqlbackend

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
var NoLicenseMaximumAllowedUserCount *int32

// NoLicenseWarningUserCount is the user count at which point a warning is shown to all users when
// there is no license, or nil if there is no limit.
var NoLicenseWarningUserCount *int32

// productSubscriptionStatus implements the GraphQL type ProductSubscriptionStatus.
type productSubscriptionStatus struct{}

func (productSubscriptionStatus) ProductNameWithBrand() (string, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}
	hasLicense := info != nil && !IsFreePlan(info)
	var licenseTags []string
	if hasLicense {
		licenseTags = info.Tags()
	}
	return GetProductNameWithBrand(hasLicense, licenseTags), nil
}

func (productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	return ActualUserCount(ctx)
}

func (productSubscriptionStatus) ActualUserCountDate(ctx context.Context) (string, error) {
	return ActualUserCountDate(ctx)
}

func (productSubscriptionStatus) NoLicenseWarningUserCount(ctx context.Context) (*int32, error) {
	if info, err := GetConfiguredProductLicenseInfo(); info != nil && !IsFreePlan(info) {
		// if a license exists, warnings never need to be shown.
		return nil, err
	}
	return NoLicenseWarningUserCount, nil
}

func (productSubscriptionStatus) MaximumAllowedUserCount(ctx context.Context) (*int32, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}
	if info != nil && !IsFreePlan(info) {
		tmp := info.UserCount()
		return &tmp, nil
	}
	return NoLicenseMaximumAllowedUserCount, nil
}

func (r productSubscriptionStatus) License() (*ProductLicenseInfo, error) {
	return GetConfiguredProductLicenseInfo()
}
