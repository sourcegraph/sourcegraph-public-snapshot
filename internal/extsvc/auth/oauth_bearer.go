package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// OAuthBearerToken implements OAuth Bearer Token authentication for extsvc
// clients.
type OAuthBearerToken struct {
	Token        string                                                                                    `json:"token"`
	TokenType    string                                                                                    `json:"token_type,omitempty"`
	RefreshToken string                                                                                    `json:"refresh_token,omitempty"`
	Expiry       time.Time                                                                                 `json:"expiry,omitempty"`
	RefreshFunc  func(context.Context, httpcli.Doer, *OAuthBearerToken) (string, string, time.Time, error) `json:"-"`
	// Number of minutes before expiry when token should be refreshed.
	NeedsRefreshBuffer int `json:"-"`
}

func (token *OAuthBearerToken) Refresh(ctx context.Context, cli httpcli.Doer) error {
	if token.RefreshToken == "" {
		return errors.New("no refresh token available")
	}

	if token.RefreshFunc == nil {
		return errors.New("refresh not implemented")
	}

	newToken, newRefreshToken, newExpiry, err := token.RefreshFunc(ctx, cli, token)
	if err != nil {
		return err
	}

	token.Token = newToken
	token.Expiry = newExpiry
	token.RefreshToken = newRefreshToken

	return nil
}

func (token *OAuthBearerToken) NeedsRefresh() bool {
	// If there is no refresh token, always return false since we can't refresh
	if token.RefreshToken == "" {
		return false
	}
	// Refresh if the current time falls within the buffer period to expiry, and is not zero
	return !token.Expiry.IsZero() && (time.Until(token.Expiry) <= time.Duration(token.NeedsRefreshBuffer)*time.Minute)
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

// WithToken returns an Oauth Bearer Token authenticator with a new token
func (token *OAuthBearerToken) WithToken(newToken string) *OAuthBearerToken {
	return &OAuthBearerToken{
		Token: newToken,
	}
}

// SetURLUser authenticates the provided URL by setting the User field.
func (token *OAuthBearerToken) SetURLUser(u *url.URL) {
	u.User = url.UserPassword("oauth2", token.Token)
}

// OAuthBearerTokenWithSSH implements OAuth Bearer Token authentication for extsvc
// clients and holds an additional RSA keypair.
type OAuthBearerTokenWithSSH struct {
	OAuthBearerToken

	PrivateKey string
	PublicKey  string
	Passphrase string
}

var (
	_ Authenticator        = &OAuthBearerTokenWithSSH{}
	_ AuthenticatorWithSSH = &OAuthBearerTokenWithSSH{}
)

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
