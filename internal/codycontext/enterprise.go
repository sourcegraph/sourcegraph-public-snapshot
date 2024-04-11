package codycontext

import (
	"slices"
	"sync"

	"github.com/grafana/regexp"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

const allowByDefault = true

type filterItem struct {
	RepoNamePattern *regexp.Regexp
}

type filtersConfig struct {
	cache   *lru.Cache[api.RepoName, bool]
	include []filterItem
	exclude []filterItem
}

type enterpriseRepoFilter struct {
	mu            sync.RWMutex
	ccf           filtersConfig
	isConfigValid bool
}

// newEnterpriseFilter creates a new RepoContentFilter that filters out
// content based on the Cody context filters value in the site config.
func newEnterpriseFilter(logger log.Logger) (RepoContentFilter, error) {
	logger = logger.Scoped("filter")

	f := &enterpriseRepoFilter{}
	err := f.configure()
	if err != nil {
		return nil, err
	}

	conf.Watch(func() {
		err := f.configure()
		if err != nil {
			logger.Error("Failed to configure filter. Defaulting to ignoring all context. Please fix cody.contextFilters in site configuration.", log.Error(err))
		}
	})
	return f, nil
}

func (f *enterpriseRepoFilter) getFiltersConfig() (_ filtersConfig, ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.ccf, f.isConfigValid
}

// GetFilter returns the list of repos that can be filtered based on the Cody context filter value in the site config.
func (f *enterpriseRepoFilter) GetFilter(repos []types.RepoIDName, _ log.Logger) ([]types.RepoIDName, FileChunkFilterFunc) {
	ccf, ok := f.getFiltersConfig()
	if !ok {
		// our configuration is invalid, so filter everything out.
		return []types.RepoIDName{}, func(fcc []FileChunkContext) []FileChunkContext { return nil }
	}

	allowedRepos := make([]types.RepoIDName, 0, len(repos))
	for _, repo := range repos {
		if ccf.isRepoAllowed(repo.Name) {
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

func (f *enterpriseRepoFilter) configure() error {
	ccf, err := parseCodyContextFilters(conf.Get().SiteConfiguration.CodyContextFilters)

	f.mu.Lock()
	defer f.mu.Unlock()

	if err != nil {
		f.isConfigValid = false
		return err
	}

	f.ccf = ccf
	f.isConfigValid = true

	return nil
}

func parseCodyContextFilters(ccf *schema.CodyContextFilters) (filtersConfig, error) {
	if ccf == nil {
		ccf = &schema.CodyContextFilters{}
	}

	var include []filterItem
	for _, p := range ccf.Include {
		re, err := regexp.Compile(p.RepoNamePattern)
		if err != nil {
			return filtersConfig{}, err
		}
		include = append(include, filterItem{RepoNamePattern: re})
	}

	var exclude []filterItem
	for _, p := range ccf.Exclude {
		re, err := regexp.Compile(p.RepoNamePattern)
		if err != nil {
			return filtersConfig{}, err
		}
		exclude = append(exclude, filterItem{RepoNamePattern: re})
	}

	// ignore error since it only happens if cache size is not positive.
	cache, _ := lru.New[api.RepoName, bool](128)

	return filtersConfig{
		cache:   cache,
		include: include,
		exclude: exclude,
	}, nil
}

// isRepoAllowed checks if repo name matches Cody context include and exclude rules from the site config and stores result in cache.
func (f filtersConfig) isRepoAllowed(repoName api.RepoName) bool {
	cached, ok := f.cache.Get(repoName)
	if ok {
		metricCacheHit.Inc()
		return cached
	}
	metricCacheMiss.Inc()

	allowed := allowByDefault

	if len(f.include) > 0 {
		for _, p := range f.include {
			include := p.RepoNamePattern.MatchString(string(repoName))
			allowed = include
			if include {
				break
			}
		}
	}

	if len(f.exclude) > 0 {
		for _, p := range f.exclude {
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

var (
	metricCacheHit = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_codycontext_filter_cache_hit",
		Help: "Incremented each time we have a cache hit on cody context filters.",
	})
	metricCacheMiss = promauto.NewCounter(prometheus.CounterOpts{
		Name: "src_codycontext_filter_cache_miss",
		Help: "Incremented each time we have a cache miss on cody context filters.",
	})
)
