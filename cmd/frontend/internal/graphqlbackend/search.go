package graphqlbackend

import (
	"context"
	"sort"
	"strings"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

	"log"

	"github.com/felixfbecker/stringscore"
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
	// score defines how well this item matches the query for sorting purposes
	score int
	// length holds the length of the item name as a second sorting criterium
	length int
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
		defer func() {
			if r := recover(); r != nil {
				log15.Error("unexpected panic while searching repos", "error", r)
				done <- nil
			}
		}()
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

	// Search search profiles
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log15.Error("unexpected panic while searching search profiles", "error", r)
				done <- nil
			}
		}()
		searchProfileResults, err := searchSearchProfiles(ctx, r, args.Query, limit)
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
			log15.Error("search error", "error", err)
		}
	}

	// Sort by score
	sort.Slice(res, func(i, j int) bool {
		a, b := res[i], res[j]
		if a.score != b.score {
			return a.score > b.score
		}
		// Prefer shorter strings for the same match score
		// E.g. prefer gorilla/mux over gorilla/muxy, Microsoft/vscode over g3ortega/vscode-crystal
		return a.length < b.length
	})

	// Limit
	if len(res) > limit {
		res = res[0:limit]
	}

	return res, nil
}

func searchSearchProfiles(ctx context.Context, rootResolver *rootResolver, query string, limit int) (res []*searchResultResolver, err error) {
	searchProfiles, err := rootResolver.SearchProfiles(ctx)
	if err != nil {
		return nil, err
	}
	for _, searchProfile := range searchProfiles {
		score := stringscore.Score(searchProfile.name, query)
		if score > 0 {
			score += 100
			res = append(res, &searchResultResolver{result: searchProfile, score: score, length: len(searchProfile.name)})
		}
	}
	return res, nil
}

func searchRepos(ctx context.Context, query string, repoURIs []string, limit int) (res []*searchResultResolver, err error) {
	opt := &sourcegraph.RepoListOptions{Query: query}
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
		score := 0
		// Score each repo URI path individually and use the sum
		// For the query "kubernetes" github.com/kubernetes/kubernetes should be higher than github.com/kubernetes/helm
		uriParts := strings.Split(repo.URI, "/")
		queryParts := strings.Split(query, "/")
		for _, uriPart := range uriParts {
			for _, queryPart := range queryParts {
				score += stringscore.Score(uriPart, queryPart)
			}
		}
		// Push forks down
		if repo.Fork {
			score -= 10
		}
		if score > 0 {
			res = append(res, &searchResultResolver{result: repoResolver, score: score, length: len(repo.URI)})
		}
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
	if commitStateResolver.cloneInProgress {
		// TODO report a cloning repo
		return res, nil
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
		score := stringscore.Score(fileResolver.name, query)
		if score > 0 {
			// Give files a slight advantage over repos
			score += 5
			res = append(res, &searchResultResolver{result: fileResolver, score: score, length: len(fileResolver.name)})
		}
	}
	return res, nil
}
