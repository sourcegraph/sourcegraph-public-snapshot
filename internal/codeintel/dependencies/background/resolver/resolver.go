package resolver

import (
	"context"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type resolver struct {
	dependenciesSvc *dependencies.Service
}

var _ goroutine.Handler = &resolver{}
var _ goroutine.ErrorHandler = &resolver{}

func (r *resolver) Handle(ctx context.Context) error {
	repoRevs, err := r.dependenciesSvc.SelectRepoRevisionsToResolve(ctx)
	if err != nil {
		return errors.Wrap(err, "dependencies.SelectRepoRevisionsToResolve")
	}

	resolved := map[string]map[string]string{}
	for repoName, commits := range repoRevs {
		resolved[repoName] = map[string]string{}

		for _, commit := range commits {
			resolved[repoName][commit] = "deadd00d" // TODO - actually resole via gitsvc
		}
	}

	if err := r.dependenciesSvc.UpdateResolvedRevisions(ctx, resolved); err != nil {
		return errors.Wrap(err, "dependencies.UpdateResolvedRevisions")
	}

	// TODO
	return nil
}

func (r *resolver) HandleError(err error) {
	// TODO
	fmt.Printf("OH NOOOOO %v\n", err)
}
