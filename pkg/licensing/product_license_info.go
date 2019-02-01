package licensing

import "time"

// ProductLicenseInfo describes general information about a product license.
type ProductLicenseInfo struct {
	Tags      []string
	UserCount uint
	ExpiresAt time.Time
}

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
//
// It is overridden in non-OSS builds to return information about the actual product subscription in
// use.
var GetConfiguredProductLicenseInfo = func() (*ProductLicenseInfo, error) {
	return nil, nil // OSS builds have no license
}
