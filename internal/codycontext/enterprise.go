package codycontext

import (
	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"slices"
	"sync"
)

const allowByDefault = true

type filterItem struct {
	RepoNamePattern regexp.Regexp
}

type filtersConfig struct {
	Include []filterItem
	Exclude []filterItem
}

type enterpriseRepoFilter struct {
	cache safeCache[api.RepoName, bool]
	ccf   filtersConfig
	mu    sync.RWMutex
}

// newEnterpriseFilter creates a new RepoContentFilter that filters out
// content based on the Cody context filters value in the site config.
func newEnterpriseFilter() (RepoContentFilter, error) {
	f := &enterpriseRepoFilter{}
	err := f.configure(conf.Get().SiteConfiguration.CodyContextFilters)
	if err != nil {
		return nil, err
	}
	// TODO: handle error!
	conf.Watch(func() {
		// TODO: how do I find out that CodyContextFilters changed?
		// If they didn't, I dont want to re-configure filter
		// TODO: what to do with error here?
		e := f.configure(conf.Get().SiteConfiguration.CodyContextFilters)

	})
	return f, nil
}

// GetFilter returns the list of repos that can be filtered based on the Cody context filter value in the site config.
func (f *enterpriseRepoFilter) GetFilter(repos []types.RepoIDName, _ log.Logger) ([]types.RepoIDName, FileChunkFilterFunc) {
	allowedRepos := make([]types.RepoIDName, 0, len(repos))
	for _, repo := range repos {
		if f.isRepoAllowed(repo.Name) {
			allowedRepos = append(allowedRepos, repo)
		}
	}
	return allowedRepos, func(fcc []FileChunkContext) []FileChunkContext {
		filtered := make([]FileChunkContext, 0, len(fcc))
		for _, fc := range fcc {
			isFromAllowedRepo := slices.ContainsFunc(allowedRepos, func(r types.RepoIDName) bool { return r.Name == fc.RepoName })
			if isFromAllowedRepo {
				filtered = append(filtered, fc)
			}
		}
		return filtered
	}
}

func (f *enterpriseRepoFilter) configure(ccf *schema.CodyContextFilters) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	// TODO: where should I apply locks?

	// TODO: should I create new cache each time?
	f.cache = newSafeCache[api.RepoName, bool](128)
	f.ccf = filtersConfig{}

	if ccf == nil {
		f.ccf.Include = make([]filterItem, 0)
		f.ccf.Exclude = make([]filterItem, 0)
		return nil
	}

	if len(ccf.Include) > 0 {
		include := make([]filterItem, 0, len(ccf.Include))
		for _, p := range ccf.Include {
			re, err := regexp.Compile(p.RepoNamePattern)
			if err != nil {
				return err
			}
			include = append(include, filterItem{RepoNamePattern: *re})
		}
		f.ccf.Include = include
	}

	if len(ccf.Exclude) > 0 {
		exclude := make([]filterItem, 0, len(ccf.Exclude))
		for _, p := range ccf.Exclude {
			re, err := regexp.Compile(p.RepoNamePattern)
			if err != nil {
				return err
			}
			exclude = append(exclude, filterItem{RepoNamePattern: *re})
		}
		f.ccf.Exclude = exclude
	}

	return nil
}

// isRepoAllowed checks if repo name matches Cody context include and exclude rules from the site config and stores result in cache.
func (f *enterpriseRepoFilter) isRepoAllowed(repoName api.RepoName) bool {
	// TODO: how do we apply locks to f.ccf and f.cache?
	cached, ok := f.cache.Get(repoName)
	if ok {
		return cached
	}

	allowed := allowByDefault

	if len(f.ccf.Include) > 0 {
		for _, p := range f.ccf.Include {
			include := p.RepoNamePattern.MatchString(string(repoName))
			allowed = include
			if include {
				break
			}
		}
	}

	if len(f.ccf.Exclude) > 0 {
		for _, p := range f.ccf.Exclude {
			exclude := p.RepoNamePattern.MatchString(string(repoName))
			if exclude {
				allowed = false
				break
			}
		}
	}

	// TODO: what if the cache has been already cleared as the new config arrived (see conf.Watch() above)?
	// We should probably compare caches (equality?) and do not write if the cache is new.
	f.cache.Add(repoName, allowed)
	return allowed
}
