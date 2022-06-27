package streaming

import (
	"sort"

	sgapi "github.com/sourcegraph/sourcegraph/internal/api"
	searchshared "github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
)

type progressAggregator struct {
	Stats streaming.Stats
	Dirty bool

	RepoNamer api.RepoNamer
}

func (p *progressAggregator) currentStats() api.ProgressStats {
	return api.ProgressStats{
		ExcludedArchived: p.Stats.ExcludedArchived,
		ExcludedForks:    p.Stats.ExcludedForks,
		Timedout:         getRepos(p.Stats, searchshared.RepoStatusTimedout),
		Missing:          getRepos(p.Stats, searchshared.RepoStatusMissing),
		Cloning:          getRepos(p.Stats, searchshared.RepoStatusCloning),
		LimitHit:         p.Stats.IsLimitHit,
	}
}

// Current returns the current progress event.
func (p *progressAggregator) Current() api.Progress {
	return api.BuildProgressEvent(p.currentStats(), p.RepoNamer)
}

// Final returns the current progress event, but with final fields set to
// indicate it is the last progress event.
func (p *progressAggregator) Final() api.Progress {
	p.Dirty = false

	event := api.BuildProgressEvent(p.currentStats(), p.RepoNamer)
	event.Done = true
	return event
}

func getRepos(stats streaming.Stats, status searchshared.RepoStatus) []sgapi.RepoID {
	var repos []sgapi.RepoID
	stats.Status.Filter(status, func(id sgapi.RepoID) {
		repos = append(repos, id)
	})
	// Filter runs in a random order (map traversal), so we should sort to
	// give deterministic messages between updates.
	sort.Slice(repos, func(i, j int) bool {
		return repos[i] < repos[j]
	})
	return repos
}
