package accesstoken

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/hashutil"
)

// personalAccessTokenPrefix is the token prefix for Sourcegraph personal access tokens. Its purpose
// is to make it easier to identify that a given string (in a file, document, etc.) is a secret
// Sourcegraph personal access token (vs. some arbitrary high-entropy hex-encoded value).
const PersonalAccessTokenPrefix = "sgph_"
const LocalInstanceIdentifier = "local" // TODO: Does this string need to be fixed length?

// ParseAccessToken parses a personal access token to extract the token that is stored in the database
// Personal access tokens can take several forms:
//   - <token>
//   - sgp_<token>
//   - sgph_<instance-identifier>_<token>
func ParsePersonalAccessToken(token string) (string, error) {
	// Iterate through all prefixes used by previous versions of Sourcegraph and remove them
	oldPersonalAccessTokenPrefixes := []string{"sgp_"}
	for _, prefix := range append(oldPersonalAccessTokenPrefixes, PersonalAccessTokenPrefix) {
		token = strings.TrimPrefix(token, prefix)
	}

	// Split a token of the form <instance_identifier>_<token> to extract just <token>.
	// If no instance identifier is present, return the full token.
	tokenParts := strings.Split(token, "_")
	switch len(tokenParts) {
	case 1:
		// No instance identifier present, return full token
		token = tokenParts[0]
	case 2:
	// Instance identifier present, return second part of token
	default:
		return "", errors.New("invalid token format")
	}

	return token, nil
}

// GeneratePersonalAccessToken generates a new personal access token.
// It returns the full token string, and the byte representation of the access token.
// Personal access tokens have the form: sgph_<instance-identifier>_<token>
func GeneratePersonalAccessToken() (string, [20]byte, error) {
	var b [20]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", b, err
	}

	// TODO: Ensure this works for local dev instances - do they have pre-set license keys?
	// Include part of the hashed license key in the token, to allow us to tie tokens back to an instance
	config := conf.Get().SiteConfig()
	var licenseKeyHash string
	if config.LicenseKey != "" {
		licenseKeyHash = hex.EncodeToString(hashutil.ToSHA256Bytes([]byte(config.LicenseKey)))[:6]
	} else {
		licenseKeyHash = LocalInstanceIdentifier
	}

	token := fmt.Sprintf("%s%s_%s", PersonalAccessTokenPrefix, licenseKeyHash, hex.EncodeToString(b[:]))

	return token, b, nil
}
