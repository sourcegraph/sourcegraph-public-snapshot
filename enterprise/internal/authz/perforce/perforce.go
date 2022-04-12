package perforce

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	jsoniter "github.com/json-iterator/go"
	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ authz.Provider = (*Provider)(nil)

// Provider implements authz.Provider for Perforce depot permissions.
type Provider struct {
	urn      string
	codeHost *extsvc.CodeHost
	depots   []extsvc.RepoID

	host     string
	user     string
	password string

	p4Execer p4Execer

	// NOTE: We do not need mutex because there is no concurrent access to these
	// 	fields in the current implementation.
	cachedAllUserEmails map[string]string   // username <-> email
	cachedGroupMembers  map[string][]string // group <-> members
}

type p4Execer interface {
	P4Exec(ctx context.Context, host, user, password string, args ...string) (io.ReadCloser, http.Header, error)
}

// NewProvider returns a new Perforce authorization provider that uses the given
// host, user and password to talk to a Perforce Server that is the source of
// truth for permissions. It assumes emails of Sourcegraph accounts match 1-1
// with emails of Perforce Server users. It uses our default gitserver client.
func NewProvider(urn, host, user, password string, depots []extsvc.RepoID, db database.DB) *Provider {
	baseURL, _ := url.Parse(host)
	return &Provider{
		urn:                urn,
		codeHost:           extsvc.NewCodeHost(baseURL, extsvc.TypePerforce),
		depots:             depots,
		host:               host,
		user:               user,
		password:           password,
		p4Execer:           gitserver.NewClient(db),
		cachedGroupMembers: make(map[string][]string),
	}
}

// FetchAccount uses given user's verified emails to match users on the Perforce
// Server. It returns when any of the verified email has matched and the match
// result is not deterministic.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User, _ []*extsvc.Account, verifiedEmails []string) (_ *extsvc.Account, err error) {
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

	emailSet := make(map[string]struct{}, len(verifiedEmails))
	for _, email := range verifiedEmails {
		emailSet[email] = struct{}{}
	}

	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.password, "users")
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}
	defer func() { _ = rc.Close() }()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		// e.g. alice <alice@example.com> (Alice) accessed 2020/12/04
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		username := fields[0]                  // e.g. alice
		email := strings.Trim(fields[1], "<>") // e.g. alice@example.com

		if _, ok := emailSet[email]; ok {
			accountData, err := jsoniter.Marshal(
				perforce.AccountData{
					Username: username,
					Email:    email,
				},
			)
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
	_, _ = io.Copy(io.Discard, rc)
	return nil, nil
}

// FetchUserPerms returns a list of depot prefixes that the given user has
// access to on the Perforce Server.
func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, errors.Errorf("not a code host of the account: want %q but have %q",
			account.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	user, err := perforce.GetExternalAccountData(&account.AccountData)
	if err != nil {
		return nil, errors.Wrap(err, "getting external account data")
	} else if user == nil {
		return nil, errors.New("no user found in the external account data")
	}

	// -u User : Displays protection lines that apply to the named user. This option
	// requires super access.
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.password, "protects", "-u", user.Username)
	if err != nil {
		return nil, errors.Wrap(err, "list ACLs by user")
	}
	defer func() { _ = rc.Close() }()

	// Pull permissions from protects file.
	perms := &authz.ExternalUserPermissions{}
	if len(p.depots) == 0 {
		err = errors.Wrap(scanProtects(rc, repoIncludesExcludesScanner(perms)), "repoIncludesExcludesScanner")
	} else {
		// SubRepoPermissions-enabled code path
		perms.SubRepoPermissions = make(map[extsvc.RepoID]*authz.SubRepoPermissions, len(p.depots))
		err = errors.Wrap(scanProtects(rc, fullRepoPermsScanner(perms, p.depots)), "fullRepoPermsScanner")
	}

	// As per interface definition for this method, implementation should return
	// partial but valid results even when something went wrong.
	return perms, errors.Wrap(err, "FetchUserPerms")
}

// FetchUserPermsByToken is the same as FetchUserPerms, but it only requires a
// token.
func (p *Provider) FetchUserPermsByToken(ctx context.Context, token string, opts authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	return nil, &authz.ErrUnimplemented{Feature: "perforce.FetchUserPermsByToken"}
}

// getAllUserEmails returns a set of username <-> email pairs of all users in the Perforce server.
func (p *Provider) getAllUserEmails(ctx context.Context) (map[string]string, error) {
	if p.cachedAllUserEmails != nil {
		return p.cachedAllUserEmails, nil
	}

	userEmails := make(map[string]string)
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.password, "users")
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}
	defer func() { _ = rc.Close() }()

	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		// e.g. alice <alice@example.com> (Alice) accessed 2020/12/04
		fields := strings.Fields(scanner.Text())
		if len(fields) < 2 {
			continue
		}
		username := fields[0]                  // e.g. alice
		email := strings.Trim(fields[1], "<>") // e.g. alice@example.com

		userEmails[username] = email
	}
	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanner.Err")
	}

	p.cachedAllUserEmails = userEmails
	return p.cachedAllUserEmails, nil
}

// getAllUsers returns a list of usernames of all users in the Perforce server.
func (p *Provider) getAllUsers(ctx context.Context) ([]string, error) {
	userEmails, err := p.getAllUserEmails(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get all user emails")
	}

	users := make([]string, 0, len(userEmails))
	for name := range userEmails {
		users = append(users, name)
	}
	return users, nil
}

// getGroupMembers returns all members of the given group in the Perforce server.
func (p *Provider) getGroupMembers(ctx context.Context, group string) ([]string, error) {
	if p.cachedGroupMembers[group] != nil {
		return p.cachedGroupMembers[group], nil
	}

	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.password, "group", "-o", group)
	if err != nil {
		return nil, errors.Wrap(err, "list group members")
	}
	defer func() { _ = rc.Close() }()

	var members []string
	startScan := false
	scanner := bufio.NewScanner(rc)
	for scanner.Scan() {
		line := scanner.Text()

		// Only start scan when we encounter the "Users:" line
		if !startScan {
			if strings.HasPrefix(line, "Users:") {
				startScan = true
			}
			continue
		}

		// Lines for users always start with a tab "\t"
		if !strings.HasPrefix(line, "\t") {
			break
		}

		members = append(members, strings.TrimSpace(line))
	}
	if err = scanner.Err(); err != nil {
		return nil, errors.Wrap(err, "scanner.Err")
	}

	// Drain remaining body
	_, _ = io.Copy(io.Discard, rc)

	p.cachedGroupMembers[group] = members
	return p.cachedGroupMembers[group], nil
}

// FetchRepoPerms returns a list of users that have access to the given
// repository on the Perforce Server.
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, opts authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
	if repo == nil {
		return nil, errors.New("no repository provided")
	} else if !extsvc.IsHostOfRepo(p.codeHost, &repo.ExternalRepoSpec) {
		return nil, errors.Errorf("not a code host of the repository: want %q but have %q",
			repo.ServiceID, p.codeHost.ServiceID)
	}

	// Disable FetchRepoPerms until we implement sub-repo permissions for it.
	if len(p.depots) > 0 {
		return nil, &authz.ErrUnimplemented{Feature: "perforce.FetchRepoPerms for sub-repo permissions"}
	}

	// -a : Displays protection lines for all users. This option requires super
	// access.
	rc, _, err := p.p4Execer.P4Exec(ctx, p.host, p.user, p.password, "protects", "-a", repo.ID)
	if err != nil {
		return nil, errors.Wrap(err, "list ACLs by depot")
	}
	defer func() { _ = rc.Close() }()

	users := make(map[string]struct{})
	if err := scanProtects(rc, allUsersScanner(ctx, p, users)); err != nil {
		return nil, errors.Wrap(err, "scanning protects")
	}

	userEmails, err := p.getAllUserEmails(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get all user emails")
	}
	extIDs := make([]extsvc.AccountID, 0, len(users))
	for user := range users {
		email, ok := userEmails[user]
		if !ok {
			continue
		}
		extIDs = append(extIDs, extsvc.AccountID(email))
	}
	return extIDs, nil
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

func (p *Provider) ValidateConnection(ctx context.Context) (problems []string) {
	// Validate the user has "super" access with "-u" option, see https://www.perforce.com/perforce/r12.1/manuals/cmdref/protects.html
	rc, _, err := p.p4Execer.P4Exec(context.Background(), p.host, p.user, p.password, "protects", "-u", p.user)
	if err == nil {
		_ = rc.Close()
		return nil
	}

	if strings.Contains(err.Error(), "You don't have permission for this operation.") {
		return []string{"the user does not have super access"}
	}
	return []string{"validate user access level: " + err.Error()}
}
