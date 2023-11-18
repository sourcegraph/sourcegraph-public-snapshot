package own

import (
	"bytes"
	"context"
	"os"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/own/codeowners"
)

// Service gives access to code ownership data.
// At this point only data from CODEOWNERS file is presented, if available.
type Service interface {
	// RulesetForRepo returns a CODEOWNERS file ruleset from a given repository at given commit ID.
	// If a CODEOWNERS file has been manually ingested for the repository, it will prioritise returning that file.
	// In the case the file cannot be found, `nil` `*codeownerspb.File` and `nil` `error` is returned.
	RulesetForRepo(context.Context, api.RepoName, api.RepoID, api.CommitID) (*codeowners.Ruleset, error)

	// AssignedOwnership returns the owners that were assigned for given repo within
	// Sourcegraph. This is an owners set that is independent of CODEOWNERS files.
	// Owners are assigned for repositories and directory hierarchies,
	// so an owner for the whole repo transitively owns all files in that repo,
	// and owner of 'src/test' in a given repo transitively owns all files within
	// the directory tree at that root like 'src/test/com/sourcegraph/Test.java'.
	AssignedOwnership(context.Context, api.RepoID, api.CommitID) (AssignedOwners, error)

	// AssignedTeams returns the teams that were assigned for given repo within
	// Sourcegraph. This is an owners set that is independent of CODEOWNERS files.
	// Teams are assigned for repositories and directory hierarchies, so an owner
	// team for the whole repo transitively owns all files in that repo, and owner
	// team of 'src/test' in a given repo transitively owns all files within the
	// directory tree at that root like 'src/test/com/sourcegraph/Test.java'.
	AssignedTeams(context.Context, api.RepoID, api.CommitID) (AssignedTeams, error)
}

type AssignedOwners map[string][]database.AssignedOwnerSummary

// Match returns all the assigned owner summaries for the given path.
// It implements inheritance of assigned ownership down the file tree,
// that is so that owners of a parent directory "a/b" are the owners
// of all files in that tree, like "a/b/c/d/foo.go".
func (ao AssignedOwners) Match(path string) []database.AssignedOwnerSummary {
	return match(ao, path)
}

type AssignedTeams map[string][]database.AssignedTeamSummary

// Match returns all the assigned team summaries for the given path.
// It implements inheritance of assigned ownership down the file tree,
// that is so that owners of a parent directory "a/b" are the owners
// of all files in that tree, like "a/b/c/d/foo.go".
func (at AssignedTeams) Match(path string) []database.AssignedTeamSummary {
	return match(at, path)
}

func match[T any](assigned map[string][]T, path string) []T {
	var summaries []T
	for lastSlash := len(path); lastSlash != -1; lastSlash = strings.LastIndex(path, "/") {
		path = path[:lastSlash]
		summaries = append(summaries, assigned[path]...)
	}
	if path != "" {
		summaries = append(summaries, assigned[""]...)
	}
	return summaries
}

var _ Service = &service{}

func NewService(g gitserver.Client, db database.DB) Service {
	return &service{
		gitserverClient: g,
		db:              db,
	}
}

type service struct {
	gitserverClient gitserver.Client
	db              database.DB
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
	var rs *codeowners.Ruleset
	if ingestedCodeowners != nil {
		rs = codeowners.NewRuleset(codeowners.IngestedRulesetSource{ID: int32(ingestedCodeowners.RepoID)}, ingestedCodeowners.Proto)
	} else {
		for _, path := range codeownersLocations {
			content, err := s.gitserverClient.ReadFile(
				ctx,
				repoName,
				commitID,
				path,
			)
			if content != nil && err == nil {
				pbfile, err := codeowners.Parse(bytes.NewReader(content))
				if err != nil {
					return nil, err
				}
				rs = codeowners.NewRuleset(codeowners.GitRulesetSource{Repo: repoID, Commit: commitID, Path: path}, pbfile)
				break
			} else if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
	}
	if rs == nil {
		return nil, nil
	}
	repo, err := s.db.Repos().Get(ctx, repoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	} else if errcode.IsNotFound(err) {
		return nil, nil
	}
	rs.SetCodeHostType(repo.ExternalRepo.ServiceType)
	return rs, nil
}

func (s *service) AssignedOwnership(ctx context.Context, repoID api.RepoID, _ api.CommitID) (AssignedOwners, error) {
	summaries, err := s.db.AssignedOwners().ListAssignedOwnersForRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}
	assignedOwners := AssignedOwners{}
	for _, summary := range summaries {
		byPath := assignedOwners[summary.FilePath]
		byPath = append(byPath, *summary)
		assignedOwners[summary.FilePath] = byPath
	}
	return assignedOwners, nil
}

func (s *service) AssignedTeams(ctx context.Context, repoID api.RepoID, _ api.CommitID) (AssignedTeams, error) {
	summaries, err := s.db.AssignedTeams().ListAssignedTeamsForRepo(ctx, repoID)
	if err != nil {
		return nil, err
	}
	assignedTeams := AssignedTeams{}
	for _, summary := range summaries {
		byPath := assignedTeams[summary.FilePath]
		byPath = append(byPath, *summary)
		assignedTeams[summary.FilePath] = byPath
	}
	return assignedTeams, nil
}
