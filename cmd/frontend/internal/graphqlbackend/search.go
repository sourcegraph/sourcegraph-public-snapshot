package graphqlbackend

import (
	"context"
	"math"
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
	// label to sort alphabetically by when all else is equal.
	label string
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
	scorer := newScorer(query)
	for _, searchProfile := range searchProfiles {
		score := scorer.calcScore(searchProfile)
		if score > 0 {
			res = append(res, newSearchResultResolver(searchProfile, score))
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
	scorer := newScorer(query)
outer:
	for _, repo := range reposList.Repos {
		// Don't suggest repos that were already added as a filter
		for _, repoURI := range repoURIs {
			if repoURI == repo.URI {
				continue outer
			}
		}
		repoResolver := &repositoryResolver{repo: repo}

		score := scorer.calcScore(repoResolver)
		if score > 0 {
			res = append(res, newSearchResultResolver(repoResolver, score))
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
	scorer := newScorer(query)
	for _, fileResolver := range treeResolver.Files() {
		if len(res) >= limit {
			return res, nil
		}
		score := scorer.calcScore(fileResolver)
		if score > 0 {
			res = append(res, newSearchResultResolver(fileResolver, score))
		}
	}
	return res, nil
}

// newSearchResultResolver returns a new searchResultResolver wrapping the
// given result.
//
// A panic occurs if the type of result is not a *repositoryResolver,
// *fileResolver, or *searchProfile.
func newSearchResultResolver(result interface{}, score int) *searchResultResolver {
	switch r := result.(type) {
	case *repositoryResolver:
		return &searchResultResolver{result: r, score: score, length: len(r.repo.URI), label: r.repo.URI}

	case *fileResolver:
		return &searchResultResolver{result: r, score: score, length: len(r.name), label: r.name}

	case *searchProfile:
		return &searchResultResolver{result: r, score: score, length: len(r.name), label: r.name}

	default:
		panic("never here")
	}
}

// scorer is a structure for holding some scorer state that can be shared
// across calcScore calls for the same query string.
type scorer struct {
	query      string
	queryParts []string
}

// newScorer returns a scorer to be used for calculating sort scores of results
// against the specified query.
func newScorer(query string) *scorer {
	return &scorer{
		query:      query,
		queryParts: strings.Split(query, "/"),
	}
}

// score values to add to different types of results to e.g. get forks lower in
// search results, etc.
const (
	// Search Profiles > Files > Repos > Forks
	scoreBumpSearchProfile = 2 * (math.MaxInt32 / 16)
	scoreBumpFile          = 1 * (math.MaxInt32 / 16)
	scoreBumpRepo          = 0 * (math.MaxInt32 / 16)
	scoreBumpFork          = -50
)

// calcScore calculates and assigns the sorting score to the given result.
//
// A panic occurs if the type of result is not a *repositoryResolver,
// *fileResolver, or *searchProfile.
func (s *scorer) calcScore(result interface{}) int {
	switch r := result.(type) {
	case *repositoryResolver:
		score := stringscore.Score(r.repo.URI, s.query)
		// Assume the query is written to match the postfix of the paths.
		// For the query "kubernetes" github.com/kubernetes/kubernetes should be higher than github.com/kubernetes/helm
		if len(s.queryParts) > 0 {
			score += postfixAlignScore(strings.Split(r.repo.URI, "/"), s.queryParts)
		}
		// Push forks down
		if r.repo.Fork {
			score += scoreBumpFork
		}
		if score > 0 {
			score += scoreBumpRepo
		}
		return score

	case *fileResolver:
		// example query: "bar/baz.go"
		// example r.path: "a/b/foo/bar/baz.go"
		score := 0
		pathParts := strings.Split(r.path, "/")
		// aligned query matches get 4x multiplier (3x here, 1x in the next loop)
		score += 3 * postfixAlignScore(pathParts, s.queryParts)
		for _, pathPart := range pathParts {
			// For all path parts (including aligned ones like above), match
			// against every query part.
			for _, queryPart := range s.queryParts {
				score += stringscore.Score(pathPart, queryPart)
			}
		}
		if score > 0 {
			score += scoreBumpFile
		}
		return score

	case *searchProfile:
		score := stringscore.Score(r.name, s.query)
		if score > 0 {
			score += scoreBumpSearchProfile
		}
		return score

	default:
		panic("never here")
	}
}

// postfixAlignScore is used to calculate how well the end of a target aligns with a query.
// targetParts and queryParts are the original target and query split into components.
//
// For example a query "b" or "a/b" will score higher to strings _ending_ with
// that instead of simply _containing_ it. i.e., this ordering:
//
// 	/a/b
// 	/x/a/b/y
//
// not:
//
// 	/x/a/b/y
// 	/a/b
func postfixAlignScore(targetParts, queryParts []string) int {
	score := 0
	for i := 1; len(targetParts)-i >= 0 && len(queryParts)-i >= 0; i++ {
		score += stringscore.Score(targetParts[len(targetParts)-i], queryParts[len(queryParts)-i])
	}
	return score
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
	if a.length != b.length {
		return a.length < b.length
	}

	// All else equal, sort alphabetically.
	return a.label < b.label
}
