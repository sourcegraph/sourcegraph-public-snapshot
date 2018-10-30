package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
)

// GetProductNameWithBrand is called to obtain the full product name (e.g., "Sourcegraph OSS") from a
// product license.
var GetProductNameWithBrand = func(hasLicense bool, licenseTags []string) string {
	return "Sourcegraph OSS"
}

// productSubscriptionStatus implements the GraphQL type ProductSubscriptionStatus.
type productSubscriptionStatus struct{}

func (productSubscriptionStatus) ProductNameWithBrand() (string, error) {
	info, err := GetConfiguredProductLicenseInfo()
	if err != nil {
		return "", err
	}
	hasLicense := info != nil
	var licenseTags []string
	if hasLicense {
		licenseTags = info.Tags()
	}
	return GetProductNameWithBrand(hasLicense, licenseTags), nil
}

func (productSubscriptionStatus) ActualUserCount(ctx context.Context) (int32, error) {
	count, err := db.Users.Count(ctx, nil)
	return int32(count), err
}

func (r productSubscriptionStatus) License() (*ProductLicenseInfo, error) {
	return GetConfiguredProductLicenseInfo()
}
