package perforce

import (
	"context"
	"net/url"
	"strings"
	"sync"
	"time"

	jsoniter "github.com/json-iterator/go"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/perforce"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/protocol"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var _ authz.Provider = (*Provider)(nil)

const cacheTTL = time.Hour

// Provider implements authz.Provider for Perforce depot permissions.
type Provider struct {
	logger log.Logger
	db     database.DB

	urn      string
	codeHost *extsvc.CodeHost
	depots   []extsvc.RepoID

	host     string
	user     string
	password string

	gitserverClient gitserver.Client

	emailsCacheMutex      sync.RWMutex
	cachedAllUserEmails   map[string]string // username -> email
	emailsCacheLastUpdate time.Time

	groupsCacheMutex      sync.RWMutex
	cachedGroupMembers    map[string][]string // group -> members
	groupsCacheLastUpdate time.Time
	ignoreRulesWithHost   bool
}

func cacheIsUpToDate(lastUpdate time.Time) bool {
	return time.Since(lastUpdate) < cacheTTL
}

// NewProvider returns a new Perforce authorization provider that uses the given
// host, user and password to talk to a Perforce Server that is the source of
// truth for permissions. It assumes emails of Sourcegraph accounts match 1-1
// with emails of Perforce Server users.
func NewProvider(logger log.Logger, db database.DB, gitserverClient gitserver.Client, urn, host, user, password string, depots []extsvc.RepoID, ignoreRulesWithHost bool) *Provider {
	baseURL, _ := url.Parse(host)
	return &Provider{
		db:                  db,
		logger:              logger,
		urn:                 urn,
		codeHost:            extsvc.NewCodeHost(baseURL, extsvc.TypePerforce),
		depots:              depots,
		host:                host,
		user:                user,
		password:            password,
		gitserverClient:     gitserverClient,
		cachedGroupMembers:  make(map[string][]string),
		ignoreRulesWithHost: ignoreRulesWithHost,
	}
}

// FetchAccount uses given user's verified emails to match users on the Perforce
// Server. It returns when any of the verified email has matched and the match
// result is not deterministic.
func (p *Provider) FetchAccount(ctx context.Context, user *types.User) (_ *extsvc.Account, err error) {
	if user == nil {
		return nil, nil
	}

	tr, ctx := trace.New(ctx, "perforce.authz.provider.FetchAccount")
	defer func() {
		tr.SetAttributes(
			attribute.String("user.name", user.Username),
			attribute.Int("user.id", int(user.ID)))

		if err != nil {
			tr.SetError(err)
		}

		tr.End()
	}()

	userEmails, err := p.db.UserEmails().ListByUser(ctx,
		database.UserEmailsListOptions{
			UserID:       user.ID,
			OnlyVerified: true,
		})
	if err != nil {
		return nil, errors.Wrap(err, "could not fetch user emails")
	}

	emailSet := make(map[string]struct{}, len(userEmails))
	for _, email := range userEmails {
		emailSet[strings.ToLower(email.Email)] = struct{}{}
	}

	users, err := p.gitserverClient.PerforceUsers(ctx, protocol.PerforceConnectionDetails{
		P4Port:   p.host,
		P4User:   p.user,
		P4Passwd: p.password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}

	for _, p4User := range users {
		if p4User.Email == "" || p4User.Username == "" {
			continue
		}
		p4Email := strings.ToLower(p4User.Email)

		if _, ok := emailSet[p4Email]; ok {
			accountData, err := jsoniter.Marshal(
				perforce.AccountData{
					Username: p4User.Username,
					Email:    p4Email,
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
					AccountID:   p4Email,
				},
				AccountData: extsvc.AccountData{
					Data: extsvc.NewUnencryptedData(accountData),
				},
			}, nil
		}
	}

	return nil, nil
}

// FetchUserPerms returns a list of depot prefixes that the given user has
// access to on the Perforce Server.
func (p *Provider) FetchUserPerms(ctx context.Context, account *extsvc.Account, _ authz.FetchPermsOptions) (*authz.ExternalUserPermissions, error) {
	if account == nil {
		return nil, errors.New("no account provided")
	} else if !extsvc.IsHostOfAccount(p.codeHost, account) {
		return nil, errors.Errorf("not a code host of the account: want %q but have %q",
			account.AccountSpec.ServiceID, p.codeHost.ServiceID)
	}

	user, err := perforce.GetExternalAccountData(ctx, &account.AccountData)
	if err != nil {
		return nil, errors.Wrap(err, "getting external account data")
	} else if user == nil {
		return nil, errors.New("no user found in the external account data")
	}

	protects, err := p.gitserverClient.PerforceProtectsForUser(ctx, protocol.PerforceConnectionDetails{
		P4Port:   p.host,
		P4User:   p.user,
		P4Passwd: p.password,
	}, user.Username)
	if err != nil {
		return nil, errors.Wrap(err, "list ACLs by user")
	}

	// Pull permissions from protects file.
	perms := &authz.ExternalUserPermissions{}
	if len(p.depots) == 0 {
		err = errors.Wrap(scanProtects(p.logger, protects, repoIncludesExcludesScanner(perms), p.ignoreRulesWithHost), "repoIncludesExcludesScanner")
	} else {
		// SubRepoPermissions-enabled code path
		perms.SubRepoPermissions = make(map[extsvc.RepoID]*authz.SubRepoPermissionsWithIPs, len(p.depots))
		err = errors.Wrap(scanProtects(p.logger, protects, fullRepoPermsScanner(p.logger, perms, p.depots), p.ignoreRulesWithHost), "fullRepoPermsScanner")
	}

	// As per interface definition for this method, implementation should return
	// partial but valid results even when something went wrong.
	return perms, errors.Wrap(err, "FetchUserPerms")
}

// getAllUserEmails returns a set of username -> email pairs of all users in the Perforce server.
func (p *Provider) getAllUserEmails(ctx context.Context) (map[string]string, error) {
	if p.cachedAllUserEmails != nil && cacheIsUpToDate(p.emailsCacheLastUpdate) {
		return p.cachedAllUserEmails, nil
	}

	userEmails := make(map[string]string)
	users, err := p.gitserverClient.PerforceUsers(ctx, protocol.PerforceConnectionDetails{
		P4Port:   p.host,
		P4User:   p.user,
		P4Passwd: p.password,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list users")
	}

	for _, p4User := range users {
		if p4User.Username == "" || p4User.Email == "" {
			continue
		}
		userEmails[p4User.Username] = p4User.Email
	}

	p.emailsCacheMutex.Lock()
	defer p.emailsCacheMutex.Unlock()
	p.cachedAllUserEmails = userEmails
	p.emailsCacheLastUpdate = time.Now()

	return p.cachedAllUserEmails, nil
}

// getAllUsers returns a list of usernames of all users in the Perforce server.
func (p *Provider) getAllUsers(ctx context.Context) ([]string, error) {
	userEmails, err := p.getAllUserEmails(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get all user emails")
	}

	// We lock here since userEmails above is a reference to the cached emails
	p.emailsCacheMutex.RLock()
	defer p.emailsCacheMutex.RUnlock()
	users := make([]string, 0, len(userEmails))
	for name := range userEmails {
		users = append(users, name)
	}
	return users, nil
}

// getGroupMembers returns all members of the given group in the Perforce server.
func (p *Provider) getGroupMembers(ctx context.Context, group string) ([]string, error) {
	if p.cachedGroupMembers[group] != nil && cacheIsUpToDate(p.groupsCacheLastUpdate) {
		return p.cachedGroupMembers[group], nil
	}

	p.groupsCacheMutex.Lock()
	defer p.groupsCacheMutex.Unlock()

	members, err := p.gitserverClient.PerforceGroupMembers(
		ctx,
		protocol.PerforceConnectionDetails{
			P4Port:   p.host,
			P4User:   p.user,
			P4Passwd: p.password,
		},
		group,
	)
	if err != nil {
		return nil, errors.Wrap(err, "list group members")
	}

	p.cachedGroupMembers[group] = members
	p.groupsCacheLastUpdate = time.Now()
	return p.cachedGroupMembers[group], nil
}

// excludeGroupMembers excludes members of a given group from provided users map
func (p *Provider) excludeGroupMembers(ctx context.Context, group string, users map[string]struct{}) error {
	members, err := p.getGroupMembers(ctx, group)
	if err != nil {
		return errors.Wrapf(err, "list members of group %q", group)
	}

	p.groupsCacheMutex.RLock()
	defer p.groupsCacheMutex.RUnlock()

	for _, member := range members {
		delete(users, member)
	}
	return nil
}

// includeGroupMembers includes members of a given group to provided users map
func (p *Provider) includeGroupMembers(ctx context.Context, group string, users map[string]struct{}) error {
	members, err := p.getGroupMembers(ctx, group)
	if err != nil {
		return errors.Wrapf(err, "list members of group %q", group)
	}

	p.groupsCacheMutex.RLock()
	defer p.groupsCacheMutex.RUnlock()

	for _, member := range members {
		users[member] = struct{}{}
	}
	return nil
}

// FetchRepoPerms returns a list of users that have access to the given
// repository on the Perforce Server.
func (p *Provider) FetchRepoPerms(ctx context.Context, repo *extsvc.Repository, _ authz.FetchPermsOptions) ([]extsvc.AccountID, error) {
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

	protects, err := p.gitserverClient.PerforceProtectsForDepot(
		ctx,
		protocol.PerforceConnectionDetails{
			P4Port:   p.host,
			P4User:   p.user,
			P4Passwd: p.password,
		},
		repo.ID,
	)
	if err != nil {
		return nil, errors.Wrap(err, "list ACLs by depot")
	}

	users := make(map[string]struct{})
	if err := scanProtects(p.logger, protects, allUsersScanner(ctx, p, users), p.ignoreRulesWithHost); err != nil {
		return nil, errors.Wrap(err, "scanning protects")
	}

	userEmails, err := p.getAllUserEmails(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "get all user emails")
	}
	extIDs := make([]extsvc.AccountID, 0, len(users))

	// We lock here since userEmails above is a reference to the cached emails
	p.emailsCacheMutex.RLock()
	defer p.emailsCacheMutex.RUnlock()
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

func (p *Provider) ValidateConnection(ctx context.Context) error {
	return p.gitserverClient.IsPerforceSuperUser(ctx, protocol.PerforceConnectionDetails{
		P4Port:   p.host,
		P4User:   p.user,
		P4Passwd: p.password,
	})
}
