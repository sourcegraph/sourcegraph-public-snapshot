package codycontext

import (
	"github.com/grafana/regexp"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
	"slices"
)

type enterpriseRepoFilter struct {
	cache safeCache[api.RepoName, bool]
	ccf   *schema.CodyContextFilters
}

// newEnterpriseFilter creates a new RepoContentFilter that filters out
// content based on the Cody context filters value in the site config.
func newEnterpriseFilter() RepoContentFilter {
	filter := &enterpriseRepoFilter{
		cache: newSafeCache[api.RepoName, bool](128),
		ccf:   conf.Get().SiteConfiguration.CodyContextFilters,
	}
	go conf.Watch(func() {
		filter.cache.Clear()
		filter.ccf = conf.Get().SiteConfiguration.CodyContextFilters
	})
	return filter
}

// GetFilter returns the list of repos that can be filtered based on the Cody context filter value in the site config.
func (f *enterpriseRepoFilter) GetFilter(repos []types.RepoIDName, _ log.Logger) ([]types.RepoIDName, FileChunkFilterFunc) {
	if f.ccf == nil {
		return repos, func(fcc []FileChunkContext) []FileChunkContext {
			return fcc
		}
	}
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

// isRepoAllowed checks if repo name matches Cody context include and exclude rules from the site config and stores result in cache.
func (f *enterpriseRepoFilter) isRepoAllowed(repoName api.RepoName) bool {
	cached, ok := f.cache.Get(repoName)
	if ok {
		return cached
	}

	allowed := true

	if len(f.ccf.Include) > 0 {
		for _, p := range f.ccf.Include {
			include := regexp.MustCompile(p.RepoNamePattern).MatchString(string(repoName))
			allowed = include
			if include {
				break
			}
		}
	}

	if len(f.ccf.Exclude) > 0 {
		for _, p := range f.ccf.Exclude {
			exclude := regexp.MustCompile(p.RepoNamePattern).MatchString(string(repoName))
			if exclude {
				allowed = false
				break
			}
		}
	}

	f.cache.Add(repoName, allowed)
	return allowed
}
