package license

import (
	"encoding/hex"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/hashutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// LicenseKeyBasedAccessTokenPrefix is the prefix used for identifying tokens
// generated for product subscriptions based on a Sourcegraph license key.
const LicenseKeyBasedAccessTokenPrefix = "slk_" // "(S)ourcegraph (L)icense (K)ey"

// GenerateLicenseKeyBasedAccessToken creates a prefixed, encoded token based on a
// Sourcegraph license key.
//
// More specifically, the format goes:
//
//	slk_$hex($sha256(licenseKey))
//	└─┬┘     └───────┬─────────┘
//	  |        "contents" extracted by ExtractLicenseKeyBasedAccessTokenContents
//	  └──> licensing.LicenseKeyBasedAccessTokenPrefix
func GenerateLicenseKeyBasedAccessToken(licenseKey string) string {
	return LicenseKeyBasedAccessTokenPrefix + hex.EncodeToString(hashutil.ToSHA256Bytes([]byte(licenseKey)))
}

// ExtractLicenseKeyBasedAccessTokenContents extracts the $sha256(licenseKey)
// portion of a license-key-based access token.
//
//	slk_$hex($sha256(licenseKey))
//	         └───────┬─────────┘
//	             "contents" extracted by ExtractLicenseKeyBasedAccessTokenContents
//
// See GenerateLicenseKeyBasedAccessToken
func ExtractLicenseKeyBasedAccessTokenContents(accessToken string) (string, error) {
	if !strings.HasPrefix(accessToken, LicenseKeyBasedAccessTokenPrefix) {
		return "", errors.New("invalid token prefix")
	}
	contents, err := hex.DecodeString(strings.TrimPrefix(accessToken, LicenseKeyBasedAccessTokenPrefix))
	if err != nil {
		return "", errors.Wrap(err, "invalid token encoding")
	}
	return string(contents), nil
}
