package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *CodeGraphPersonNodeResolver) Dependents(ctx context.Context) ([]string, error) {
	// Find all callers to this person's authored symbols.
	authoredSymbols, err := r.symbols(ctx)
	if err != nil {
		return nil, err
	}

	var dependents []string
	for _, symbol := range authoredSymbols {
		loc := symbol.Locations[0]

		codeIntelResolver, err := r.resolver.codeIntelResolver().QueryResolver(ctx, &graphqlbackend.GitBlobLSIFDataArgs{
			Repo:   &types.Repo{ID: api.RepoID(symbol.Dump.RepositoryID), Name: api.RepoName(symbol.Dump.RepositoryName)},
			Commit: api.CommitID(symbol.Locations[0].AdjustedCommit),
			Path:   loc.Path,
		})
		if err != nil {
			return nil, err
		}
		if codeIntelResolver == nil {
			continue
		}

		// TODO(sqs): could simplify lookup and add a new codeIntelResolver.ReferencesByMoniker
		// instead of references-by-position API.
		const limit = 20 // TODO(sqs): un-hardcode
		refLocations, _, err := codeIntelResolver.References(ctx, loc.AdjustedRange.Start.Line, loc.AdjustedRange.Start.Character, limit, "")
		if err != nil {
			return nil, err
		}

		// Blame reference locations to find callers.
		for _, refLocation := range refLocations {
			hunks, err := git.BlameFile(ctx, api.RepoName(refLocation.Dump.RepositoryName), refLocation.Path, &git.BlameOptions{
				NewestCommit: api.CommitID(refLocation.AdjustedCommit),
				StartLine:    refLocation.AdjustedRange.Start.Line + 1,
				EndLine:      refLocation.AdjustedRange.End.Line + 1,
			})
			if err != nil {
				return nil, err
			}

			for _, hunk := range hunks {
				dependents = append(dependents, hunk.Author.Email)
			}
		}
	}

	return dependents, nil
}
