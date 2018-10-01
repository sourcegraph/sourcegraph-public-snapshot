package graphqlbackend

import (
	"time"
)

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
//
// Exactly 1 of its return values must be non-nil.
//
// It is overridden in non-OSS builds to return information about the actual product subscription in
// use.
var GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds have no license
}

// ProductLicenseInfo implements the GraphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	TagsValue      []string
	UserCountValue uint
	ExpiresAtValue time.Time
}

func (r ProductLicenseInfo) FullProductName() string { return GetFullProductName(true, r.TagsValue) }

func (r ProductLicenseInfo) Tags() []string { return r.TagsValue }

func (r ProductLicenseInfo) UserCount() int32 {
	return int32(r.UserCountValue)
}

func (r ProductLicenseInfo) ExpiresAt() string {
	return r.ExpiresAtValue.Format(time.RFC3339)
}
