package own

import (
	"bytes"
	"context"
	"os"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Service gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type Service interface {
	// RulesetForRepo returns a CODEOWNERS file ruleset from a given repository at given commit ID.
	// In the case the file cannot be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	RulesetForRepo(context.Context, api.RepoName, api.CommitID) (*codeowners.Ruleset, error)

	// ResolveOwnersWithType takes a list of codeownerspb.Owner and attempts to retrieve more information about the
	// owner from the users and teams databases.
	ResolveOwnersWithType(context.Context, []*codeownerspb.Owner, OwnerResolutionContext) (map[OwnerKey]codeowners.ResolvedOwner, error)
}

func NewOwnerKey(handle, email string, resCtx OwnerResolutionContext) OwnerKey {
	return OwnerKey{
		resCtx,
		handle,
		email,
	}
}

type OwnerResolutionContext struct {
	RepoName api.RepoName
	RepoID   api.RepoID
}

var _ Service = &service{}

func NewService(g gitserver.Client, db database.DB) Service {
	return &service{
		gitserverClient: g,
		db:              db,
		userStore:       db.Users(),
		teamStore:       db.Teams(),
		// TODO: Potentially long living struct, we don't do cache invalidation here.
		// Might want to remove caching here and do that externally.
		// ownerCache: make(map[OwnerKey]codeowners.ResolvedOwner),
		fileCache: make(map[fileCacheKey]*codeowners.Ruleset),
	}
}

type service struct {
	gitserverClient gitserver.Client
	userStore       database.UserStore
	teamStore       database.TeamStore
	db              database.DB

	// mu         sync.RWMutex
	// ownerCache map[OwnerKey]codeowners.ResolvedOwner

	// TODO: Move outside of this service.
	mu        sync.RWMutex
	fileCache map[fileCacheKey]*codeowners.Ruleset
}

type fileCacheKey struct {
	repoName api.RepoName
	commitID api.CommitID
}

type OwnerKey struct {
	OwnerResolutionContext
	handle, email string
}

// codeownersLocations contains the locations where CODEOWNERS file
// is expected to be found relative to the repository root directory.
// These are in line with GitHub and GitLab documentation.
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
var codeownersLocations = []string{
	".github/test.CODEOWNERS", // hardcoded test file for internal dogfooding, first for priority.

	"CODEOWNERS",
	".github/CODEOWNERS",
	".gitlab/CODEOWNERS",
	"docs/CODEOWNERS",
}

// RulesetForRepo makes a best effort attempt to return a CODEOWNERS file ruleset
// from one of the possible codeownersLocations. It returns nil if no match is found.
func (s *service) RulesetForRepo(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (*codeowners.Ruleset, error) {
	key := fileCacheKey{repoName, commitID}
	s.mu.RLock()
	file, ok := s.fileCache[key]
	s.mu.RUnlock()
	if ok {
		return file, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	file, ok = s.fileCache[key]
	if ok {
		return file, nil
	}

	for _, path := range codeownersLocations {
		content, err := s.gitserverClient.ReadFile(
			ctx,
			authz.DefaultSubRepoPermsChecker,
			repoName,
			commitID,
			path,
		)
		if content != nil && err == nil {
			file, err = codeowners.Parse(bytes.NewReader(content))
			// Warn: retry loop if err, not persisted.
			if err != nil {
				return nil, err
			}
			s.fileCache[key] = file
			return file, nil
		} else if os.IsNotExist(err) {
			continue
		}
		return nil, err
	}
	s.fileCache[key] = nil
	return nil, nil
}

func (s *service) ResolveOwnersWithType(ctx context.Context, protoOwners []*codeownerspb.Owner, resCtx OwnerResolutionContext) (map[OwnerKey]codeowners.ResolvedOwner, error) {
	resolved := make(codeowners.ResolvedOwners, len(protoOwners))
	ret := make(map[OwnerKey]codeowners.ResolvedOwner)

	var contextStr string
	var serviceType string
	var repo *types.Repo
	var err error
	var eas []*extsvc.Account
	var p providers.Provider

	if resCtx.RepoID != 0 || resCtx.RepoName != "" {
		if resCtx.RepoID != 0 {
			repo, err = s.db.Repos().Get(ctx, resCtx.RepoID)
			if err != nil {
				return nil, err
			}
		} else {
			repo, err = s.db.Repos().GetByName(ctx, resCtx.RepoName)
			if err != nil {
				return nil, err
			}
		}
		eas, err = s.db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
			ServiceType:    repo.ExternalRepo.ServiceType,
			ServiceID:      repo.ExternalRepo.ServiceID,
			ExcludeExpired: true,
		})
		if err != nil {
			return nil, err
		}
		// TODO: How does this distinguish github and github enterprise?
		p = providers.GetProviderbyServiceType(repo.ExternalRepo.ServiceType)

		serviceType = repo.ExternalRepo.ServiceType
		contextStr = repo.ExternalRepo.ServiceType + ":" + repo.ExternalRepo.ServiceID
	}

	uc := &usersCache{t: make(map[string]*types.User), db: s.db}
	tc := &teamsCache{t: make(map[string]*types.Team), db: s.db}

	// We have to look up owner by owner because of the branching conditions:
	// We first try to find a user given the owner information. If we cannot find a user, we try to match a team.
	// If all fails, we return an unknown owner type with the information we have from the proto.
	for _, po := range protoOwners {
		ownerIdentifier := OwnerKey{resCtx, po.Handle, po.Email}
		resolvedOwner, err := resolveWithContext(ctx, s.db, po.Handle, po.Email, contextStr, serviceType, repo, p, eas, tc, uc)
		if err != nil {
			return nil, err
		}
		dedup, _ := resolved.Add(resolvedOwner)
		// Store reference to deduplicated resolvedOwner.
		ret[ownerIdentifier] = dedup
	}

	return ret, nil
}

// TODO: Make this lazy to only happen on the first comparison, so we don't need to
// match 10000 users if only one rule matches anyways.
func resolveWithContext(
	ctx context.Context,
	db database.DB,
	handle, email string,
	contextStr string,
	serviceType string,
	repo *types.Repo,
	p providers.Provider,
	eas []*extsvc.Account,
	tc *teamsCache,
	uc *usersCache,
) (owner codeowners.ResolvedOwner, err error) {
	var user *types.User
	if email != "" {
		user, err = uc.GetByVerifiedEmail(ctx, email)
		if err != nil {
			return nil, err
		}
		// email cannot match team, so can return a person early.
		if user == nil {
			return &codeowners.Person{
				Handle:  handle,
				Context: contextStr,
			}, nil
		}
	} else if repo != nil {
		for _, ea := range eas {
			data, err := p.ExternalAccountInfo(ctx, *ea)
			if err != nil {
				return nil, err
			}
			if data.Login != nil && *data.Login == handle {
				user, err = db.Users().GetByID(ctx, ea.UserID)
				if err != nil {
					if errcode.IsNotFound(err) {
						continue
					}
					return nil, err
				}
			}
		}
	} else {
		user, err = uc.GetByUsername(ctx, handle)
		if err != nil {
			return nil, err
		}
	}

	if user == nil {
		team, err := tc.GetTeamByName(ctx, teamName(handle, serviceType))
		if err != nil {
			return nil, err
		}
		if team != nil {
			// Team it is!
			return &codeowners.Team{Team: team}, nil
		}
		return &codeowners.Person{
			Handle:  handle,
			Context: contextStr,
		}, nil
	}

	teamOwners := []*codeowners.Team{}
	teams, _, err := db.Teams().ListTeams(ctx, database.ListTeamsOpts{ForUserMember: user.ID})
	if err != nil {
		return nil, err
	}
	for _, team := range teams {
		teamOwners = append(teamOwners, &codeowners.Team{Handle: team.Name, Team: team})
	}

	return &codeowners.User{User: user, Teams: teamOwners}, nil
}

func teamName(handle string, serviceType string) string {
	switch serviceType {
	case "github":
		split := strings.SplitN(handle, "/", 2)
		if len(split) == 2 {
			return split[1]
		} else {
			return split[0]
		}
	default:
	}
	return handle
}

// TODO: Does this cache help?
type usersCache struct {
	t  map[string]*types.User
	db database.DB
}

func (c *usersCache) GetByUsername(ctx context.Context, name string) (*types.User, error) {
	if user, ok := c.t[name]; ok {
		return user, nil
	}

	user, err := c.db.Users().GetByUsername(ctx, name)
	if err != nil {
		if errcode.IsNotFound(err) {
			c.t[name] = nil
			return nil, nil
		}
		return nil, err
	}
	c.t[name] = user
	return user, nil
}

func (c *usersCache) GetByVerifiedEmail(ctx context.Context, email string) (*types.User, error) {
	if user, ok := c.t[email]; ok {
		return user, nil
	}

	user, err := c.db.Users().GetByVerifiedEmail(ctx, email)
	if err != nil {
		if errcode.IsNotFound(err) {
			c.t[email] = nil
			return nil, nil
		}
		return nil, err
	}
	c.t[email] = user
	return user, nil
}

type teamsCache struct {
	t  map[string]*types.Team
	db database.DB
}

func (c *teamsCache) GetTeamByName(ctx context.Context, name string) (*types.Team, error) {
	if team, ok := c.t[name]; ok {
		return team, nil
	}

	team, err := c.db.Teams().GetTeamByName(ctx, name)
	if err != nil {
		if errcode.IsNotFound(err) {
			c.t[name] = nil
			return nil, nil
		}
		return nil, err
	}
	c.t[name] = team
	return team, nil
}
