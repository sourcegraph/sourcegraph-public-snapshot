package productlicense

import (
	"time"
)

// Info contains information about a Sourcegraph license key.
type Info struct {
	Tags      []string  // tags that denote features/restrictions (e.g., "starter" or "dev")
	UserCount uint      // the number of users that this license is valid for
	ExpiresAt time.Time // the date when this license expires
}

// GetConfiguredProductLicenseInfo is called to obtain the product subscription info when creating
// the GraphQL resolver for the GraphQL type ProductLicenseInfo.
//
// A nil return value for license info indicates the instance either is an OSS build or
// doesn't have a license key defined in site configuration. If the site license key is invalid,
// a non-nil error is returned.
//
// It is overridden in non-OSS builds to return information about the actual product subscription in
// use.
var GetConfiguredProductLicenseInfo = func() (*Info, error) {
	return nil, nil // OSS builds have no license
}
