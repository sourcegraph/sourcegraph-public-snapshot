package types

import (
	"context"
	"time"

	ghauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SiteCredential struct {
	ID                  int64
	ExternalServiceType string
	ExternalServiceID   string
	CreatedAt           time.Time
	UpdatedAt           time.Time
	GitHubAppID         int

	Credential *database.EncryptableCredential
}

// IsGitHubApp returns true if the site credential is a GitHub App.
func (sc *SiteCredential) IsGitHubApp() bool { return sc.GitHubAppID > 0 }

// Authenticator decrypts and creates the authenticator associated with the site credential.
func (sc *SiteCredential) Authenticator(ctx context.Context, opts ghauth.CreateAuthenticatorForCredentialOpts) (auth.Authenticator, error) {
	if sc.IsGitHubApp() {
		return ghauth.CreateAuthenticatorForCredential(ctx, sc.GitHubAppID, opts)
	}

	decrypted, err := sc.Credential.Decrypt(ctx)
	if err != nil {
		return nil, err
	}

	a, err := database.UnmarshalAuthenticator(decrypted)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshalling authenticator")
	}

	return a, nil
}

// SetAuthenticator encrypts and sets the authenticator within the site credential.
func (sc *SiteCredential) SetAuthenticator(ctx context.Context, a auth.Authenticator) error {
	if sc.Credential == nil {
		sc.Credential = database.NewUnencryptedCredential(nil)
	}

	if sc.IsGitHubApp() {
		return nil
	}

	raw, err := database.MarshalAuthenticator(a)
	if err != nil {
		return err
	}

	sc.Credential = database.NewUnencryptedCredential([]byte(raw))
	return nil
}
