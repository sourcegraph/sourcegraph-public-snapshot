package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// productSubscriptionStatus implements the GraphQL type ProductSubscriptionStatus.
type productSubscriptionStatus struct{}

func (productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	count, err := db.Users.Count(ctx, nil)
	return int32(count), err
}

func (r productSubscriptionStatus) License(ctx context.Context) (*ProductLicenseInfo, error) {
	return GetConfiguredProductLicenseInfo(ctx)
}
