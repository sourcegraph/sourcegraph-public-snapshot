package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/redispool"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// productSubscriptionStatus implements the GraphQL type ProductSubscriptionStatus.
type productSubscriptionStatus struct {
	kv redispool.KeyValue
}

func (productSubscriptionStatus) ProductNameWithBrand() (string, error) {
	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}

	return licensing.ProductNameWithBrand(info.Tags()), nil
}

func (r *productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	return licensing.ActualUserCount(ctx, r.kv)
}

func (r *productSubscriptionStatus) ActualUserCountDate(ctx context.Context) (string, error) {
	return licensing.ActualUserCountDate(ctx, r.kv)
}

func (productSubscriptionStatus) NoLicenseWarningUserCount(ctx context.Context) (*int32, error) {
	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}

	// We only show this warning to free license instances.
	if !info.Plan.IsFreePlan() {
		return nil, nil
	}

	return pointers.Ptr(licensing.NoLicenseWarningUserCount), nil
}

func (productSubscriptionStatus) MaximumAllowedUserCount(ctx context.Context) (*int32, error) {
	info, err := getConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}
	if !info.Plan.IsFreePlan() {
		tmp := info.UserCount()
		return &tmp, nil
	}
	return pointers.Ptr(licensing.NoLicenseMaximumAllowedUserCount), nil
}

func (r productSubscriptionStatus) License() (*ProductLicenseInfo, error) {
	return getConfiguredProductLicenseInfo()
}
