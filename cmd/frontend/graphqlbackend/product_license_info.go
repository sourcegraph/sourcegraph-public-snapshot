package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
//
// Exactly 1 of its return values must be non-nil.
func getConfiguredProductLicenseInfo() (*ProductLicenseInfo, error) {
	info, err := licensing.GetConfiguredProductLicenseInfo()
	if err != nil {
		return nil, err
	}
	hashedKeyValue := conf.HashedCurrentLicenseKeyForAnalytics()
	return &ProductLicenseInfo{
		Plan:                         info.Plan(),
		TagsValue:                    info.Tags,
		UserCountValue:               info.UserCount,
		ExpiresAtValue:               info.ExpiresAt,
		IsValidValue:                 licensing.IsLicenseValid(),
		LicenseInvalidityReasonValue: pointers.NonZeroPtr(licensing.GetLicenseInvalidReason()),
		HashedKeyValue:               &hashedKeyValue,
	}, nil
}

// ProductLicenseInfo implements the GraphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	Plan                          licensing.Plan
	TagsValue                     []string
	UserCountValue                uint
	ExpiresAtValue                time.Time
	SalesforceSubscriptionIDValue *string
	SalesforceOpportunityIDValue  *string
	IsValidValue                  bool
	LicenseInvalidityReasonValue  *string
	HashedKeyValue                *string
}

func (r ProductLicenseInfo) ProductNameWithBrand() string {
	return licensing.ProductNameWithBrand(r.TagsValue)
}

func (r ProductLicenseInfo) IsFreePlan() bool {
	return r.Plan.IsFreePlan()
}

func (r ProductLicenseInfo) Tags() []string { return r.TagsValue }

func (r ProductLicenseInfo) UserCount() int32 {
	return int32(r.UserCountValue)
}

func (r ProductLicenseInfo) ExpiresAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.ExpiresAtValue}
}

func (r ProductLicenseInfo) SalesforceSubscriptionID() *string {
	return r.SalesforceSubscriptionIDValue
}

func (r ProductLicenseInfo) SalesforceOpportunityID() *string {
	return r.SalesforceOpportunityIDValue
}

func (r ProductLicenseInfo) IsValid() bool {
	return r.IsValidValue
}

func (r ProductLicenseInfo) LicenseInvalidityReason() *string {
	return r.LicenseInvalidityReasonValue
}

func (r ProductLicenseInfo) HashedKey() *string {
	return r.HashedKeyValue
}
