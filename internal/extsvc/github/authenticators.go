package github

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// gitHubAppAuthenticator implements OAuth Bearer Token authentication for
// GitHub Apps.
type gitHubAppAuthenticator struct {
	appID  string
	key    *rsa.PrivateKey
	rawKey []byte
}

// NewGitHubAppAuthenticator constructs a new OAuth Bearer Token
// authenticator for GitHub Apps using given appID and private key.
func NewGitHubAppAuthenticator(appID string, privateKey []byte) (auth.Authenticator, error) {
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
	refreshFunc             func(context.Context, httpcli.Doer) (string, time.Time, error)
}

// NewGitHubAppAuthenticator constructs a new OAuth Bearer Token
// authenticator for GitHub Apps using given appID and private key.
func NewGitHubAppInstallationAuthenticator(installationID int64, installationAccessToken string, expiry time.Time, refreshFunc func(context.Context, httpcli.Doer) (string, time.Time, error)) (auth.AuthenticatorWithRefresh, error) {
	return &GitHubAppInstallationAuthenticator{
		installationID:          installationID,
		InstallationAccessToken: installationAccessToken,
		Expiry:                  expiry,
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

func (token *GitHubAppInstallationAuthenticator) NeedsRefresh() bool {
	if !token.Expiry.IsZero() {
		return time.Until(token.Expiry) < 5*time.Minute
	}

	// If no expiry is set we default to False
	return false
}

func (token *GitHubAppInstallationAuthenticator) Refresh(ctx context.Context, cli httpcli.Doer) error {
	newToken, newExpiry, err := token.refreshFunc(ctx, cli)
	if err != nil {
		return err
	}

	token.InstallationAccessToken = newToken
	token.Expiry = newExpiry

	return nil
}
