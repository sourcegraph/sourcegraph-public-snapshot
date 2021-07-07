package resolvers

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	codeintelresolvers "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var (
	dependenciesOnce   sync.Once
	dependenciesResult []codeintelresolvers.AdjustedLocation
	dependenciesErr    error
)

func (r *CodeGraphPersonNodeResolver) dependencies(ctx context.Context) ([]codeintelresolvers.AdjustedLocation, error) {
	do := func(ctx context.Context) ([]codeintelresolvers.AdjustedLocation, error) {
		// Find all calls authored by this person.
		var allRefs []codeintelresolvers.AdjustedLocation
		for _, repoName := range repoNames {
			repo, err := backend.Repos.GetByName(ctx, repoName)
			if err != nil {
				return nil, err
			}

			commitID, err := backend.Repos.ResolveRev(ctx, repo, "HEAD")
			if err != nil {
				return nil, err
			}

			files, err := git.LsFiles(ctx, repo.Name, commitID)
			if err != nil {
				return nil, err
			}

			for _, path := range files {
				if !strings.HasSuffix(path, ".go") {
					continue
				}

				codeIntelResolver, err := r.resolver.codeIntelResolver().QueryResolver(ctx, &graphqlbackend.GitBlobLSIFDataArgs{
					Repo:   repo,
					Commit: commitID,
					Path:   path,
				})
				if err != nil {
					return nil, err
				}
				if codeIntelResolver == nil {
					continue
				}

				hunks, err := git.BlameFile(ctx, repo.Name, path, &git.BlameOptions{NewestCommit: api.CommitID(commitID)})
				if err != nil {
					return nil, err
				}
				log.Printf("Blamed file %q %q, hunks: %d", repo.Name, path, len(hunks))

				for _, hunk := range hunks {
					normalizeHunkAuthor(hunk)
					if hunk.Author.Email != myEmail {
						continue
					}

					// TODO(sqs): does this get xrefs?
					// TODO(sqs): git hunk to codeintel range off-by-1 conversion?
					ranges, err := codeIntelResolver.Ranges(ctx, hunk.StartLine-1, hunk.EndLine-1)
					if err != nil {
						return nil, err
					}

					for i, rng := range ranges {
						// HACK skip tiny ranges
						if rng.Range.Start.Line == rng.Range.End.Line && rng.Range.End.Character-rng.Range.Start.Character < 5 {
							continue
						}
						if rng.HoverText == "" {
							continue
						}
						if i > 500 {
							continue
						}

						// TODO(sqs): slow
						// Call references again to get external references as well.
						locs, _, err := codeIntelResolver.References(ctx, rng.Range.Start.Line, rng.Range.Start.Character, 123, "")
						if err != nil {
							return nil, err
						}
						for _, loc := range locs {
							if loc.Dump.RepositoryID != int(repo.ID) { // xrefs only
								allRefs = append(allRefs, locs...)
							}
						}
					}
				}

			}
		}

		return allRefs, nil
	}

	const cachePath = "/tmp/user-node-dependencies.json"
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
	defer func() { writeCache(dependenciesResult) }()

	dependenciesOnce.Do(func() { dependenciesResult, dependenciesErr = do(ctx) })
	return dependenciesResult, dependenciesErr
}

func (r *CodeGraphPersonNodeResolver) Dependencies(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	dependencies, err := r.dependencies(ctx)
	if err != nil {
		return nil, err
	}

	repos := map[string] /* repo name */ *graphqlbackend.RepositoryResolver{}
	for _, d := range dependencies {
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

func (r *CodeGraphPersonNodeResolver) Authors(ctx context.Context) ([]*graphqlbackend.PersonResolver, error) {
	dependencies, err := r.dependencies(ctx)
	if err != nil {
		return nil, err
	}

	personSet := map[string] /* email */ *graphqlbackend.PersonResolver{}

	// Blame reference locations to find callers.
	if max := 3000; len(dependencies) > max {
		// dependencies = dependencies[:max]
	}
	// x := 100
	log.Printf("dependencies: %d", dependencies)
	for i, dependency := range dependencies {
		// if !strings.Contains(dependency.Dump.RepositoryName, "blackfriday") {
		// 	continue
		// }
		if i%17 != 0 {
			continue
		}

		hunks, err := git.BlameFile(ctx, api.RepoName(dependency.Dump.RepositoryName), dependency.Path, &git.BlameOptions{
			NewestCommit: api.CommitID(dependency.AdjustedCommit),
			StartLine:    dependency.AdjustedRange.Start.Line + 1,
			EndLine:      dependency.AdjustedRange.End.Line + 1,
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
