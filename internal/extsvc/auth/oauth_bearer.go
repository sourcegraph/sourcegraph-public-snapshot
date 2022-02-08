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

// oauthBearerTokenWithGitHubApp implements OAuth Bearer Token authentication for
// GitHub Apps.
type oauthBearerTokenWithGitHubApp struct {
	appID  string
	key    *rsa.PrivateKey
	rawKey []byte
}

// NewOAuthBearerTokenWithGitHubApp constructs a new OAuth Bearer Token
// authenticator for GitHub Apps using given appID and private key.
func NewOAuthBearerTokenWithGitHubApp(appID string, privateKey []byte) (Authenticator, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "parse private key")
	}
	return &oauthBearerTokenWithGitHubApp{
		appID:  appID,
		key:    key,
		rawKey: privateKey,
	}, nil
}

// Authenticate is a modified version of
// https://github.com/bradleyfalzon/ghinstallation/blob/24e56b3fb7669f209134a01eff731d7e2ef72a5c/appsTransport.go#L66.
func (token *oauthBearerTokenWithGitHubApp) Authenticate(r *http.Request) error {
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

func (token *oauthBearerTokenWithGitHubApp) Hash() string {
	shaSum := sha256.Sum256(token.rawKey)
	return hex.EncodeToString(shaSum[:])
}
