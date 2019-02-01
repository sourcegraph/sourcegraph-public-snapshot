package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/pkg/licensing"
)

// productSubscriptionStatus implements the GraphQL type ProductSubscriptionStatus.
type productSubscriptionStatus struct{}

func (productSubscriptionStatus) ProductNameWithBrand() (string, error) {
	return licensing.ProductNameWithBrand()
}

func (productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	return licensing.ActualUserCount(ctx)
}

func (productSubscriptionStatus) ActualUserCountDate(ctx context.Context) (string, error) {
	return licensing.ActualUserCountDate(ctx)
}

func (productSubscriptionStatus) MaximumAllowedUserCount(ctx context.Context) (*int32, error) {
	v, err := licensing.MaximumAllowedUserCount(ctx)
	if err != nil {
		return nil, err
	}
	tmp := int32(v)
	return &tmp, nil
}

func (r productSubscriptionStatus) License() (*ProductLicenseInfo, error) {
	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}
	if info == nil {
		return nil, nil
	}
	return &ProductLicenseInfo{*info}, nil
}
