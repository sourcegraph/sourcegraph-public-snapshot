package own

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners"
	codeownerspb "github.com/sourcegraph/sourcegraph/enterprise/internal/own/codeowners/v1"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// Service gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type Service interface {
	// RulesetForRepo returns a CODEOWNERS file ruleset from a given repository at given commit ID.
	// If a CODEOWNERS file has been manually ingested for the repository, it will prioritise returning that file.
	// In the case the file cannot be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	RulesetForRepo(context.Context, api.RepoName, api.RepoID, api.CommitID) (*codeowners.Ruleset, error)

	// ResolveOwnersWithType takes a list of codeownerspb.Owner and attempts to retrieve more information about the
	// owner from the users and teams databases.
	ResolveOwnersWithType(context.Context, []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error)
}

var _ Service = &service{}

func NewService(g gitserver.Client, db database.DB) Service {
	return &service{
		gitserverClient: g,
		db:              edb.NewEnterpriseDB(db),
		ownerCache:      make(map[ownerKey]codeowners.ResolvedOwner),
	}
}

type service struct {
	gitserverClient gitserver.Client
	db              edb.EnterpriseDB

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

// RulesetForRepo makes a best effort attempt to return a CODEOWNERS file ruleset
// from one of the possible codeownersLocations, or the ingested codeowners files. It returns nil if no match is found.
func (s *service) RulesetForRepo(ctx context.Context, repoName api.RepoName, repoID api.RepoID, commitID api.CommitID) (*codeowners.Ruleset, error) {
	ingestedCodeowners, err := s.db.Codeowners().GetCodeownersForRepo(ctx, repoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}
	if ingestedCodeowners != nil {
		return codeowners.NewRuleset(codeowners.IngestedRulesetSource{ID: int32(ingestedCodeowners.RepoID)}, ingestedCodeowners.Proto), nil
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
			pbfile, err := codeowners.Parse(bytes.NewReader(content))
			if err != nil {
				return nil, err
			}
			return codeowners.NewRuleset(codeowners.GitRulesetSource{Repo: repoID, Commit: commitID, Path: path}, pbfile), nil
		} else if os.IsNotExist(err) {
			continue
		}
		return nil, err
	}
	return nil, nil
}

func (s *service) ResolveOwnersWithType(ctx context.Context, protoOwners []*codeownerspb.Owner) ([]codeowners.ResolvedOwner, error) {
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

func (s *service) resolveOwner(ctx context.Context, handle, email string) (codeowners.ResolvedOwner, error) {
	var resolvedOwner codeowners.ResolvedOwner
	var err error
	if handle != "" {
		resolvedOwner, err = tryGetUserThenTeam(ctx, handle, s.db.Users().GetByUsername, s.whichTeamGetter(ctx))
		if err != nil {
			return personOrError(handle, email, err)
		}
	} else if email != "" {
		// Teams cannot be identified by emails, so we do not pass in a team getter here.
		resolvedOwner, err = tryGetUserThenTeam(ctx, email, s.db.Users().GetByVerifiedEmail, nil)
		if err != nil {
			return personOrError(handle, email, err)
		}
	} else {
		return nil, nil
	}
	resolvedOwner.SetOwnerData(handle, email)
	return resolvedOwner, nil
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
				return &codeowners.Team{Team: team}, nil
			}
		}
		return nil, err
	}
	return &codeowners.Person{User: user}, nil
}

func (s *service) whichTeamGetter(ctx context.Context) teamGetterFunc {
	fmt.Println("conf", conf.Get().OwnBestEffortTeamMatching)
	if conf.Get().OwnBestEffortTeamMatching != nil && !*conf.Get().OwnBestEffortTeamMatching {
		return s.db.Teams().GetTeamByName
	}
	return s.bestEffortTeamGetter
}

func (s *service) bestEffortTeamGetter(ctx context.Context, teamHandle string) (*types.Team, error) {
	team, err := s.db.Teams().GetTeamByName(ctx, teamHandle)
	if err == nil {
		return team, nil
	}
	if !errcode.IsNotFound(err) {
		return nil, err
	}
	// If the team handle is potentially embedded we will do best-effort matching on the last part of the team handle.
	if strings.Contains(teamHandle, "/") {
		return s.db.Teams().GetTeamByName(ctx, getLastPartOfTeamHandle(teamHandle))
	}
	return nil, err
}

func getLastPartOfTeamHandle(teamHandle string) string {
	// invariant: teamHandle contains a /.
	if len(teamHandle) == 1 {
		return teamHandle
	}
	lastSlashPos := strings.LastIndex(teamHandle, "/")
	return teamHandle[lastSlashPos+1:]
}

func personOrError(handle, email string, err error) (*codeowners.Person, error) {
	if errcode.IsNotFound(err) {
		return &codeowners.Person{Handle: handle, Email: email}, nil
	}
	return nil, err
}
