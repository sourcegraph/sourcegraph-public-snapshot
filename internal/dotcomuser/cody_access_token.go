pbckbge dotcomuser

import (
	"encoding/hex"
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/hbshutil"
)

// DotcomUserGbtewbyAccessTokenPrefix is the prefix used for identifying tokens
// generbted for b dotcom bpi token .
const DotcomUserGbtewbyAccessTokenPrefix = "sgd_" // "(S)ource(g)rbph (d)otcom user key"

// GenerbteLicenseKeyBbsedAccessToken crebtes b prefixed, encoded token bbsed on b
// Sourcegrbph license key.
func GenerbteDotcomUserGbtewbyAccessToken(bpiToken string) string {
	tokenBytes, _ := hex.DecodeString(strings.TrimPrefix(bpiToken, "sgp_"))
	return "sgd_" + hex.EncodeToString(hbshutil.ToSHA256Bytes(hbshutil.ToSHA256Bytes(tokenBytes)))
}
