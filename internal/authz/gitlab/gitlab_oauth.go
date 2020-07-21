// Package gitlab contains an authorization provider for GitLab that uses GitLab OAuth
// authenetication.
package gitlab

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

var _ authz.Provider = (*OAuthProvider)(nil)

type OAuthProvider struct {
	// The token is the access token used for syncing repositories from the code host,
	// but it may or may not be a sudo-scoped.
	token string

	urn            string
	clientProvider *gitlab.ClientProvider
	clientURL      *url.URL
	codeHost       *extsvc.CodeHost
}

type OAuthProviderOp struct {
	// The unique resource identifier of the external service where the provider is defined.
	URN string

	// BaseURL is the URL of the GitLab instance.
	BaseURL *url.URL

	// Token is an access token with api scope, it may or may not have sudo scope.
	//
	// 🚨 SECURITY: This value contains secret information that must not be shown to non-site-admins.
	Token string
}

func newOAuthProvider(op OAuthProviderOp, cli httpcli.Doer) *OAuthProvider {
	return &OAuthProvider{
		token: op.Token,

		urn:            op.URN,
		clientProvider: gitlab.NewClientProvider(op.BaseURL, cli),
		clientURL:      op.BaseURL,
		codeHost:       extsvc.NewCodeHost(op.BaseURL, extsvc.TypeGitLab),
	}
}

func (p *OAuthProvider) Validate() (problems []string) {
	return nil
}

func (p *OAuthProvider) URN() string {
	return p.urn
}

func (p *OAuthProvider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *OAuthProvider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *OAuthProvider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account) (mine *extsvc.Account, err error) {
	return nil, nil
}
