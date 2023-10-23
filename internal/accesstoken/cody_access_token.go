package accesstoken

import (
	"encoding/hex"

	"github.com/sourcegraph/sourcegraph/internal/hashutil"
)

// DotcomUserGatewayAccessTokenPrefix is the prefix used for identifying tokens
// generated for a dotcom api token .
const DotcomUserGatewayAccessTokenPrefix = "sgd_" // "(S)ource(g)raph (d)otcom user key"

// GenerateLicenseKeyBasedAccessToken creates a prefixed, encoded token based on a
// Sourcegraph license key.
func GenerateDotcomUserGatewayAccessToken(apiToken string) (string, error) {
	token, err := ParsePersonalAccessToken(apiToken)
	if err != nil {
		return "", err
	}

	tokenBytes, err := hex.DecodeString(token)
	if err != nil {
		return "", err
	}

	return "sgd_" + hex.EncodeToString(hashutil.ToSHA256Bytes(hashutil.ToSHA256Bytes(tokenBytes))), nil
}
