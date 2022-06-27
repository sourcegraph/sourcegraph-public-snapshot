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
		Timedout: getRepos(p.Stats, searchshared.RepoStatusTimedout),
	}
}

// Current returns the current progress event.
func (p *progressAggregator) Current() api.Progress {
	return api.BuildProgressEvent(p.currentStats(), p.RepoNamer)
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
