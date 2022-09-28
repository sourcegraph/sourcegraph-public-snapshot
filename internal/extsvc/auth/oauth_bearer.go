package auth

import (
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// OAuthBearerToken implements OAuth Bearer Token authentication for extsvc
// clients.
type OAuthBearerToken struct {
	AccessToken  string    `json:"access_token"`
	TokenType    string    `json:"token_type"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	RefreshFunc  func(*OAuthBearerToken) (*OAuthBearerToken, error)
}

var _ Authenticator = &OAuthBearerToken{}

func (token *OAuthBearerToken) Authenticate(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+token.AccessToken)
	return nil
}

func (token *OAuthBearerToken) Hash() string {
	key := sha256.Sum256([]byte(token.AccessToken))
	return hex.EncodeToString(key[:])
}

// WithToken returns an Oauth Bearer Token authenticator with a new token
func (token *OAuthBearerToken) WithToken(newToken string) *OAuthBearerToken {
	return &OAuthBearerToken{
		AccessToken: newToken,
	}
}

func (token *OAuthBearerToken) Refresh() error {
	if token.RefreshFunc == nil {
		return errors.New("refresh not implemented")
	}

	newToken, err := token.RefreshFunc(token)
	if err != nil {
		return err
	}

	token.AccessToken = newToken.AccessToken
	token.Expiry = newToken.Expiry
	token.RefreshToken = newToken.RefreshToken

	return nil
}

func (token *OAuthBearerToken) ShouldRefresh() bool {
	// Refresh 5 minutes before expiry
	return time.Until(token.Expiry) > 5*time.Minute
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
	shaSum := sha256.Sum256([]byte(token.AccessToken + token.PrivateKey + token.Passphrase + token.PublicKey))
	return hex.EncodeToString(shaSum[:])
}

// gitHubAppAuthenticator implements OAuth Bearer Token authentication for
// GitHub Apps.
type gitHubAppAuthenticator struct {
	appID  string
	key    *rsa.PrivateKey
	rawKey []byte
}

// NewGitHubAppAuthenticator constructs a new OAuth Bearer Token
// authenticator for GitHub Apps using given appID and private key.
func NewGitHubAppAuthenticator(appID string, privateKey []byte) (Authenticator, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "parse private key")
	}
	return &gitHubAppAuthenticator{
		appID:  appID,
		key:    key,
		rawKey: privateKey,
	}, nil
}

// Authenticate is a modified version of
// https://github.com/bradleyfalzon/ghinstallation/blob/24e56b3fb7669f209134a01eff731d7e2ef72a5c/appsTransport.go#L66.
func (token *gitHubAppAuthenticator) Authenticate(r *http.Request) error {
	// The payload computation is following GitHub App's Ruby example shown in
	// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app.
	//
	// NOTE: GitHub rejects expiry and issue timestamps that are not an integer,
	// while the jwt-go library serializes to fractional timestamps. Truncate them
	// before passing to jwt-go.
	iss := time.Now().Add(-time.Minute).Truncate(time.Second)
	exp := iss.Add(10 * time.Minute)
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(iss),
		ExpiresAt: jwt.NewNumericDate(exp),
		Issuer:    token.appID,
	}
	bearer := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedString, err := bearer.SignedString(token.key)
	if err != nil {
		return errors.Wrap(err, "sign JWT")
	}

	r.Header.Set("Authorization", "Bearer "+signedString)
	return nil
}

func (token *gitHubAppAuthenticator) Hash() string {
	shaSum := sha256.Sum256(token.rawKey)
	return hex.EncodeToString(shaSum[:])
}

// GitHubAppInstallationAuthenticator implements OAuth Bearer Token authentication for
// GitHub Apps.
type GitHubAppInstallationAuthenticator struct {
	installationID          int64
	InstallationAccessToken string
	Expiry                  time.Time
	refreshFunc             func(*GitHubAppInstallationAuthenticator) error
}

// NewGitHubAppAuthenticator constructs a new OAuth Bearer Token
// authenticator for GitHub Apps using given appID and private key.
func NewGitHubAppInstallationAuthenticator(installationID int64, installationAccessToken string, refreshFunc func(*GitHubAppInstallationAuthenticator) error) (Authenticator, error) {
	return &GitHubAppInstallationAuthenticator{
		installationID:          installationID,
		InstallationAccessToken: installationAccessToken,
		refreshFunc:             refreshFunc,
	}, nil
}

func (token *GitHubAppInstallationAuthenticator) Authenticate(r *http.Request) error {
	r.Header.Set("Authorization", "Bearer "+token.InstallationAccessToken)
	return nil
}

func (token *GitHubAppInstallationAuthenticator) Hash() string {
	shaSum := sha256.Sum256([]byte(token.InstallationAccessToken))
	return hex.EncodeToString(shaSum[:])
}

func (token *GitHubAppInstallationAuthenticator) ShouldRefresh() bool {
	return time.Until(token.Expiry) < 5*time.Minute
}

func (token *GitHubAppInstallationAuthenticator) Refresh() error {
	return token.refreshFunc(token)
}
