package resolvers

import (
	"context"
	"io/fs"
	"sort"
	"sync"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

type fileInfo struct {
	fs.FileInfo
	repo   api.RepoName
	commit api.CommitID
}

func (r *componentResolver) allFilesInSourceLocations(ctx context.Context) ([]fileInfo, error) {
	slocs, err := r.sourceLocations(ctx)
	if err != nil {
		return nil, err
	}

	var allFiles []fileInfo
	for _, sloc := range slocs {
		for _, path := range sloc.paths {
			// TODO(sqs): doesnt check perms? SECURITY
			entries, err := git.ReadDir(ctx, sloc.repoName, sloc.commitID, path, true)
			if err != nil {
				return nil, err
			}
			for _, e := range entries {
				if !e.Mode().IsRegular() {
					continue // ignore dirs and submodules
				}
				allFiles = append(allFiles, fileInfo{
					FileInfo: e,
					repo:     sloc.repoName,
					commit:   sloc.commitID,
				})
			}
		}
	}
	return allFiles, nil
}

func (r *componentResolver) Authors(ctx context.Context) (*[]gql.ComponentAuthorEdgeResolver, error) {
	allFiles, err := r.allFilesInSourceLocations(ctx)
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
			component:      r,
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

	return &edges, nil
}

type componentAuthorEdgeResolver struct {
	db             database.DB
	component      *componentResolver
	data           *blameAuthor
	totalLineCount int
}

func (r *componentAuthorEdgeResolver) Component() gql.ComponentResolver {
	return r.component
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
