package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
//
// Exactly 1 of its return values must be non-nil.
var GetConfiguredProductLicenseInfo func() (*ProductLicenseInfo, error)

var IsFreePlan = func(*ProductLicenseInfo) bool {
	return true
}

// ProductLicenseInfo implements the GraphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	TagsValue                     []string
	UserCountValue                uint
	ExpiresAtValue                time.Time
	RevokedAtValue                *time.Time
	SalesforceSubscriptionIDValue *string
	SalesforceOpportunityIDValue  *string
	IsValidValue                  bool
	LicenseInvalidityReasonValue  *string
	HashedKeyValue                *string
}

func (r ProductLicenseInfo) ProductNameWithBrand() string {
	return GetProductNameWithBrand(!IsFreePlan(&r), r.TagsValue)
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
