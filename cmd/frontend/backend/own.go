package backend

import (
	"bytes"
	"context"
	"os"
	"sync"

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
	ResolveOwnersWithType(context.Context, []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error)
}

var _ OwnService = &ownService{}

func NewOwnService(g gitserver.Client, db database.DB) OwnService {
	return &ownService{
		gitserverClient: g,
		userStore:       db.Users(),
		teamStore:       db.Teams(),
		ownerCache:      make(map[ownerKey]codeowners.ResolvedOwner),
	}
}

type ownService struct {
	gitserverClient gitserver.Client
	userStore       database.UserStore
	teamStore       database.TeamStore

	mu         sync.Mutex
	ownerCache map[ownerKey]codeowners.ResolvedOwner
}

type ownerKey struct {
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

func (s *ownService) ResolveOwnersWithType(ctx context.Context, protoOwners []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error) {
	resolved := make([]codeowners.ResolvedOwner, 0, len(protoOwners))

	// We have to look up owner by owner because of the branching conditions:
	// We first try to find a user given the owner information. If we cannot find a user, we try to match a team.
	// If all fails, we return an unknown owner type with the information we have from the proto.
	for _, po := range protoOwners {
		ownerIdentifier := ownerKey{po.Handle, po.Email}
		s.mu.Lock()
		cached, ok := s.ownerCache[ownerIdentifier]
		s.mu.Unlock()
		if ok {
			resolved = append(resolved, cached)
			continue
		}

		resolvedOwner, err := s.resolveOwner(ctx, po.Handle, po.Email)
		if err != nil {
			return nil, err
		}
		if resolvedOwner == nil {
			// This is a safeguard in case somehow neither email nor handle are set.
			continue
		}
		resolved = append(resolved, resolvedOwner)
		s.mu.Lock()
		s.ownerCache[ownerIdentifier] = resolvedOwner
		s.mu.Unlock()
	}

	return resolved, nil
}

func (s *ownService) resolveOwner(ctx context.Context, handle, email string) (codeowners.ResolvedOwner, error) {
	if handle != "" {
		resolvedOwner, err := tryGetUserThenTeam(ctx, handle, s.userStore.GetByUsername, s.teamStore.GetTeamByName)
		if err != nil {
			return unknownOwnerOrError(handle, email, err)
		}
		return resolvedOwner, nil
	} else if email != "" {
		// Teams cannot be identified by emails, so we do not pass in a team getter here.
		resolvedOwner, err := tryGetUserThenTeam(ctx, email, s.userStore.GetByVerifiedEmail, nil)
		if err != nil {
			return unknownOwnerOrError(handle, email, err)
		}
		return resolvedOwner, nil
	}
	return nil, nil
}

type userGetterFunc func(context.Context, string) (*types.User, error)
type teamGetterFunc func(context.Context, string) (*types.Team, error)

func tryGetUserThenTeam(ctx context.Context, identifier string, userGetter userGetterFunc, teamGetter teamGetterFunc) (codeowners.ResolvedOwner, error) {
	user, err := userGetter(ctx, identifier)
	if err != nil {
		if errcode.IsNotFound(err) {
			if teamGetter != nil {
				team, err := teamGetter(ctx, identifier)
				if err != nil {
					return nil, err
				}
				return &codeowners.Team{Team: team, OwnerIdentifier: identifier}, nil
			}
		}
		return nil, err
	}
	return &codeowners.Person{User: user, OwnerIdentifier: identifier}, nil
}

func unknownOwnerOrError(handle, email string, err error) (*codeowners.UnknownOwner, error) {
	if errcode.IsNotFound(err) {
		return &codeowners.UnknownOwner{Handle: handle, Email: email}, nil
	}
	return nil, err
}
