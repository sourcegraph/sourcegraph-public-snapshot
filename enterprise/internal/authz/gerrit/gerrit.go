package gerrit

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type Provider struct {
	urn      string
	codeHost *extsvc.CodeHost
}

func NewProvider(urn, host string) *Provider {
	baseURL, _ := url.Parse(host)
	return &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypeGerrit),
	}
}

func (p Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account, verifiedEmails []string) (mine *extsvc.Account, err error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchAccount"}
}

func (p Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPerms"}
}

func (p Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchRepoPerms"}
}

func (p Provider) FetchUserPermsByToken(ctx context.Context, token string, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "gerrit.FetchUserPermsByToken"}
}

func (p Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p Provider) URN() string {
	return p.urn
}

func (p Provider) ValidateConnection(ctx context.Context) (warnings []string) {
	return nil
}
