package perforce

import (
	"context"
	"net/url"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var _ authz.Provider = (*Provider)(nil)

// Provider implements authz.Provider for Perforce depot permissions.
type Provider struct {
	userEmailsStore UserEmailsStore

	urn      string
	codeHost *extsvc.CodeHost

	host     string
	user     string
	password string
}

// NewProvider returns a new Perforce authorization provider that uses the given
// host, user and password to talk to a Perforce Server that is the source of
// truth for permissions. It assumes emails of Sourcegraph accounts match 1-1
// with emails of Perforce Server users.
func NewProvider(userEmailsStore UserEmailsStore, urn, host, user, password string) *Provider {
	baseURL, _ := url.Parse(host)
	return &Provider{
		userEmailsStore: userEmailsStore,
		urn:             urn,
		codeHost:        extsvc.NewCodeHost(baseURL, extsvc.TypePerforce),
		host:            host,
		user:            user,
		password:        password,
	}
}

// TODO:
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account) (mine *extsvc.Account, err error) {
	if user == nil {
		return nil, nil
	}

	emails, err := p.userEmailsStore.ListByUser(ctx,
		database.UserEmailsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "list verified emails")
	}

	// TODO: Fetch by list all of user' verified emails and try to find match in output of "p4 users"
	_ = emails

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
	// TODO: Validate with "p4 protects -u <p.user>" to make sure this is a super user to fetch ACL
	panic("implement me")
}
