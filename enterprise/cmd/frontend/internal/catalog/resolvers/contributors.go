package resolvers

import (
	"context"
	"sort"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func (r *componentResolver) Contributors(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.ContributorConnectionResolver, error) {
	slocs, err := r.sourceLocationSetResolver(ctx)
	if err != nil {
		return nil, err
	}
	return slocs.Contributors(ctx, args)
}

func (r *rootResolver) GitTreeEntryContributors(ctx context.Context, treeEntry *gql.GitTreeEntryResolver, args *graphqlutil.ConnectionArgs) (gql.ContributorConnectionResolver, error) {
	return sourceLocationSetResolverFromTreeEntry(treeEntry, r.db).Contributors(ctx, args)
}

func (r *sourceLocationSetResolver) Contributors(ctx context.Context, args *graphqlutil.ConnectionArgs) (gql.ContributorConnectionResolver, error) {
	allFiles, err := r.allFiles(ctx)
	if err != nil {
		return nil, err
	}

	var (
		mu             sync.Mutex
		all            = map[string]*blameAuthor{}
		totalLineCount int
		allErr         error
		wg             sync.WaitGroup
	)
	for _, f := range allFiles {
		if f.IsDir() {
			continue
		}

		wg.Add(1)
		go func(f fileInfo) {
			defer wg.Done()

			authorsByEmail, lineCount, err := getBlameAuthors(ctx, f.repo, f.Name(), git.BlameOptions{NewestCommit: f.commit})
			if err != nil {
				mu.Lock()
				if allErr == nil {
					allErr = err
				}
				mu.Unlock()
				return
			}

			mu.Lock()
			defer mu.Unlock()
			totalLineCount += lineCount
			for email, a := range authorsByEmail {
				ca := all[email]
				if ca == nil {
					all[email] = a
				} else {
					ca.LineCount += a.LineCount
					if a.LastCommitDate.After(ca.LastCommitDate) {
						ca.Name = a.Name // use latest name in case it changed over time
						ca.LastCommit = a.LastCommit
						ca.LastCommitDate = a.LastCommitDate
					}
				}
			}
		}(f)
	}
	wg.Wait()
	if allErr != nil {
		return nil, allErr
	}

	edges := make([]gql.ComponentAuthorEdgeResolver, 0, len(all))
	for _, a := range all {
		edges = append(edges, &componentAuthorEdgeResolver{
			db:             r.db,
			data:           a,
			totalLineCount: totalLineCount,
		})
	}

	sort.Slice(edges, func(i, j int) bool {
		ei, ej := edges[i], edges[j]
		if ei.AuthoredLineCount() != ej.AuthoredLineCount() {
			return ei.AuthoredLineCount() > ej.AuthoredLineCount()
		}
		return ei.Person().Email() < ej.Person().Email()
	})

	return &contributorConnectionResolver{edges: edges, first: args.First}, nil
}

type componentAuthorEdgeResolver struct {
	db             database.DB
	data           *blameAuthor
	totalLineCount int
}

func (r *componentAuthorEdgeResolver) Person() *gql.PersonResolver {
	return gql.NewPersonResolver(r.db, r.data.Name, r.data.Email, true)
}

func (r *componentAuthorEdgeResolver) AuthoredLineCount() int32 {
	return int32(r.data.LineCount)
}

func (r *componentAuthorEdgeResolver) AuthoredLineProportion() float64 {
	return float64(r.data.LineCount) / float64(r.totalLineCount)
}

func (r *componentAuthorEdgeResolver) LastCommit(ctx context.Context) (*gql.GitCommitResolver, error) {
	repo, err := r.db.Repos().GetByName(ctx, r.data.LastCommitRepo)
	if err != nil {
		return nil, err
	}
	repoResolver := gql.NewRepositoryResolver(r.db, repo)
	return gql.NewGitCommitResolver(r.db, repoResolver, r.data.LastCommit, nil), nil
}

type contributorConnectionResolver struct {
	edges []gql.ComponentAuthorEdgeResolver
	first *int32
}

func (r *contributorConnectionResolver) Edges() []gql.ComponentAuthorEdgeResolver {
	if r.first != nil && len(r.edges) > int(*r.first) {
		return r.edges[:int(*r.first)]
	}
	return r.edges
}

func (r *contributorConnectionResolver) TotalCount() int32 {
	return int32(len(r.edges))
}

func (r *contributorConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	return graphqlutil.HasNextPage(r.first != nil && int(*r.first) < len(r.edges))
}
