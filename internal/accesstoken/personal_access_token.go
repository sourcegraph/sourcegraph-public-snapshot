package accesstoken

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/hashutil"
)

// personalAccessTokenPrefix is the token prefix for Sourcegraph personal access tokens. Its purpose
// is to make it easier to identify that a given string (in a file, document, etc.) is a secret
// Sourcegraph personal access token (vs. some arbitrary high-entropy hex-encoded value).
const PersonalAccessTokenPrefix = "sgph_"
const LocalInstanceIdentifier = "local"
const InstanceIdentifierLength = 10

// ParseAccessToken parses a personal access token to remove prefixes and extract the <token> that is stored in the database
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

	// TODO: Side-effect of this is that it will strip any prefix, e.g. asdf_<token>
	// Remove <instance-identifier> from token, if present
	tokenParts := strings.Split(token, "_")
	switch len(tokenParts) {
	case 1:
		// No instance identifier present, return full token
		token = tokenParts[0]
	case 2:
		// Instance identifier present, return second part of token
		token = tokenParts[1]
	default:
		return "", errors.New("invalid token format")
	}

	return token, nil
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
