package codycontext

import (
	"fmt"
	"github.com/grafana/regexp"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type FileChunkFilterFunc func([]FileChunkContext) []FileChunkContext
type safeCache struct {
	cache *lru.Cache[api.RepoName, bool]
}

func (c *safeCache) Add(key api.RepoName, value bool) (evicted bool) {
	if c.cache != nil {
		return c.cache.Add(key, value)
	}
	return false
}

func (c *safeCache) Get(key api.RepoName) (bool, bool) {
	if c.cache != nil {
		return c.cache.Get(key)

	}
	return false, false
}

func (c *safeCache) Clear() {
	if c.cache != nil {
		c.cache.Purge()

	}
}

type repoFilter struct {
	cache safeCache
	ccf   *schema.CodyContextFilters
}

type RepoContentFilter interface {
	GetFilter(repos []types.RepoIDName, logger log.Logger) ([]types.RepoIDName, FileChunkFilterFunc)
}

func NewCodyIgnoreFilter() RepoContentFilter {
	c, _ := lru.New[api.RepoName, bool](128)
	filter := &repoFilter{
		cache: safeCache{c},
		ccf:   conf.Get().SiteConfiguration.CodyContextFilters,
	}
	go conf.Watch(func() {
		fmt.Println("NewCodyIgnoreFilter: config changed, Cody filters changed: ", filter.ccf != conf.Get().SiteConfiguration.CodyContextFilters)
		filter.cache.Clear()
		filter.ccf = conf.Get().SiteConfiguration.CodyContextFilters
	})
	return filter
}

func (f *repoFilter) GetFilter(repos []types.RepoIDName, logger log.Logger) ([]types.RepoIDName, FileChunkFilterFunc) {
	allowedRepos := map[api.RepoName]bool{}
	for _, repo := range repos {
		allowed, err := f.isRepoAllowed(repo.Name)
		fmt.Printf("repoFilter.GetFilter: %v is allowed - %v", repo.Name, allowed)
		if err != nil {
			// TODO: context filter might be misconfigured: invalid regex or whatever
			continue
		}
		if allowed {
			allowedRepos[repo.Name] = true
		}
	}
	return repos, func(fcc []FileChunkContext) []FileChunkContext {
		filtered := make([]FileChunkContext, 0, len(fcc))
		for _, fc := range fcc {
			if allowedRepos[fc.RepoName] {
				filtered = append(filtered, fc)
			}
		}
		return filtered
	}
}

func (f *repoFilter) isRepoAllowed(repoName api.RepoName) (bool, error) {
	cached, ok := f.cache.Get(repoName)
	if ok {
		return cached, nil
	}

	allowed := true

	if len(f.ccf.Include) > 0 {
		for _, p := range f.ccf.Include {
			include, err := regexp.MatchString(p.RepoNamePattern, string(repoName))
			if err != nil {
				// TODO: return descriptive error
				return false, err
			}
			allowed = include
			if include {
				break
			}
		}
	}

	if len(f.ccf.Exclude) > 0 {
		for _, p := range f.ccf.Exclude {
			exclude, err := regexp.MatchString(p.RepoNamePattern, string(repoName))
			if err != nil {
				// TODO: return descriptive error
				return false, err
			}
			if exclude {
				allowed = false
				break
			}
		}
	}

	f.cache.Add(repoName, allowed)
	return allowed, nil
}
