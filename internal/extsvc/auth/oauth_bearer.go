package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// OAuthBearerToken implements OAuth Bearer Token authentication for extsvc
// clients.
type OAuthBearerToken string

var _ Authenticator = OAuthBearerToken("")

func (token OAuthBearerToken) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+string(token))
	return nil
}

func (token OAuthBearerToken) Hash() string {
	return hashString(string(token))
}

func hashString(s string) string {
	key := sha256.Sum256([]byte(s))
	return hex.EncodeToString(key[:])
}
