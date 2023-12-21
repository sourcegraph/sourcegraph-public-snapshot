package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// productSubscriptionStatus implements the GraphQL type ProductSubscriptionStatus.
type productSubscriptionStatus struct{}

func (productSubscriptionStatus) ProductNameWithBrand() (string, error) {
	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}

	return licensing.ProductNameWithBrand(info.Tags()), nil
}

func (productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	return licensing.ActualUserCount(ctx)
}

func (productSubscriptionStatus) ActualUserCountDate(ctx context.Context) (string, error) {
	return licensing.ActualUserCountDate(ctx)
}

func (productSubscriptionStatus) NoLicenseWarningUserCount(ctx context.Context) (*int32, error) {
	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}

	if !info.Plan.IsFree() {
		// if a license exists, warnings never need to be shown.
		return nil, err
	}

	return pointers.Ptr(licensing.NoLicenseWarningUserCount), nil
}

func (productSubscriptionStatus) MaximumAllowedUserCount(ctx context.Context) (*int32, error) {
	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}
	if !info.Plan.IsFree() {
		tmp := info.UserCount()
		return &tmp, nil
	}
	return pointers.Ptr(licensing.NoLicenseMaximumAllowedUserCount), nil
}

func (r productSubscriptionStatus) License() (*ProductLicenseInfo, error) {
	return getConfiguredProductLicenseInfo()
}
