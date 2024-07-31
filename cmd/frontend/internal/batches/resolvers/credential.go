package resolvers

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/githubapp"
	btypes "github.com/sourcegraph/sourcegraph/internal/batches/types"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/auth"
	ghauth "github.com/sourcegraph/sourcegraph/internal/github_apps/auth"
	ghastore "github.com/sourcegraph/sourcegraph/internal/github_apps/store"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const batchChangesCredentialIDKind = "BatchChangesCredential"

const (
	siteCredentialPrefix = "site"
	userCredentialPrefix = "user"
)

func marshalBatchChangesCredentialID(id int64, isSiteCredential bool) graphql.ID {
	var idStr string
	if isSiteCredential {
		idStr = fmt.Sprintf("%s:%d", siteCredentialPrefix, id)
	} else {
		idStr = fmt.Sprintf("%s:%d", userCredentialPrefix, id)
	}
	return relay.MarshalID(batchChangesCredentialIDKind, idStr)
}

func unmarshalBatchChangesCredentialID(id graphql.ID) (credentialID int64, isSiteCredential bool, err error) {
	var strID string
	if err := relay.UnmarshalSpec(id, &strID); err != nil {
		return credentialID, isSiteCredential, err
	}

	parts := strings.SplitN(strID, ":", 2)
	if len(parts) != 2 {
		return credentialID, isSiteCredential, errors.New("invalid id")
	}

	kind := parts[0]
	switch strings.ToLower(kind) {
	case siteCredentialPrefix:
		isSiteCredential = true
	case userCredentialPrefix:
	default:
		return credentialID, isSiteCredential, errors.Errorf("invalid id, unsupported credential kind %q", kind)
	}

	parsedID, err := strconv.Atoi(parts[1])
	return int64(parsedID), isSiteCredential, err
}

func commentSSHKey(ssh auth.AuthenticatorWithSSH) string {
	url := conf.ExternalURLParsed()
	if url != nil && url.Host != "" {
		return strings.TrimRight(ssh.SSHPublicKey(), "\n") + " Sourcegraph " + url.Host
	}
	return ssh.SSHPublicKey()
}

type batchChangesUserCredentialResolver struct {
	credential *database.UserCredential

	repo    *types.Repo
	ghStore ghastore.GitHubAppsStore

	db     database.DB
	logger log.Logger
}

var _ graphqlbackend.BatchChangesCredentialResolver = &batchChangesUserCredentialResolver{}

func (c *batchChangesUserCredentialResolver) ID() graphql.ID {
	return marshalBatchChangesCredentialID(c.credential.ID, false)
}

func (c *batchChangesUserCredentialResolver) ExternalServiceKind() string {
	return extsvc.TypeToKind(c.credential.ExternalServiceType)
}

func (c *batchChangesUserCredentialResolver) ExternalServiceURL() string {
	// This is usually the code host URL.
	return c.credential.ExternalServiceID
}

func (c *batchChangesUserCredentialResolver) SSHPublicKey(ctx context.Context) (*string, error) {
	a, err := c.credential.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{
		Repo:           c.repo,
		GitHubAppStore: c.ghStore,
	})
	if err != nil {
		return nil, errors.Wrap(err, "retrieving authenticator")
	}

	if ssh, ok := a.(auth.AuthenticatorWithSSH); ok {
		publicKey := commentSSHKey(ssh)
		return &publicKey, nil
	}
	return nil, nil
}

func (c *batchChangesUserCredentialResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: c.credential.CreatedAt}
}

func (c *batchChangesUserCredentialResolver) IsSiteCredential() bool {
	return false
}

func (c *batchChangesUserCredentialResolver) authenticator(ctx context.Context) (auth.Authenticator, error) {
	return c.credential.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{
		Repo:           c.repo,
		GitHubAppStore: c.ghStore,
	})
}

func (c *batchChangesUserCredentialResolver) IsGitHubApp() bool { return c.credential.GitHubAppID > 0 }

func (c *batchChangesUserCredentialResolver) GitHubAppID() int {
	return c.credential.GitHubAppID
}

type batchChangesSiteCredentialResolver struct {
	credential *btypes.SiteCredential

	repo    *types.Repo
	ghStore ghastore.GitHubAppsStore

	db     database.DB
	logger log.Logger
}

func (c *batchChangesUserCredentialResolver) GitHubApp(ctx context.Context) (graphqlbackend.GitHubAppResolver, error) {
	if !c.IsGitHubApp() {
		return nil, nil
	}
	switch c.credential.ExternalServiceType {
	case extsvc.TypeGitHub:
		ghapp, err := c.ghStore.GetByID(ctx, c.GitHubAppID())
		if err != nil {
			if _, ok := err.(ghastore.ErrNoGitHubAppFound); ok {
				return nil, nil
			} else {
				return nil, err
			}
		}
		return githubapp.NewGitHubAppResolver(c.db, ghapp, c.logger), nil
	}
	return nil, nil
}

var _ graphqlbackend.BatchChangesCredentialResolver = &batchChangesSiteCredentialResolver{}

func (c *batchChangesSiteCredentialResolver) ID() graphql.ID {
	return marshalBatchChangesCredentialID(c.credential.ID, true)
}

func (c *batchChangesSiteCredentialResolver) ExternalServiceKind() string {
	return extsvc.TypeToKind(c.credential.ExternalServiceType)
}

func (c *batchChangesSiteCredentialResolver) ExternalServiceURL() string {
	// This is usually the code host URL.
	return c.credential.ExternalServiceID
}

func (c *batchChangesSiteCredentialResolver) SSHPublicKey(ctx context.Context) (*string, error) {
	a, err := c.credential.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{
		Repo:           c.repo,
		GitHubAppStore: c.ghStore,
	})
	if err != nil {
		return nil, errors.Wrap(err, "decrypting authenticator")
	}

	if ssh, ok := a.(auth.AuthenticatorWithSSH); ok {
		publicKey := commentSSHKey(ssh)
		return &publicKey, nil
	}
	return nil, nil
}

func (c *batchChangesSiteCredentialResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: c.credential.CreatedAt}
}

func (c *batchChangesSiteCredentialResolver) IsSiteCredential() bool {
	return true
}

func (c *batchChangesSiteCredentialResolver) authenticator(ctx context.Context) (auth.Authenticator, error) {
	return c.credential.Authenticator(ctx, ghauth.CreateAuthenticatorForCredentialOpts{
		Repo:           c.repo,
		GitHubAppStore: c.ghStore,
	})
}

func (c *batchChangesSiteCredentialResolver) IsGitHubApp() bool { return c.credential.GitHubAppID > 0 }

func (c *batchChangesSiteCredentialResolver) GitHubAppID() int { return c.credential.GitHubAppID }

func (c *batchChangesSiteCredentialResolver) GitHubApp(ctx context.Context) (graphqlbackend.GitHubAppResolver, error) {
	if !c.IsGitHubApp() {
		return nil, nil
	}
	switch c.credential.ExternalServiceType {
	case extsvc.TypeGitHub:
		ghapp, err := c.ghStore.GetByID(ctx, c.GitHubAppID())
		if err != nil {
			if _, ok := err.(ghastore.ErrNoGitHubAppFound); ok {
				return nil, nil
			} else {
				return nil, err
			}
		}
		return githubapp.NewGitHubAppResolver(c.db, ghapp, c.logger), nil
	}
	return nil, nil
}
