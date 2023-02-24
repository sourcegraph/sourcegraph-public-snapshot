package backend

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
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// OwnService gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type OwnService interface {
	// OwnersFile returns a CODEOWNERS file from a given repository at given commit ID.
	// In the case the file cannot be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	OwnersFile(context.Context, api.RepoName, api.CommitID) (*codeownerspb.File, error)

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

var _ OwnService = &ownService{}

func NewOwnService(g gitserver.Client, db database.DB) OwnService {
	return &ownService{
		gitserverClient: g,
		db:              db,
		userStore:       db.Users(),
		teamStore:       db.Teams(),
		// TODO: Potentially long living struct, we don't do cache invalidation here.
		// Might want to remove caching here and do that externally.
		ownerCache: make(map[OwnerKey]codeowners.ResolvedOwner),
	}
}

type ownService struct {
	gitserverClient gitserver.Client
	userStore       database.UserStore
	teamStore       database.TeamStore
	db              database.DB

	mu         sync.RWMutex
	ownerCache map[OwnerKey]codeowners.ResolvedOwner
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

// OwnersFile makes a best effort attempt to return a CODEOWNERS file from one of
// the possible codeownersLocations. It returns nil if no match is found.
func (s *ownService) OwnersFile(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (*codeownerspb.File, error) {
	for _, path := range codeownersLocations {
		content, err := s.gitserverClient.ReadFile(
			ctx,
			authz.DefaultSubRepoPermsChecker,
			repoName,
			commitID,
			path,
		)
		if content != nil && err == nil {
			return codeowners.Parse(bytes.NewReader(content))
		} else if os.IsNotExist(err) {
			continue
		}
		return nil, err
	}
	return nil, nil
}

func (s *ownService) ResolveOwnersWithType(ctx context.Context, protoOwners []*codeownerspb.Owner, resCtx OwnerResolutionContext) (map[OwnerKey]codeowners.ResolvedOwner, error) {
	resolved := make(codeowners.ResolvedOwners, len(protoOwners))
	ret := make(map[OwnerKey]codeowners.ResolvedOwner)

	// We have to look up owner by owner because of the branching conditions:
	// We first try to find a user given the owner information. If we cannot find a user, we try to match a team.
	// If all fails, we return an unknown owner type with the information we have from the proto.
	for _, po := range protoOwners {
		ownerIdentifier := OwnerKey{resCtx, po.Handle, po.Email}
		s.mu.RLock()
		cached, ok := s.ownerCache[ownerIdentifier]
		s.mu.RUnlock()
		if ok {
			resolved.Add(cached)
			continue
		}

		resolvedOwner, err := resolveWithContext(ctx, s.db, po.Handle, po.Email, resCtx)
		if err != nil {
			return nil, err
		}
		dedup, _ := resolved.Add(resolvedOwner)
		// Store reference to deduplicated resolvedOwner.
		ret[ownerIdentifier] = dedup
		s.mu.Lock()
		s.ownerCache[ownerIdentifier] = dedup
		s.mu.Unlock()
	}

	return ret, nil
}

// TODO: Make this lazy to only happen on the first comparison, so we don't need to
// match 10000 users if only one rule matches anyways.
func resolveWithContext(ctx context.Context, db database.DB, handle, email string, resCtx OwnerResolutionContext) (owner codeowners.ResolvedOwner, err error) {
	var contextStr string
	var serviceType string
	var user *types.User
	if email != "" {
		user, err = db.Users().GetByVerifiedEmail(ctx, email)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
	} else if resCtx.RepoID != 0 || resCtx.RepoName != "" {
		var r *types.Repo
		if resCtx.RepoID != 0 {
			r, err = db.Repos().Get(ctx, resCtx.RepoID)
			if err != nil {
				return nil, err
			}
		} else {
			r, err = db.Repos().GetByName(ctx, resCtx.RepoName)
			if err != nil {
				return nil, err
			}
		}
		eas, err := db.UserExternalAccounts().List(ctx, database.ExternalAccountsListOptions{
			ServiceType:    r.ExternalRepo.ServiceType,
			ServiceID:      r.ExternalRepo.ServiceID,
			ExcludeExpired: true,
		})
		if err != nil {
			return nil, err
		}

		// TODO: How does this distinguish github and github enterprise?
		p := providers.GetProviderbyServiceType(r.ExternalRepo.ServiceType)
		serviceType = r.ExternalRepo.ServiceType
		contextStr = r.ExternalRepo.ServiceType + ":" + r.ExternalRepo.ServiceID
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
		user, err = db.Users().GetByUsername(ctx, handle)
		if err != nil && !errcode.IsNotFound(err) {
			return nil, err
		}
	}

	if user == nil {
		teamHandle := handle
		switch serviceType {
		case "github":
			split := strings.SplitN(teamHandle, "/", 2)
			if len(split) == 2 {
				teamHandle = split[1]
			} else {
				teamHandle = split[0]
			}
		default:
		}
		team, err := db.Teams().GetTeamByName(ctx, teamHandle)
		if err != nil {
			if errcode.IsNotFound(err) {
				return &codeowners.Person{
					Handle:  handle,
					Context: contextStr,
				}, nil
			}
			return nil, err
		}
		// Team it is!
		return &codeowners.Team{Team: team}, nil
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
