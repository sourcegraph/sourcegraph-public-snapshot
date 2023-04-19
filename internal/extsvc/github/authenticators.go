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
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubAppAuthenticator struct {
	appID  string
	key    *rsa.PrivateKey
	rawKey []byte
}

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

func (a *GitHubAppAuthenticator) generateJWT() (string, error) {
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
		Issuer:    a.appID,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	return token.SignedString(a.key)
}

func (a *GitHubAppAuthenticator) Hash() string {
	shaSum := sha256.Sum256(a.rawKey)
	return hex.EncodeToString(shaSum[:])
}

type GitHubAppInstallationAuthenticator struct {
	installationID          int
	InstallationAccessToken string
	Expiry                  time.Time
	appAuthenticator        *GitHubAppAuthenticator
}

func NewGitHubAppInstallationAuthenticator(
	logger log.Logger,
	installationID int,
	installationAccessToken string,
	appAuthenticator *GitHubAppAuthenticator,
) *GitHubAppInstallationAuthenticator {
	auther := &GitHubAppInstallationAuthenticator{
		installationID:          installationID,
		InstallationAccessToken: installationAccessToken,
		appAuthenticator:        appAuthenticator,
	}
	if installationAccessToken == "" {
		// TODO: auther.Refresh()
	}
	return auther
}

func (a *GitHubAppInstallationAuthenticator) Refresh(ctx context.Context, cli httpcli.Doer) error {
	_, err := a.appAuthenticator.generateJWT()
	if err != nil {
		return err
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

func (a *GitHubAppInstallationAuthenticator) Authenticate(r *http.Request) error {
	r.Header.Set("Authorization", "Bearer "+a.InstallationAccessToken)
	return nil
}

func (a *GitHubAppInstallationAuthenticator) Hash() string {
	sum := sha256.Sum256([]byte(strconv.Itoa(int(a.installationID))))
	return hex.EncodeToString(sum[:])
}

func (a *GitHubAppInstallationAuthenticator) NeedsRefresh() bool {
	if !a.Expiry.IsZero() {
		return time.Until(a.Expiry) < 5*time.Minute
	}
	return false
}
