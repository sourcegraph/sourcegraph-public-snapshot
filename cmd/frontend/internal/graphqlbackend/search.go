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

	// Sort search results.
	sort.Sort(searchResultSorter(res))

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
		score := calcScore(query, searchProfile)
		if score > 0 {
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

		score := calcScore(query, repoResolver)
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
		score := calcScore(query, fileResolver)
		if score > 0 {
			res = append(res, &searchResultResolver{result: fileResolver, score: score, length: len(fileResolver.name)})
		}
	}
	return res, nil
}

// calcScore calculates and assigns the sorting score to the given result.
func calcScore(query string, result interface{}) int {
	switch r := result.(type) {
	case *repositoryResolver:
		score := stringscore.Score(r.repo.URI, query)
		// Assume the query is written to match the postfix of the paths.
		// For the query "kubernetes" github.com/kubernetes/kubernetes should be higher than github.com/kubernetes/helm
		queryParts := strings.Split(query, "/")
		if len(queryParts) > 0 {
			repoParts := strings.Split(r.repo.URI, "/")
			for i := 1; len(queryParts)-i >= 0 && len(repoParts)-i >= 0; i++ {
				score += stringscore.Score(repoParts[len(repoParts)-i], queryParts[len(queryParts)-i])
			}
		}
		// Push forks down
		if r.repo.Fork {
			score -= 10
		}
		return score

	case *fileResolver:
		score := 0
		// Score each path component individually and use the sum
		// We don't want the query "openerService" to match
		//   src/vs/workbench/parts/execution/electron-browser/terminalService.ts
		// with a higher score than
		//   src/vs/platform/opener/browser/openerService.ts
		pathParts := strings.Split(r.name, "/")
		queryParts := strings.Split(query, "/")
		for _, pathPart := range pathParts {
			for _, queryPart := range queryParts {
				score += stringscore.Score(pathPart, queryPart)
			}
		}
		if score > 0 {
			// Give files a slight advantage over repos
			score += 5
		}
		return score

	case *searchProfile:
		score := stringscore.Score(r.name, query)
		if score > 0 {
			score += 100
		}
		return score
	default:
		panic("never here")
	}
}

// searchResultSorter implements the sort.Interface interface to sort a list of
// searchResultResolvers.
type searchResultSorter []*searchResultResolver

func (s searchResultSorter) Len() int      { return len(s) }
func (s searchResultSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s searchResultSorter) Less(i, j int) bool {
	// Sort by score
	a, b := s[i], s[j]
	if a.score != b.score {
		return a.score > b.score
	}
	// Prefer shorter strings for the same match score
	// E.g. prefer gorilla/mux over gorilla/muxy, Microsoft/vscode over g3ortega/vscode-crystal
	return a.length < b.length
}
