package graphqlbackend

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
func GetConfiguredProductLicenseInfo() (*ProductLicenseInfo, error) {
	info, err := licensing.GetConfiguredProductLicenseInfo()
	if info == nil || err != nil {
		return nil, err
	}
	hashedKeyValue := conf.HashedCurrentLicenseKeyForAnalytics()
	return &ProductLicenseInfo{
		TagsValue:                    info.Tags,
		UserCountValue:               info.UserCount,
		ExpiresAtValue:               info.ExpiresAt,
		IsValidValue:                 licensing.IsLicenseValid(),
		LicenseInvalidityReasonValue: pointers.NonZeroPtr(licensing.GetLicenseInvalidReason()),
		HashedKeyValue:               &hashedKeyValue,
	}, nil
}

func IsFreePlan(info *ProductLicenseInfo) bool {
	for _, tag := range info.Tags() {
		if tag == fmt.Sprintf("plan:%s", licensing.PlanFree0) || tag == fmt.Sprintf("plan:%s", licensing.PlanFree1) {
			return true
		}
	}

	return false
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
	return licensing.ProductNameWithBrand(!IsFreePlan(&r), r.TagsValue)
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
