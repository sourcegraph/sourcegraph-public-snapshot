package github

import (
	"context"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// GitHubAppAuthenticator is used to authenticate requests to the GitHub API
// using a GitHub App. It contains the ID and private key associated with
// the GitHub App.
type GitHubAppAuthenticator struct {
	appID  string
	key    *rsa.PrivateKey
	rawKey []byte
}

// NewGitHubAppAuthenticator creates an Authenticator that can be used to authenticate requests
// to the GitHub API as a GitHub App. It requires the GitHub App ID and RSA private key.
//
// The returned Authenticator can be used to sign requests to the GitHub API on behalf of the GitHub App.
// The requests will contain a JSON Web Token (JWT) in the Authorization header with claims identifying
// the GitHub App.
func NewGitHubAppAuthenticator(appID string, privateKey []byte) (auth.Authenticator, error) {
	key, err := jwt.ParseRSAPrivateKeyFromPEM(privateKey)
	if err != nil {
		return nil, errors.Wrap(err, "parse private key")
	}
	return &GitHubAppAuthenticator{
		appID:  appID,
		key:    key,
		rawKey: privateKey,
	}, nil
}

func (a *GitHubAppAuthenticator) Authenticate(r *http.Request) error {
	token, err := a.generateJWT()
	if err != nil {
		return err
	}
	r.Header.Set("Authorization", "Bearer "+token)
	return nil
}

// generateJWT generates a JSON Web Token (JWT) signed with the GitHub App's private key.
// The JWT contains claims identifying the GitHub App.
//
// The payload computation is following GitHub App's Ruby example shown in
// https://docs.github.com/en/developers/apps/building-github-apps/authenticating-with-github-apps#authenticating-as-a-github-app.
//
// NOTE: GitHub rejects expiry and issue timestamps that are not an integer,
// while the jwt-go library serializes to fractional timestamps. Truncate them
// before passing to jwt-go.
//
// The returned JWT can be used to authenticate requests to the GitHub API as the GitHub App.
func (a *GitHubAppAuthenticator) generateJWT() (string, error) {
	iss := time.Now().Add(-time.Minute).Truncate(time.Second)
	exp := iss.Add(10 * time.Minute)
	claims := &jwt.RegisteredClaims{
		IssuedAt:  jwt.NewNumericDate(iss),
		ExpiresAt: jwt.NewNumericDate(exp),
		Issuer:    a.appID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(a.key)
}

func (a *GitHubAppAuthenticator) Hash() string {
	shaSum := sha256.Sum256(a.rawKey)
	return hex.EncodeToString(shaSum[:])
}

// GitHubAppInstallationAuthenticator is used to authenticate requests to the GitHub API
// using an installation access token associated with a GitHub App installation.
// It contains the installation ID, installation access token, and expiry time.
// It also contains an appAuthenticator which is used to refresh the installation access token.
type GitHubAppInstallationAuthenticator struct {
	installationID          int
	InstallationAccessToken string
	Expiry                  time.Time
	appAuthenticator        auth.Authenticator
}

// NewGitHubAppInstallationAuthenticator creates an Authenticator that can be used to authenticate requests
// to the GitHub API using an installation access token associated with a GitHub App installation.
//
// The returned Authenticator can be used to authenticate requests to the GitHub API on behalf of the installation.
// The requests will contain the installation access token in the Authorization header.
// When the installation access token expires, the appAuthenticator will be used to generate a new one.
func NewGitHubAppInstallationAuthenticator(
	installationID int,
	installationAccessToken string,
	appAuthenticator auth.Authenticator,
) *GitHubAppInstallationAuthenticator {
	auther := &GitHubAppInstallationAuthenticator{
		installationID:          installationID,
		InstallationAccessToken: installationAccessToken,
		appAuthenticator:        appAuthenticator,
	}
	return auther
}

// Refresh generates a new installation access token for the GitHub App installation.
//
// It makes a request to the GitHub API to generate a new installation access token for the
// installation associated with the Authenticator. It updates the Authenticator with the new
// installation access token and expiry time.
//
// Returns an error if the request fails, or if there is no Authenticator to authenticate
// the token refresh request.
func (a *GitHubAppInstallationAuthenticator) Refresh(ctx context.Context, cli httpcli.Doer) error {
	if a.appAuthenticator == nil {
		return errors.New("appAuthenticator is nil")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("/app/installations/%d/access_tokens", a.installationID), nil)
	if err != nil {
		return err
	}
	a.appAuthenticator.Authenticate(req)

	resp, err := cli.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var accessToken struct {
		Token     string    `json:"token"`
		ExpiresAt time.Time `json:"expires_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&accessToken); err != nil {
		return err
	}
	a.InstallationAccessToken = accessToken.Token
	a.Expiry = accessToken.ExpiresAt
	return nil
}

// Authenticate adds an Authorization header to the request containing
// the installation access token associated with the GitHub App installation.
func (a *GitHubAppInstallationAuthenticator) Authenticate(r *http.Request) error {
	r.Header.Set("Authorization", "Bearer "+a.InstallationAccessToken)
	return nil
}

func (a *GitHubAppInstallationAuthenticator) Hash() string {
	sum := sha256.Sum256([]byte(strconv.Itoa(a.installationID)))
	return hex.EncodeToString(sum[:])
}

// NeedsRefresh checks if the installation access token associated with the
// GitHubAppInstallationAuthenticator needs to be refreshed.
//
// It returns true if the expiry time of the current installation access token
// is within 5 minutes, indicating a new access token should be requested.
// It returns false if the access token does not need to be refreshed yet.
func (a *GitHubAppInstallationAuthenticator) NeedsRefresh() bool {
	if !a.Expiry.IsZero() {
		return time.Until(a.Expiry) < 5*time.Minute
	}
	return false
}
