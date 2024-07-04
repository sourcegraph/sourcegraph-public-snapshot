package auth

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/github"
	ghstore "github.com/sourcegraph/sourcegraph/internal/github_apps/store"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CreateAuthenticatorForCredentialOpts struct {
	Repo           *types.Repo
	GitHubAppStore ghstore.GitHubAppsStore
}

func CreateAuthenticatorForCredential(ctx context.Context, ghAppID int, opts CreateAuthenticatorForCredentialOpts) (auth.Authenticator, error) {
	var authenticator auth.Authenticator

	ghApp, err := opts.GitHubAppStore.GetByID(ctx, ghAppID)
	if err != nil {
		return nil, err
	}

	authenticator, err = NewGitHubAppAuthenticator(ghApp.AppID, []byte(ghApp.PrivateKey))
	if err != nil {
		return nil, err
	}

	if opts.Repo != nil {
		baseURL, err := url.Parse(opts.Repo.ExternalRepo.ServiceID)
		if err != nil {
			return nil, err
		}

		md, ok := opts.Repo.Metadata.(*github.Repository)
		if !ok {
			return nil, errors.Newf("expected repo metadata to be a github.Repository, but got %T", opts.Repo.Metadata)
		}

		owner, _, err := github.SplitRepositoryNameWithOwner(md.NameWithOwner)
		if err != nil {
			return nil, err
		}
		installID, err := opts.GitHubAppStore.GetInstallID(ctx, ghApp.AppID, owner)
		if err != nil {
			return nil, err
		}
		authenticator = NewInstallationAccessToken(baseURL, installID, authenticator, keyring.Default().GitHubAppKey)
	}
	return authenticator, nil
}
