package codycontext

import (
	"bytes"
	"context"
	"io"
	"os"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/zoekt/ignore"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

const codyIgnoreFile = ".cody/ignore"

var (
	emptyMatcher ignore.Matcher = ignore.Matcher{}

	ignoreFileCacheHitCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "cody_ignore_file_cache_hit_count",
	})
	ignoreFileCacheMissCount = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "src",
		Name:      "cody_ignore_file_cache_miss_count",
	})
)

type repoRevision struct {
	Repo   types.RepoIDName
	Commit api.CommitID
}

type filterFunc func(string) bool
type FileChunkFilterFunc func([]FileChunkContext) []FileChunkContext
type safeCache struct {
	cache *lru.Cache[repoRevision, ignore.Matcher]
}

func (c *safeCache) Add(key repoRevision, value ignore.Matcher) (evicted bool) {
	if c.cache != nil {
		return c.cache.Add(key, value)
	}
	return false
}

func (c *safeCache) Get(key repoRevision) (ignore.Matcher, bool) {
	if c.cache != nil {
		v, ok := c.cache.Get(key)
		if ok {
			ignoreFileCacheHitCount.Inc()
		} else {
			ignoreFileCacheMissCount.Inc()
		}
		return v, ok
	}

	ignoreFileCacheMissCount.Inc()
	return ignore.Matcher{}, false
}

type repoFilter struct {
	cache  safeCache
	client gitserver.Client
}

// GetFilter returns the list of repos that can be filtered
// their .cody/ignore files (or don't have one). If an error
// occurs that repo will be excluded.
func (f *repoFilter) GetFilter(repos []types.RepoIDName) ([]types.RepoIDName, FileChunkFilterFunc) {
	filters := make(map[api.RepoName]filterFunc, len(repos))
	filterableRepos := make([]types.RepoIDName, 0, len(repos))
	// use the internal actor to ensure access to repo and ignore files
	ctx := actor.WithInternalActor(context.Background())
	for _, repo := range repos {

		_, commit, err := f.client.GetDefaultBranch(ctx, repo.Name, true)
		if err != nil {
			continue
		}
		// No commit signals an empty repo, should be nothing to filter
		// Also we can't lookup the ignore file without a commit
		if commit == "" {
			continue
		}
		matcher, err := getIgnoreMatcher(ctx, f.cache, f.client, repo, commit)
		if err != nil {
			continue
		}

		filters[repo.Name] = matcher.Match
		filterableRepos = append(filterableRepos, repo)
	}

	return filterableRepos, func(fcc []FileChunkContext) []FileChunkContext {
		filtered := make([]FileChunkContext, 0, len(fcc))
		for _, fc := range fcc {
			remove, ok := filters[fc.RepoName]
			if !ok {
				filtered = append(filtered, fc)
				continue
			}
			if !remove(fc.Path) {
				filtered = append(filtered, fc)
			}
		}
		return filtered
	}
}

type RepoContentFilter interface {
	GetFilter(repos []types.RepoIDName) ([]types.RepoIDName, FileChunkFilterFunc)
}

// NewCodyIgnoreFilter creates a new RepoContentFilter that filters out
// content based on the .cody/ignore file at the head of the default branch
// for the given repositories.
func NewCodyIgnoreFilter(client gitserver.Client) RepoContentFilter {
	c, _ := lru.New[repoRevision, ignore.Matcher](128)
	return &repoFilter{
		cache:  safeCache{c},
		client: client,
	}
}

func getIgnoreMatcher(ctx context.Context, cache safeCache, client gitserver.Client, repo types.RepoIDName, commit api.CommitID) (*ignore.Matcher, error) {
	cached, ok := cache.Get(repoRevision{Repo: repo, Commit: commit})
	if ok {
		return &cached, nil
	}

	fr, err := client.NewFileReader(
		ctx,
		repo.Name,
		commit,
		codyIgnoreFile,
	)
	if err != nil {
		// We do not ignore anything if the ignore file does not exist.
		if os.IsNotExist(err) {
			cache.Add(repoRevision{Repo: repo, Commit: commit}, emptyMatcher)
			return &emptyMatcher, nil
		}
		return nil, err
	}
	defer fr.Close()

	ignoreFileBytes, err := io.ReadAll(fr)
	if err != nil {
		return nil, err
	}
	ig, err := ignore.ParseIgnoreFile(bytes.NewReader(ignoreFileBytes))
	if err != nil {
		return nil, err
	}
	cache.Add(repoRevision{Repo: repo, Commit: commit}, *ig)
	return ig, nil
}
