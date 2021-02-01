package perforce

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var _ authz.Provider = (*Provider)(nil)

// Provider implements authz.Provider for Perforce depot permissions.
type Provider struct {
	urn      string
	codeHost *extsvc.CodeHost

	host     string
	user     string
	password string
}

func NewProvider(urn, host, user, password string) *Provider {
	baseURL, _ := url.Parse(host)
	return &Provider{
		urn:      urn,
		codeHost: extsvc.NewCodeHost(baseURL, extsvc.TypePerforce),
		host:     host,
		user:     user,
		password: password,
	}
}

func (p *Provider) FetchAccount(ctx context.Context, user *types.User, current []*extsvc.Account) (mine *extsvc.Account, err error) {
	// TODO: Fetch by list all of user' verified emails and try to find match in output of "p4 users"
	panic("implement me")
}

func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account) ([]extsvc.RepoID, error) {
	// TODO: Fetch via "p4 protects -u alice"
	panic("implement me")
}

func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository) ([]extsvc.AccountID, error) {
	// TODO: Fetch via "p4 protects -a //depot-alice/", "p4 users", "p4 groups"
	panic("implement me")
}

func (p *Provider) ServiceType() string {
	return p.codeHost.ServiceType
}

func (p *Provider) ServiceID() string {
	return p.codeHost.ServiceID
}

func (p *Provider) URN() string {
	return p.urn
}

func (p *Provider) Validate() (problems []string) {
	// TODO: Validate with a ping
	panic("implement me")
}
