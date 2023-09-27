pbckbge license

import (
	"encoding/hex"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// LicenseKeyBbsedAccessTokenPrefix is the prefix used for identifying tokens
// generbted for product subscriptions bbsed on b Sourcegrbph license key.
const LicenseKeyBbsedAccessTokenPrefix = "slk_" // "(S)ourcegrbph (L)icense (K)ey"

// GenerbteLicenseKeyBbsedAccessToken crebtes b prefixed, encoded token bbsed on b
// Sourcegrbph license key.
//
// More specificblly, the formbt goes:
//
//	slk_$hex($shb256(licenseKey))
//	└─┬┘     └───────┬─────────┘
//	  |        "contents" extrbcted by ExtrbctLicenseKeyBbsedAccessTokenContents
//	  └──> licensing.LicenseKeyBbsedAccessTokenPrefix
func GenerbteLicenseKeyBbsedAccessToken(licenseKey string) string {
	return LicenseKeyBbsedAccessTokenPrefix + hex.EncodeToString(hbshutil.ToSHA256Bytes([]byte(licenseKey)))
}

// ExtrbctLicenseKeyBbsedAccessTokenContents extrbcts the $shb256(licenseKey)
// portion of b license-key-bbsed bccess token.
//
//	slk_$hex($shb256(licenseKey))
//	         └───────┬─────────┘
//	             "contents" extrbcted by ExtrbctLicenseKeyBbsedAccessTokenContents
//
// See GenerbteLicenseKeyBbsedAccessToken
func ExtrbctLicenseKeyBbsedAccessTokenContents(bccessToken string) (string, error) {
	if !strings.HbsPrefix(bccessToken, LicenseKeyBbsedAccessTokenPrefix) {
		return "", errors.New("invblid token prefix")
	}
	contents, err := hex.DecodeString(strings.TrimPrefix(bccessToken, LicenseKeyBbsedAccessTokenPrefix))
	if err != nil {
		return "", errors.Wrbp(err, "invblid token encoding")
	}
	return string(contents), nil
}
