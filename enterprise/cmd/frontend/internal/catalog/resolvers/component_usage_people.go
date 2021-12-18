package resolvers

import (
	"context"
	"sort"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type blameLocation struct {
	repoName                                         api.RepoName
	commit                                           api.CommitID
	path                                             string
	startLine, startCharacter, endLine, endCharacter int
}

func (r *componentUsageResolver) People(ctx context.Context) ([]gql.ComponentUsedByPersonEdgeResolver, error) {
	results, err := r.cachedResults(ctx)
	if err != nil {
		return nil, err
	}

	var locs []blameLocation
	for _, res := range results.Results() {
		if fm, ok := res.ToFileMatch(); ok {
			for _, m := range fm.LineMatches() {
				locs = append(locs, blameLocation{
					repoName:       fm.RepoName().Name,
					commit:         fm.CommitID,
					path:           fm.Path,
					startLine:      int(m.LineNumber()),
					startCharacter: int(m.OffsetAndLengths()[0][0]),
					endLine:        int(m.LineNumber()),
					endCharacter:   int(m.OffsetAndLengths()[0][0] + m.OffsetAndLengths()[0][1]),
				})
			}
		}
	}

	type blameAuthorWithLocation struct {
		blameAuthor
		locations []blameLocation
	}
	all := map[string]*blameAuthorWithLocation{}
	for _, loc := range locs {
		authorsByEmail, _, err := getBlameAuthors(ctx, loc.repoName, loc.path, git.BlameOptions{
			NewestCommit: loc.commit,
			StartLine:    loc.startLine,
			EndLine:      loc.endLine,
		})
		if err != nil {
			return nil, err
		}
		for email, a := range authorsByEmail {
			ca := all[email]
			if ca == nil {
				all[email] = &blameAuthorWithLocation{
					blameAuthor: *a,
					locations:   []blameLocation{loc},
				}
			} else {
				ca.locations = append(ca.locations, loc)
				ca.LineCount += a.LineCount
				if a.LastCommitDate.After(ca.LastCommitDate) {
					ca.Name = a.Name // use latest name in case it changed over time
					ca.LastCommit = a.LastCommit
					ca.LastCommitDate = a.LastCommitDate
				}
			}
		}
	}

	edges := make([]gql.ComponentUsedByPersonEdgeResolver, 0, len(all))
	for _, a := range all {
		edges = append(edges, &componentUsedByPersonEdgeResolver{
			db:        r.db,
			component: r.component,
			data:      &a.blameAuthor,
			locations: a.locations,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		return edges[i].AuthoredLineCount() > edges[j].AuthoredLineCount()
	})

	return edges, nil
}

type componentUsedByPersonEdgeResolver struct {
	db        database.DB
	component *componentResolver
	data      *blameAuthor
	locations []blameLocation
}

func (r *componentUsedByPersonEdgeResolver) Node() *gql.PersonResolver {
	return gql.NewPersonResolver(r.db, r.data.Name, r.data.Email, true)
}

func (r *componentUsedByPersonEdgeResolver) Locations(ctx context.Context) (gql.LocationConnectionResolver, error) {
	var locationResolvers []gql.LocationResolver
	for _, loc := range r.locations {
		// TODO(sqs): SECURITY does this bypass repo perms?
		repoResolver := gql.NewRepositoryResolver(r.db, &types.Repo{Name: loc.repoName})
		commitResolver := gql.NewGitCommitResolver(r.db, repoResolver, loc.commit, nil)
		file := gql.NewGitTreeEntryResolver(r.db, commitResolver, gql.CreateFileInfo(loc.path, false)) // loc.path
		locationResolvers = append(locationResolvers, gql.NewLocationResolver(file, &lsp.Range{
			Start: lsp.Position{Line: loc.startLine, Character: 0},
			End:   lsp.Position{Line: loc.endLine, Character: 0},
		}))
	}
	return gql.NewStaticLocationConnectionResolver(locationResolvers, false), nil
}

func (r *componentUsedByPersonEdgeResolver) AuthoredLineCount() int32 {
	return int32(r.data.LineCount)
}

func (r *componentUsedByPersonEdgeResolver) LastCommit(ctx context.Context) (*gql.GitCommitResolver, error) {
	// TODO(sqs): assumes usage site is in same repo as component, which is not generally true
	repoResolver, err := r.component.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}

	return gql.NewGitCommitResolver(r.db, repoResolver, r.data.LastCommit, nil), nil
}
