package perforce

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
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

// FetchAccount uses given user's verified emails to match users on the Perforce
// Server. It returns when any of the verified email has matched and the match
// result is not deterministic.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account) (_ *extsvc.Account, err error) {
	if user == nil {
		return nil, nil
	}

	tr, ctx := trace.New(ctx, "perforce.authz.provider.FetchAccount", "")
	defer func() {
		tr.LogFields(
			otlog.String("user.name", user.Username),
			otlog.Int32("user.id", user.ID),
		)

		if err != nil {
			tr.SetError(err)
		}

		tr.Finish()
	}()

	emails, err := p.userEmailsStore.ListByUser(ctx,
		database.UserEmailsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		},
	)
	if err != nil {
		return nil, errors.Wrap(err, "list verified emails")
	}

	emailSet := make(map[string]struct{}, len(emails))
	for _, email := range emails {
		emailSet[email.Email] = struct{}{}
	}

	rc, _, err := gitserver.DefaultClient.P4Exec(ctx, p.host, p.user, p.password, "users")
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}
	defer func() { _ = rc.Close() }()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		// e.g. alice <alice@example.com> (Alice) accessed 2020/12/04
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) < 2 {
			continue
		}
		username := fields[0]                  // e.g. alice
		email := strings.Trim(fields[1], "<>") // e.g. alice@example.com

		if _, ok := emailSet[email]; ok {
			accountData, err := jsoniter.Marshal(map[string]string{
				"username": username,
				"email":    email,
			})
			if err != nil {
				return nil, err
			}

			return &extsvc.Account{
				UserID: user.ID,
				AccountSpec: extsvc.AccountSpec{
					ServiceType: p.codeHost.ServiceType,
					ServiceID:   p.codeHost.ServiceID,
					AccountID:   email,
				},
				AccountData: extsvc.AccountData{
					Data: (*json.RawMessage)(&accountData),
				},
			}, nil
		}
	}
	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanner.Err")
	}

	// Drain remaining body
	_, _ = io.Copy(ioutil.Discard, rc)

	return nil, nil
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
