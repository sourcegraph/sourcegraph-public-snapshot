package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
)

// OAuthBearerToken implements OAuth Bearer Token authentication for extsvc
// clients.
type OAuthBearerToken struct {
	Token string
}

var _ Authenticator = &OAuthBearerToken{}

func (token *OAuthBearerToken) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+token.Token)
	return nil
}

func (token *OAuthBearerToken) Hash() string {
	key := sha256.Sum256([]byte(token.Token))
	return hex.EncodeToString(key[:])
}
