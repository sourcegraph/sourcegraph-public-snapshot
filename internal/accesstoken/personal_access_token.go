package accesstoken

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// personalAccessTokenPrefix is the token prefix for Sourcegraph personal access tokens. Its purpose
// is to make it easier to identify that a given string (in a file, document, etc.) is a secret
// Sourcegraph personal access token (vs. some arbitrary high-entropy hex-encoded value).
const PersonalAccessTokenPrefix = "sgp_"
const LocalInstanceIdentifier = "local"
const InstanceIdentifierLength = 16
const InstanceIdentifierHmacKey = "instance_identifier_hmac_key" // Public as we are using HMAC for key derivation, not for authentication

var personalAccessTokenRegex = lazyregexp.New("^(?:sgp_|sgph_)?(?:[a-fA-F0-9]{16}_|local_)?([a-fA-F0-9]{40})$")

// ParseAccessToken parses a personal access token to remove prefixes and extract the <token> that is stored in the database
// Personal access tokens can take several forms:
//   - <token>
//   - sgp_<token>
//   - sgp_<instance-identifier>_<token>
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
// Personal access tokens have the form: sgp_<instance-identifier>_<token>
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
		h := hmac.New(sha256.New, []byte(InstanceIdentifierHmacKey))
		h.Write([]byte(licenseKey))

		instanceIdentifier = hex.EncodeToString(h.Sum(nil))[:InstanceIdentifierLength]
	}
	instanceIdentifier = instanceIdentifier + "_"

	token := fmt.Sprintf("%s%s%s", PersonalAccessTokenPrefix, instanceIdentifier, hex.EncodeToString(b[:]))

	return token, b, nil
}
