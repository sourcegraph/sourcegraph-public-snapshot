package licensing

import (
	"encoding/hex"

	"github.com/sourcegraph/sourcegraph/internal/hashutil"
)

// LicenseKeyBasedAccessTokenPrefix is the prefix used for identifying tokens
// generated for product subscriptions based on a Sourcegraph license key.
const LicenseKeyBasedAccessTokenPrefix = "slk_" // "(S)ourcegraph (L)icense (K)ey"

// GenerateLicenseKeyBasedAccessToken creates a prefixed, encoded token based on a
// Sourcegraph license key.
func GenerateLicenseKeyBasedAccessToken(licenseKey string) string {
	return LicenseKeyBasedAccessTokenPrefix + hex.EncodeToString(hashutil.ToSHA256Bytes([]byte(licenseKey)))
}
