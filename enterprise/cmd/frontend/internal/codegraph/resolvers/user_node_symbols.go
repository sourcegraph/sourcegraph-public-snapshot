package resolvers

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var (
	symbolsOnce   sync.Once
	symbolsResult []codeintelresolvers.AdjustedMonikerLocations
	symbolsErr    error
)

// TODO(sqs): could also use ctags symbols then do find-references on those.
func (r *CodeGraphPersonNodeResolver) symbols(ctx context.Context) ([]codeintelresolvers.AdjustedMonikerLocations, error) {
	do := func(ctx context.Context, email string) ([]codeintelresolvers.AdjustedMonikerLocations, error) {
		// TODO(sqs): un-hardcode
		repos := []*types.Repo{
			{ID: 1, Name: "github.com/sourcegraph/sourcegraph"},
			{ID: 2, Name: "github.com/hashicorp/go-multierror"},
			{ID: 3, Name: "github.com/hashicorp/errwrap"},
		}

		// Get a list of all symbols authored by this person.
		var authoredSymbols []codeintelresolvers.AdjustedMonikerLocations
		for _, repo := range repos {
			commitID, err := backend.Repos.ResolveRev(ctx, repo, "HEAD")
			if err != nil {
				return nil, err
			}

			codeIntelResolver, err := r.resolver.codeIntelResolver().QueryResolver(ctx, &graphqlbackend.GitBlobLSIFDataArgs{
				Repo:   repo,
				Commit: commitID,
				Path:   "/",
			})
			if err != nil {
				return nil, err
			}
			if codeIntelResolver == nil {
				continue // no LSIF data
			}

			symbols, err := codeIntelResolver.Symbols(ctx)
			if err != nil {
				return nil, err
			}

			// Find which symbols were authored by this person.
			for _, symbol := range symbols {
				// TODO(sqs): assume 1st location is the definition
				loc := symbol.Locations[0]
				hunks, err := git.BlameFile(ctx, repo.Name, loc.Path, &git.BlameOptions{
					// TODO(sqs): should be loc.Dump.Commit or loc.AdjustedCommit?
					NewestCommit: api.CommitID(loc.AdjustedCommit),
					StartLine:    loc.AdjustedRange.Start.Line + 1,
					EndLine:      loc.AdjustedRange.End.Line + 1,
				})
				if err != nil {
					return nil, err
				}
				var isPersonAuthor bool
				for _, hunk := range hunks {
					if strings.Contains(hunk.Author.Email, email) {
						isPersonAuthor = true
						break
					}
				}
				if isPersonAuthor {
					authoredSymbols = append(authoredSymbols, symbol)
				}
			}
		}

		return authoredSymbols, nil
	}

	symbolsOnce.Do(func() { symbolsResult, symbolsErr = do(ctx, "@sourcegraph.com") })
	return symbolsResult, symbolsErr
}

func (r *CodeGraphPersonNodeResolver) Symbols(ctx context.Context) ([]string, error) {
	authoredSymbols, err := r.symbols(ctx)
	if err != nil {
		return nil, err
	}

	var result []string
	for _, s := range authoredSymbols {
		result = append(result, s.Identifier)
	}

	return result, nil
}
