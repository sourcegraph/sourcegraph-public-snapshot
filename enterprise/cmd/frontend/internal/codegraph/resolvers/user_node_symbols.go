package resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"
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
	do := func(ctx context.Context, myEmail string) ([]codeintelresolvers.AdjustedMonikerLocations, error) {
		// Get a list of all symbols authored by this person.
		var authoredSymbols []codeintelresolvers.AdjustedMonikerLocations
		for _, repoName := range repoNames {
			repo, err := backend.Repos.GetByName(ctx, repoName)
			if err != nil {
				return nil, err
			}

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
					normalizeHunkAuthor(hunk)
					if hunk.Author.Email == myEmail {
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

	const cachePath = "/tmp/user-node-symbols.json"
	readCache := func() (result []codeintelresolvers.AdjustedMonikerLocations) {
		data, _ := ioutil.ReadFile(cachePath)
		if data == nil {
			return nil
		}
		json.Unmarshal(data, &result)
		return result
	}
	writeCache := func(result []codeintelresolvers.AdjustedMonikerLocations) {
		data, _ := json.Marshal(result)
		ioutil.WriteFile(cachePath, data, 0600)
	}
	if result := readCache(); result != nil {
		return result, nil
	}
	defer func() { writeCache(symbolsResult) }()

	symbolsOnce.Do(func() { symbolsResult, symbolsErr = do(ctx, myEmail) })
	return symbolsResult, symbolsErr
}

func (r *CodeGraphPersonNodeResolver) Symbols(ctx context.Context) ([]string, error) {
	authoredSymbols, err := r.symbols(ctx)
	if err != nil {
		return nil, err
	}

	// Find the person's most-used symbols.
	type monikerRefCount struct {
		symbol codeintelresolvers.AdjustedMonikerLocations
		count  int
	}
	symbolCounts := make([]monikerRefCount, len(authoredSymbols))
	for i, symbol := range authoredSymbols {
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

		const limit = 50 // TODO(sqs): un-hardcode
		locations, _, err := codeIntelResolver.References(ctx, loc.AdjustedRange.Start.Line, loc.AdjustedRange.Start.Character, limit, "")
		if err != nil {
			return nil, err
		}

		symbolCounts[i] = monikerRefCount{symbol, len(locations)}
	}

	sort.Slice(symbolCounts, func(i, j int) bool { return symbolCounts[i].count > symbolCounts[j].count })

	var result []string
	for _, sc := range symbolCounts {
		if sc.count == 0 {
			continue
		}
		result = append(result, fmt.Sprintf("%s (%d)", sc.symbol.Identifier, sc.count))
	}

	return result, nil
}
