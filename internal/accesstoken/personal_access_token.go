package accesstoken

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/hashutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

// personalAccessTokenPrefix is the token prefix for Sourcegraph personal access tokens. Its purpose
// is to make it easier to identify that a given string (in a file, document, etc.) is a secret
// Sourcegraph personal access token (vs. some arbitrary high-entropy hex-encoded value).
const PersonalAccessTokenPrefix = "sgph_"
const LocalInstanceIdentifier = "local"
const InstanceIdentifierLength = 10

var personalAccessTokenRegex = lazyregexp.New("^(?:sgp_|sgph_)?(?:[a-fA-F0-9]{8,16}_)?([a-fA-F0-9]{40})$")

// ParseAccessToken parses a personal access token to remove prefixes and extract the <token> that is stored in the database
// Personal access tokens can take several forms:
//   - <token>
//   - sgp_<token>
//   - sgph_<instance-identifier>_<token>
func ParsePersonalAccessToken(token string) (string, error) {
	tokenMatches := personalAccessTokenRegex.FindStringSubmatch(token)
	if len(tokenMatches) <= 1 {
		return "", errors.New("invalid token format")
	}
	tokenValue := tokenMatches[1]

	return tokenValue, nil
}

// GeneratePersonalAccessToken generates a new personal access token.
// It returns the full token string, and the byte representation of the access token.
// Personal access tokens have the form: sgph_<instance-identifier>_<token>
func GeneratePersonalAccessToken(licenseKey string, isDevInstance bool) (string, [20]byte, error) {
	var b [20]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", b, err
	}

	// Include part of the hashed license key in the token, to allow us to tie tokens back to an instance
	// If no license key is set or this is a dev instance, use a placeholder value
	var instanceIdentifier string
	if isDevInstance || licenseKey == "" {
		instanceIdentifier = LocalInstanceIdentifier
	} else {
		instanceIdentifier = hex.EncodeToString(hashutil.ToSHA256Bytes([]byte(licenseKey)))[:InstanceIdentifierLength]
	}

	token := fmt.Sprintf("%s%s_%s", PersonalAccessTokenPrefix, instanceIdentifier, hex.EncodeToString(b[:]))

	return token, b, nil
}
