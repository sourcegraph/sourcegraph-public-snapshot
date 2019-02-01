package graphqlbackend

import (
	"time"

	"github.com/sourcegraph/sourcegraph/pkg/licensing"
)

// ProductLicenseInfo implements the GraphQL type ProductLicenseInfo.
type ProductLicenseInfo struct {
	info licensing.ProductLicenseInfo
}

func (r ProductLicenseInfo) ProductNameWithBrand() string {
	return licensing.GetProductNameWithBrand(true, r.info.Tags)
}

func (r ProductLicenseInfo) Tags() []string { return r.info.Tags }

func (r ProductLicenseInfo) UserCount() int32 {
	return int32(r.info.UserCount)
}

func (r ProductLicenseInfo) ExpiresAt() string {
	return r.info.ExpiresAt.Format(time.RFC3339)
}
