package backend

import (
	"bytes"
	"context"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// OwnService gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type OwnService interface {
	// OwnersFile returns a CODEOWNERS file from a given repository at given commit ID.
	// In the case the file cannot be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	OwnersFile(context.Context, api.RepoName, api.CommitID) (*codeownerspb.File, error)

	// todo maybe better name
	// ResolveOwnersWithType takes a list of codeownerspb.Owner and attempts to retrieve more information about the owner
	// from the users and teams databases.
	ResolveOwnersWithType(context.Context, []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error)
}

var _ OwnService = ownService{}

func NewOwnService(g gitserver.Client, db database.DB) OwnService {
	return ownService{
		gitserverClient: g,
		userStore:       db.Users(),
		ownerCache:      make(map[string]codeowners.ResolvedOwner),
	}
}

type ownService struct {
	gitserverClient gitserver.Client
	userStore       database.UserStore
	// todo add team store

	ownerCache map[string]codeowners.ResolvedOwner // handle/email -> ResolvedOwner
}

// codeownersLocations contains the locations where CODEOWNERS file
// is expected to be found relative to the repository root directory.
// These are in line with GitHub and GitLab documentation.
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
// https://docs.github.com/en/repositories/managing-your-repositorys-settings-and-features/customizing-your-repository/about-code-owners
var codeownersLocations = []string{
	"CODEOWNERS",
	".github/CODEOWNERS",
	".gitlab/CODEOWNERS",
	"docs/CODEOWNERS",
}

// OwnersFile makes a best effort attempt to return a CODEOWNERS file from one of
// the possible codeownersLocations. It returns nil if no match is found.
func (s ownService) OwnersFile(ctx context.Context, repoName api.RepoName, commitID api.CommitID) (*codeownerspb.File, error) {
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

func (s ownService) ResolveOwnersWithType(ctx context.Context, protoOwners []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error) {
	resolved := make([]codeowners.ResolvedOwner, 0, len(protoOwners))

	for _, po := range protoOwners {
		// an owner proto should have either handle or email set.
		ownerIdentifier := getHandleOrEmail(po)
		if cached, ok := s.ownerCache[ownerIdentifier]; ok {
			resolved = append(resolved, cached)
			continue
		}
		if ownerIdentifier == "" {
			// safeguard, maybe error
			continue
		}

		// we have to look up owner by owner because of the branching conditions.
		var resolvedOwner codeowners.ResolvedOwner
		if po.Handle != "" {
			user, err := s.userStore.GetByUsername(ctx, po.Handle)
			if err != nil {
				if errcode.IsNotFound(err) {
					// attempt team lookup
				} else {
					return nil, err
				}
			}
			resolvedOwner = codeowners.Person{Handle: po.Handle, User: user}
		} else if po.Email != "" {
			user, err := s.userStore.GetByVerifiedEmail(ctx, po.Email)
			if err != nil {
				if errcode.IsNotFound(err) {
					// attempt team lookup
				} else {
					return nil, err
				}
			}
			resolvedOwner = codeowners.Person{Email: po.Email, User: user}
		}
		resolved = append(resolved, resolvedOwner)
		s.ownerCache[ownerIdentifier] = resolvedOwner
	}
	return nil, nil
}

func NewPerson(user *types.User) codeowners.ResolvedOwner {
	return codeowners.Person{User: user}
}

func getHandleOrEmail(owner *codeownerspb.Owner) string {
	if owner.Handle != "" {
		return owner.Handle
	}
	return owner.Email
}
