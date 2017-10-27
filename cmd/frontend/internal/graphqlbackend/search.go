package graphqlbackend

import (
	"context"
	"fmt"
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

var mockSearchFilesForRepoURI func(query string, repoURI string, limit int) ([]*searchResultResolver, error)

func searchFilesForRepoURI(ctx context.Context, query string, repoURI string, limit int) (res []*searchResultResolver, err error) {
	if mockSearchFilesForRepoURI != nil {
		return mockSearchFilesForRepoURI(query, repoURI, limit)
	}

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
	if commitResolver == nil {
		return nil, fmt.Errorf("unable to resolve commit for repo %s", repoURI)
	}
	treeResolver, err := commitResolver.Tree(ctx, &struct {
		Path      string
		Recursive bool
	}{Path: "", Recursive: true})
	if err != nil {
		return nil, err
	}
	scorer := newScorer(query)
	for _, fileResolver := range treeResolver.Files() {
		score := scorer.calcScore(fileResolver)
		if score > 0 {
			res = append(res, newSearchResultResolver(fileResolver, score))
		}
	}

	sort.Sort(searchResultSorter(res))
	if len(res) > limit {
		res = res[:limit]
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
	queryEmpty bool
	queryParts []string
}

// newScorer returns a scorer to be used for calculating sort scores of results
// against the specified query.
func newScorer(query string) *scorer {
	return &scorer{
		query:      query,
		queryEmpty: strings.TrimSpace(query) == "",
		queryParts: splitNoEmpty(query, "/"),
	}
}

// score values to add to different types of results to e.g. get forks lower in
// search results, etc.
const (
	// Search Profiles > Files > Repos > Forks
	scoreBumpSearchProfile = 2 * (math.MaxInt32 / 16)
	scoreBumpFile          = 1 * (math.MaxInt32 / 16)
	scoreBumpRepo          = 0 * (math.MaxInt32 / 16)
	scoreBumpFork          = -10
)

// calcScore calculates and assigns the sorting score to the given result.
//
// A panic occurs if the type of result is not a *repositoryResolver,
// *fileResolver, or *searchProfile.
func (s *scorer) calcScore(result interface{}) int {
	var score int
	if s.queryEmpty {
		// If no query, then it will show *all* results; score must be nonzero in order to
		// have scoreBump* constants applied.
		score = 1
	}

	switch r := result.(type) {
	case *repositoryResolver:
		if !s.queryEmpty {
			score = postfixFuzzyAlignScore(splitNoEmpty(r.repo.URI, "/"), s.queryParts)
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
		if !s.queryEmpty {
			pathParts := splitNoEmpty(r.path, "/")
			score = postfixFuzzyAlignScore(pathParts, s.queryParts)
		}
		if score > 0 {
			score += scoreBumpFile
		}
		return score

	case *searchProfile:
		if !s.queryEmpty {
			score = stringscore.Score(r.name, s.query)
		}
		if score > 0 {
			score += scoreBumpSearchProfile
		}
		return score

	default:
		panic("never here")
	}
}

// postfixFuzzyAlignScore is used to calculate how well a targets component
// matches a query from the back. It rewards consecutive alignment as well as
// aligning to the right. For example for the query "a/b" we get the
// following ranking:
//
//   /a/b == /x/a/b
//   /a/b/x
//   /a/x/b
//
// The following will get zero score
//
//   /x/b
//   /ab/
func postfixFuzzyAlignScore(targetParts, queryParts []string) int {
	total := 0
	consecutive := true
	queryIdx := len(queryParts) - 1
	for targetIdx := len(targetParts) - 1; targetIdx >= 0 && queryIdx >= 0; targetIdx-- {
		score := stringscore.Score(targetParts[targetIdx], queryParts[queryIdx])
		if score <= 0 {
			consecutive = false
			continue
		}
		// Consecutive and align bonus
		if consecutive {
			score *= 2
		}
		consecutive = true
		total += score
		queryIdx--
	}
	// Did not match whole of queryIdx
	if queryIdx >= 0 {
		return 0
	}
	return total
}

// splitNoEmpty is like strings.Split except empty strings are removed.
func splitNoEmpty(s, sep string) []string {
	split := strings.Split(s, sep)
	res := make([]string, 0, len(split))
	for _, part := range split {
		if part != "" {
			res = append(res, part)
		}
	}
	return res
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
