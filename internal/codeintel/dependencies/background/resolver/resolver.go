package resolver

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitService interface {
	GetCommits(ctx context.Context, repoCommits []api.RepoCommit, ignoreErrors bool) ([]*gitdomain.Commit, error)
}

type resolver struct {
	dependenciesSvc *dependencies.Service
	gitSvc          GitService
}

var _ goroutine.Handler = &resolver{}
var _ goroutine.ErrorHandler = &resolver{}

func (r *resolver) Handle(ctx context.Context) error {
	repoRevs, err := r.dependenciesSvc.SelectRepoRevisionsToResolve(ctx)
	if err != nil {
		return errors.Wrap(err, "dependencies.SelectRepoRevisionsToResolve")
	}

	repoCommits := []api.RepoCommit{}
	for repoName, revSpecs := range repoRevs {
		for _, revSpec := range revSpecs {
			repoCommits = append(repoCommits, api.RepoCommit{
				Repo:     api.RepoName(repoName),
				CommitID: api.CommitID(revSpec),
			})
		}
	}

	resolvedCommits, err := r.gitSvc.GetCommits(ctx, repoCommits, true)
	if err != nil {
		return errors.Wrap(err, "gitservice.GetCommits")
	}

	resolved := map[string]map[string]string{}
	for i, commit := range repoCommits {
		resolvedCommit := resolvedCommits[i]
		if resolvedCommit == nil {
			// TODO: UpdateResolvedRevisions should accepted nil-values to
			// "unresolve" revisions
			continue
		}

		repoName := string(commit.Repo)
		revSpec := string(commit.CommitID)
		resolvedCommitID := string(resolvedCommit.ID)

		_, ok := resolved[repoName]
		if !ok {
			resolved[repoName] = map[string]string{}
		}

		resolved[repoName][revSpec] = resolvedCommitID
	}

	if err := r.dependenciesSvc.UpdateResolvedRevisions(ctx, resolved); err != nil {
		return errors.Wrap(err, "dependencies.UpdateResolvedRevisions")
	}

	return nil
}

func (r *resolver) HandleError(err error) {
	// TODO
	fmt.Printf("OH NOOOOO %v\n", err)
}
