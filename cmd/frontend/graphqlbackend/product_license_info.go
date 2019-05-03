package graphqlbackend

import (
	"time"
)

// ProductLicenseInfo implements the GraphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	TagsValue      []string
	UserCountValue uint
	ExpiresAtValue time.Time
}

func (r ProductLicenseInfo) ProductNameWithBrand() string {
	return GetProductNameWithBrand(true, r.TagsValue)
}

func (r ProductLicenseInfo) Tags() []string { return r.TagsValue }

func (r ProductLicenseInfo) UserCount() int32 {
	return int32(r.UserCountValue)
}

func (r ProductLicenseInfo) ExpiresAt() string {
	return r.ExpiresAtValue.Format(time.RFC3339)
}
