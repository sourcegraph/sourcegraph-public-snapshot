package resolvers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var (
	dependentsOnce   sync.Once
	dependentsResult []codeintelresolvers.AdjustedLocation
	dependentsErr    error
)

func (r *CodeGraphPersonNodeResolver) dependents(ctx context.Context) ([]codeintelresolvers.AdjustedLocation, error) {
	do := func(ctx context.Context) ([]codeintelresolvers.AdjustedLocation, error) {
		// Find all callers to this person's authored symbols.
		authoredSymbols, err := r.symbols(ctx)
		if err != nil {
			return nil, err
		}

		var allLocations []codeintelresolvers.AdjustedLocation
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
			locations, _, err := codeIntelResolver.References(ctx, loc.AdjustedRange.Start.Line, loc.AdjustedRange.Start.Character, limit, "")
			if err != nil {
				return nil, err
			}

			allLocations = append(allLocations, locations...)
		}

		return allLocations, nil
	}

	const cachePath = "/tmp/user-node-dependents.json"
	readCache := func() (result []codeintelresolvers.AdjustedLocation) {
		data, _ := ioutil.ReadFile(cachePath)
		if data == nil {
			return nil
		}
		json.Unmarshal(data, &result)
		return result
	}
	writeCache := func(result []codeintelresolvers.AdjustedLocation) {
		data, _ := json.Marshal(result)
		ioutil.WriteFile(cachePath, data, 0600)
	}
	if result := readCache(); result != nil {
		return result, nil
	}
	defer func() { writeCache(dependentsResult) }()

	dependentsOnce.Do(func() { dependentsResult, dependentsErr = do(ctx) })
	return dependentsResult, dependentsErr
}

func (r *CodeGraphPersonNodeResolver) Dependents(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	dependents, err := r.dependents(ctx)
	if err != nil {
		return nil, err
	}

	repos := map[string] /* repo name */ *graphqlbackend.RepositoryResolver{}
	for _, d := range dependents {
		if _, seen := repos[d.Dump.RepositoryName]; !seen {
			repos[d.Dump.RepositoryName] = graphqlbackend.NewRepositoryResolver(dbconn.Global, &types.Repo{ID: api.RepoID(d.Dump.RepositoryID), Name: api.RepoName(d.Dump.RepositoryName)})
		}
	}

	repoResolvers := make([]*graphqlbackend.RepositoryResolver, 0, len(repos))
	for _, repoResolver := range repos {
		repoResolvers = append(repoResolvers, repoResolver)
	}

	return repoResolvers, nil
}

func (r *CodeGraphPersonNodeResolver) Callers(ctx context.Context) ([]*graphqlbackend.PersonResolver, error) {
	dependents, err := r.dependents(ctx)
	if err != nil {
		return nil, err
	}

	personSet := map[string] /* email */ *graphqlbackend.PersonResolver{}

	// Blame reference locations to find callers.
	for _, dependent := range dependents {
		hunks, err := git.BlameFile(ctx, api.RepoName(dependent.Dump.RepositoryName), dependent.Path, &git.BlameOptions{
			NewestCommit: api.CommitID(dependent.AdjustedCommit),
			StartLine:    dependent.AdjustedRange.Start.Line + 1,
			EndLine:      dependent.AdjustedRange.End.Line + 1,
		})
		if err != nil {
			return nil, err
		}

		for _, hunk := range hunks {
			normalizeHunkAuthor(hunk)
			if _, seen := personSet[hunk.Author.Email]; !seen {
				personSet[hunk.Author.Email] = graphqlbackend.NewPersonResolver(dbconn.Global, hunk.Author.Name, hunk.Author.Email, true)
			}
		}
	}

	personResolvers := make([]*graphqlbackend.PersonResolver, 0, len(personSet))
	for _, personResolver := range personSet {
		personResolvers = append(personResolvers, personResolver)
	}
	return personResolvers, nil
}

// TODO(sqs): hack
func normalizeHunkAuthor(hunk *git.Hunk) {
	const me = "quinn@slack.org"
	if hunk.Author.Email == "sqs@sourcegraph.com" || hunk.Author.Email == "qslack@qslack.com" {
		hunk.Author.Email = me
	}
}
