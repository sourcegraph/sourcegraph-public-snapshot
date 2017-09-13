package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

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
	// result is either a repositoryResolver or a fileResolver
	result interface{}
	// distance is the string distance of this item from the search query
	distance int
}

func (r *searchResultResolver) ToRepository() (*repositoryResolver, bool) {
	res, ok := r.result.(*repositoryResolver)
	return res, ok
}

func (r *searchResultResolver) ToFile() (*fileResolver, bool) {
	res, ok := r.result.(*fileResolver)
	return res, ok
}

func (r *searchResultResolver) ToSearchProfile() (*searchProfile, bool) {
	res, ok := r.result.(*searchProfile)
	return res, ok
}

// Search searches over repos and their files
func (r *rootResolver) Search(ctx context.Context, args *searchArgs) ([]*searchResultResolver, error) {
	limit := 50
	if args.First != nil {
		limit = int(*args.First)
		if limit > 1000 {
			limit = 1000
		}
	}
	query := strings.ToLower(args.Query)

	var (
		resMu sync.Mutex
		res   []*searchResultResolver
	)

	done := make(chan error, 3)

	// Search files
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log15.Error("unexpected panic while searching files", "error", r)
				done <- nil
			}
		}()
		fileResults, err := searchFiles(ctx, query, args.Repositories, limit)
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
		defer func() {
			if r := recover(); r != nil {
				log15.Error("unexpected panic while searching repos", "error", r)
				done <- nil
			}
		}()
		repoResults, err := searchRepos(ctx, query, args.Repositories, limit)
		if err != nil {
			done <- err
			return
		}
		resMu.Lock()
		res = append(res, repoResults...)
		resMu.Unlock()
		done <- nil
	}()

	// Search search profiles
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log15.Error("unexpected panic while searching search profiles", "error", r)
				done <- nil
			}
		}()
		searchProfileResults, err := searchSearchProfiles(ctx, r, query, limit)
		if err != nil {
			done <- err
			return
		}
		resMu.Lock()
		res = append(res, searchProfileResults...)
		resMu.Unlock()
		done <- nil
	}()

	for i := 0; i < 3; i++ {
		if err := <-done; err != nil {
			// TODO collect error
			log.Println("search error: " + err.Error())
		}
	}

	if len(res) > limit {
		res = res[0:limit]
	}

	sort.Slice(res, func(i, j int) bool {
		return res[i].distance < res[j].distance
	})

	return res, nil
}

func searchSearchProfiles(ctx context.Context, rootResolver *rootResolver, query string, limit int) (res []*searchResultResolver, err error) {
	searchProfiles, err := rootResolver.SearchProfiles(ctx)
	if err != nil {
		return nil, err
	}
	queryRunes := []rune(query)
	for _, searchProfile := range searchProfiles {
		lowerName := strings.ToLower(searchProfile.name)
		if strings.Contains(lowerName, query) {
			distance := levenshtein.DistanceForStrings(queryRunes, []rune(lowerName), levenshtein.DefaultOptions)
			res = append(res, &searchResultResolver{result: searchProfile, distance: distance})
		}
	}
	return res, nil
}

func searchRepos(ctx context.Context, query string, repoURIs []string, limit int) (res []*searchResultResolver, err error) {
	opt := &sourcegraph.RepoListOptions{Query: query, RemoteSearch: false}
	opt.PerPage = int32(limit)
	reposList, err := backend.Repos.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	queryRunes := []rune(query)
outer:
	for _, repo := range reposList.Repos {
		// Don't suggest repos that were already added as a filter
		for _, repoURI := range repoURIs {
			if repoURI == repo.URI {
				continue outer
			}
		}
		repoResolver := &repositoryResolver{repo: repo}
		distance := levenshtein.DistanceForStrings(queryRunes, []rune(repo.URI), levenshtein.DefaultOptions)
		res = append(res, &searchResultResolver{result: repoResolver, distance: distance})
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
	queryRunes := []rune(query)
	for _, fileResolver := range treeResolver.Files() {
		if len(res) >= limit {
			return res, nil
		}
		name := strings.ToLower(fileResolver.Name())
		if strings.Contains(name, query) {
			distance := levenshtein.DistanceForStrings(queryRunes, []rune(name), levenshtein.DefaultOptions)
			res = append(res, &searchResultResolver{result: fileResolver, distance: distance})
		}
	}
	return res, nil
}
