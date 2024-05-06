package codycontext

import (
	"context"
	"slices"
	"sync"

	"github.com/grafana/regexp"
	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

const allowByDefault = true

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

type filterItem struct {
	RepoNamePattern *regexp.Regexp
}

type filtersConfig struct {
	cache   *lru.Cache[api.RepoID, bool]
	include []filterItem
	exclude []filterItem
}

type enterpriseRepoFilter struct {
	mu            sync.RWMutex
	logger        log.Logger
	db            database.DB
	fc            filtersConfig
	isConfigValid bool
}

// newEnterpriseFilter creates a new repoContentFilter that filters out
// content based on the Cody context filters value in the site config.
func newEnterpriseFilter(logger log.Logger, db database.DB) repoContentFilter {
	f := &enterpriseRepoFilter{
		logger: logger.Scoped("filter"),
		db:     db,
	}
	f.configure()
	conf.Watch(func() {
		f.configure()
	})
	return f
}

func (f *enterpriseRepoFilter) getFiltersConfig() (_ filtersConfig, ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.fc, f.isConfigValid
}

// getMatcher returns the list of repos that can be filtered based on the Cody context filter value in the site config.
func (f *enterpriseRepoFilter) getMatcher(ctx context.Context, repos []types.RepoIDName) ([]types.RepoIDName, fileMatcher, error) {
	// TODO: remove this check after `CodyContextFilters` support is added to the IDE clients.
	enabled, err := checkFeatureFlagEnabled(ctx, f.db)
	if err != nil {
		return []types.RepoIDName{}, func(api.RepoID, string) bool { return false }, err
	}
	if !enabled {
		// Cody context filters are not enabled, so allow everything.
		return repos, func(api.RepoID, string) bool { return true }, nil
	}

	fc, ok := f.getFiltersConfig()
	if !ok {
		// our configuration is invalid, so filter everything out.
		return []types.RepoIDName{}, func(api.RepoID, string) bool { return false }, errors.New("Cody context filters configuration is invalid. Please contact your admin.")
	}

	allowedRepos := make([]types.RepoIDName, 0, len(repos))
	for _, repo := range repos {
		if fc.isRepoAllowed(repo) {
			allowedRepos = append(allowedRepos, repo)
		}
	}

	return allowedRepos, func(repo api.RepoID, path string) bool {
		return slices.ContainsFunc(allowedRepos, func(r types.RepoIDName) bool { return r.ID == repo })
	}, nil
}

func (f *enterpriseRepoFilter) configure() {
	fc, err := parseCodyContextFilters(conf.Get().SiteConfiguration.CodyContextFilters)

	f.mu.Lock()
	defer f.mu.Unlock()

	if err != nil {
		f.logger.Error("Failed to configure filter. Defaulting to ignoring all context. Please fix cody.contextFilters in site configuration.", log.Error(err))
		f.isConfigValid = false
		return
	}

	f.fc = fc
	f.isConfigValid = true
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

	// ignore error since it only happens if cache size is not positive
	cache, _ := lru.New[api.RepoID, bool](128)

	return filtersConfig{
		cache:   cache,
		include: include,
		exclude: exclude,
	}, nil
}

// isRepoAllowed checks if repo name matches Cody context include and exclude rules from the site config and stores result in cache.
func (f filtersConfig) isRepoAllowed(repo types.RepoIDName) bool {
	cached, ok := f.cache.Get(repo.ID)
	if ok {
		metricCacheHit.Inc()
		return cached
	}
	metricCacheMiss.Inc()

	allowed := allowByDefault

	if len(f.include) > 0 {
		for _, p := range f.include {
			include := p.RepoNamePattern.MatchString(string(repo.Name))
			allowed = include
			if include {
				break
			}
		}
	}

	if len(f.exclude) > 0 {
		for _, p := range f.exclude {
			exclude := p.RepoNamePattern.MatchString(string(repo.Name))
			if exclude {
				allowed = false
				break
			}
		}
	}

	f.cache.Add(repo.ID, allowed)
	return allowed
}
