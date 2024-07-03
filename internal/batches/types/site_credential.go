package types

import (
	"context"
	"time"

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
func (sc *SiteCredential) IsGitHubApp() bool { return sc.GitHubAppID != 0 }

// Authenticator decrypts and creates the authenticator associated with the site credential.
func (sc *SiteCredential) Authenticator(ctx context.Context) (auth.Authenticator, error) {
	if sc.IsGitHubApp() {
		return sc.githubAppAuthenticator()
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

func (sc *SiteCredential) githubAppAuthenticator() (auth.Authenticator, error) {
	return nil, errors.New("not implemented")
	// var authenticator auth.Authenticator
	// ghApp, err := opts.GitHubAppStore.GetByID(ctx, uc.GitHubAppID)
	// if err != nil {
	// 	return nil, err
	// }

	// authenticator, err = ghappAuth.NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
	// if err != nil {
	// 	return nil, err
	// }

	// if opts.Repo != nil {
	// 	baseURL, err := url.Parse(opts.Repo.ExternalRepo.ServiceID)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	md, ok := opts.Repo.Metadata.(*github.Repository)
	// 	if !ok {
	// 		return nil, errors.Newf("expected repo metadata to be a github.Repository, but got %T", opts.Repo.Metadata)
	// 	}

	// 	owner, _, err := github.SplitRepositoryNameWithOwner(md.NameWithOwner)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	installID, err := opts.GitHubAppStore.GetInstallID(ctx, ghApp.AppID, owner)
	// 	fmt.Println("install id is", installID)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	authenticator = ghappAuth.NewInstallationAccessToken(baseURL, installID, authenticator, keyring.Default().GitHubAppKey)
	// }
	// return authenticator, nil
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
