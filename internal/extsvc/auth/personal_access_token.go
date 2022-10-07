package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"
)

// PersonalAccessToken implements Personal Access Token authentication for extsvc
// clients.
type PersonalAccessToken struct {
	Token  string     `json:"token"`
	Expiry *time.Time `json:"expiry,omitempty"`
}

func (token *PersonalAccessToken) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+token.Token)
	return nil
}

func (token *PersonalAccessToken) Hash() string {
	key := sha256.Sum256([]byte(token.Token))
	return hex.EncodeToString(key[:])
}

// WithToken returns an Oauth Bearer Token authenticator with a new token
func (token *PersonalAccessToken) WithToken(newToken string) *PersonalAccessToken {
	return &PersonalAccessToken{
		Token: newToken,
	}
}

// PersonalAccessTokenWithSSH implements Personal Access Token authentication for extsvc
// clients and holds an additional RSA keypair.
type PersonalAccessTokenWithSSH struct {
	PersonalAccessToken

	PrivateKey string
	PublicKey  string
	Passphrase string
}

func (token *PersonalAccessTokenWithSSH) SSHPrivateKey() (privateKey, passphrase string) {
	return token.PrivateKey, token.Passphrase
}

func (token *PersonalAccessTokenWithSSH) SSHPublicKey() string {
	return token.PublicKey
}

func (token *PersonalAccessTokenWithSSH) Hash() string {
	shaSum := sha256.Sum256([]byte(token.Token + token.PrivateKey + token.Passphrase + token.PublicKey))
	return hex.EncodeToString(shaSum[:])
}
