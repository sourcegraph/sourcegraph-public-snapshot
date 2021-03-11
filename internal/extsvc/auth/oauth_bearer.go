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

// OAuthBearerTokenWithSSH implements OAuth Bearer Token authentication for extsvc
// clients and holds an additional RSA keypair.
type OAuthBearerTokenWithSSH struct {
	OAuthBearerToken

	PrivateKey string
	PublicKey  string
	Passphrase string
}

var _ Authenticator = &OAuthBearerTokenWithSSH{}
var _ AuthenticatorWithSSH = &OAuthBearerTokenWithSSH{}

func (token *OAuthBearerTokenWithSSH) SSHPrivateKey() (privateKey, passphrase string) {
	return token.PrivateKey, token.Passphrase
}

func (token *OAuthBearerTokenWithSSH) SSHPublicKey() string {
	return token.PublicKey
}

func (token *OAuthBearerTokenWithSSH) Hash() string {
	shaSum := sha256.Sum256([]byte(token.Token + token.PrivateKey + token.Passphrase + token.PublicKey))
	return hex.EncodeToString(shaSum[:])
}
