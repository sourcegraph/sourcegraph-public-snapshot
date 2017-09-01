package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"sync"

	"log"

	"github.com/texttheater/golang-levenshtein/levenshtein"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/backend"
)

type searchArgs struct {
	// Query specifies
	Query string
	// Repositories specifies a list of repos to search for files
	Repositories []string
	// First limits the result
	First *int32
}

type searchResultResolver struct {
	// Either a repositoryResolver or a fileResolver
	result interface{}
	// The string this element should be sorted by
	sortText string
}

func (r *searchResultResolver) ToRepository() (*repositoryResolver, bool) {
	res, ok := r.result.(*repositoryResolver)
	return res, ok
}

func (r *searchResultResolver) ToFile() (*fileResolver, bool) {
	res, ok := r.result.(*fileResolver)
	return res, ok
}

// Search searches over repos and their files
func (*rootResolver) Search(ctx context.Context, args *searchArgs) ([]*searchResultResolver, error) {
	limit := 50
	if args.First != nil {
		limit = int(*args.First)
		if limit > 1000 {
			limit = 1000
		}
	}

	var (
		resMu sync.Mutex
		res   []*searchResultResolver
	)

	done := make(chan error, 2)

	// Search files
	go func() {
		fileResults, err := searchFiles(ctx, args.Query, args.Repositories, limit)
		if err != nil {
			done <- err
			return
		}
		resMu.Lock()
		res = append(res, fileResults...)
		resMu.Unlock()
		done <- nil
	}()

	// Search repos
	go func() {
		repoResults, err := searchRepos(ctx, args.Query, args.Repositories, limit)
		if err != nil {
			done <- err
			return
		}
		resMu.Lock()
		res = append(res, repoResults...)
		resMu.Unlock()
		done <- nil
	}()

	for i := 0; i < 2; i++ {
		if err := <-done; err != nil {
			// TODO collect error
			log.Println("search error: " + err.Error())
		}
	}

	if len(res) > limit {
		res = res[0:limit]
	}

	sort.Slice(res, func(i, j int) bool {
		query := []rune(args.Query)
		distI := levenshtein.DistanceForStrings(query, []rune(res[i].sortText), levenshtein.DefaultOptions)
		distJ := levenshtein.DistanceForStrings(query, []rune(res[j].sortText), levenshtein.DefaultOptions)
		return distI < distJ
	})

	return res, nil
}

func searchRepos(ctx context.Context, query string, repoURIs []string, limit int) (res []*searchResultResolver, err error) {
	opt := &sourcegraph.RepoListOptions{Query: query, RemoteSearch: false}
	opt.PerPage = int32(limit)
	reposList, err := backend.Repos.List(ctx, opt)
	if err != nil {
		return nil, err
	}
outer:
	for _, repo := range reposList.Repos {
		// Don't suggest repos that were already added as a filter
		for _, repoURI := range repoURIs {
			if repoURI == repo.URI {
				continue outer
			}
		}
		repoResolver := &repositoryResolver{repo: repo}
		res = append(res, &searchResultResolver{result: repoResolver, sortText: repo.URI})
	}
	return res, nil
}

func searchFiles(ctx context.Context, query string, repoURIs []string, limit int) ([]*searchResultResolver, error) {
	var (
		resMu sync.Mutex
		res   []*searchResultResolver
	)
	done := make(chan error, len(repoURIs))
	for _, repoURI := range repoURIs {
		repoURI := repoURI
		go func() {
			fileResults, err := searchFilesForRepoURI(ctx, query, repoURI, limit)
			if err != nil {
				done <- err
				return
			}
			resMu.Lock()
			res = append(res, fileResults...)
			resMu.Unlock()
			done <- nil
		}()
	}
	for range repoURIs {
		if err := <-done; err != nil {
			// TODO collect error
			log.Println("searchFiles error: " + err.Error())
		}
	}
	return res, nil
}

func searchFilesForRepoURI(ctx context.Context, query string, repoURI string, limit int) (res []*searchResultResolver, err error) {
	repo, err := backend.Repos.GetByURI(ctx, repoURI)
	if err != nil {
		return nil, err
	}
	repoResolver := &repositoryResolver{repo: repo}
	commitStateResolver, err := repoResolver.Commit(ctx, &struct {
		Rev string
	}{Rev: ""})
	if err != nil {
		return nil, err
	}
	commitResolver := commitStateResolver.Commit()
	treeResolver, err := commitResolver.Tree(ctx, &struct {
		Path      string
		Recursive bool
	}{Path: "", Recursive: true})
	if err != nil {
		return nil, err
	}
	for _, fileResolver := range treeResolver.Files() {
		if len(res) >= limit {
			return res, nil
		}
		name := fileResolver.Name()
		if strings.Contains(name, query) {
			res = append(res, &searchResultResolver{result: fileResolver, sortText: name})
		}
	}
	return res, nil
}
