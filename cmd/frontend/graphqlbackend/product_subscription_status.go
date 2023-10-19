package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/licensing"
)

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
	return licensing.ProductNameWithBrand(hasLicense, licenseTags), nil
}

func (productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	return licensing.ActualUserCount(ctx)
}

func (productSubscriptionStatus) ActualUserCountDate(ctx context.Context) (string, error) {
	return licensing.ActualUserCountDate(ctx)
}

func (productSubscriptionStatus) NoLicenseWarningUserCount(ctx context.Context) (*int32, error) {
	if info, err := GetConfiguredProductLicenseInfo(); info != nil && !IsFreePlan(info) {
		// if a license exists, warnings never need to be shown.
		return nil, err
	}
	c := licensing.NoLicenseWarningUserCount
	return &c, nil
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
	c := licensing.NoLicenseMaximumAllowedUserCount
	return &c, nil
}

func (r productSubscriptionStatus) License() (*ProductLicenseInfo, error) {
	return GetConfiguredProductLicenseInfo()
}
