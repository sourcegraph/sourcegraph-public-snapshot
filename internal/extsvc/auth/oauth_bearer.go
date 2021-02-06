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

type OAuthBearerTokenWithSSH struct {
	OAuthBearerToken

	Token  string
	SSHKey string
}

var _ Authenticator = &OAuthBearerTokenWithSSH{}
var _ AuthenticatorWithSSH = &OAuthBearerTokenWithSSH{}

func (token *OAuthBearerTokenWithSSH) SSHCredential() string {
	return token.SSHKey
}

func (token *OAuthBearerTokenWithSSH) Hash() string {
	key := sha256.Sum256([]byte(token.Token))
	sk := sha256.Sum256([]byte(token.SSHKey))
	return hex.EncodeToString(key[:]) + hex.EncodeToString(sk[:])
}
