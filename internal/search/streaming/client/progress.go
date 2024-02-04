package client

import (
	"slices"
	"time"

	sgapi "github.com/sourcegraph/sourcegraph/internal/api"
	searchshared "github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

type ProgressAggregator struct {
	Start        time.Time
	MatchCount   int
	Stats        streaming.Stats
	Limit        int
	DisplayLimit int
	Trace        string // may be empty

	RepoNamer api.RepoNamer

	// Dirty is true if p has changed since the last call to Current.
	Dirty bool
}

func (p *ProgressAggregator) Update(event streaming.SearchEvent) {
	if len(event.Results) == 0 && event.Stats.Zero() {
		return
	}

	if p.Stats.Repos == nil {
		p.Stats.Repos = map[sgapi.RepoID]struct{}{}
	}

	p.Dirty = true
	p.Stats.Update(&event.Stats)
	for _, match := range event.Results {
		p.MatchCount += match.ResultCount()

		// Historically we only had one event populate Stats.Repos and it was
		// the full universe of repos. With Repo Pagination this is no longer
		// true. Rather than updating every backend to populate this field, we
		// iterate over results and union in the result IDs.
		p.Stats.Repos[match.RepoName().ID] = struct{}{}
	}

	if p.MatchCount > p.Limit {
		p.MatchCount = p.Limit
		p.Stats.IsLimitHit = true
	}
}

func (p *ProgressAggregator) currentStats() api.ProgressStats {
	// Suggest the next 1000 after rounding off.
	suggestedLimit := (p.Limit + 1500) / 1000 * 1000

	return api.ProgressStats{
		MatchCount:          p.MatchCount,
		ElapsedMilliseconds: int(time.Since(p.Start).Milliseconds()),
		BackendsMissing:     p.Stats.BackendsMissing,
		ExcludedArchived:    p.Stats.ExcludedArchived,
		ExcludedForks:       p.Stats.ExcludedForks,
		Timedout:            getRepos(p.Stats, searchshared.RepoStatusTimedOut),
		Missing:             getRepos(p.Stats, searchshared.RepoStatusMissing),
		Cloning:             getRepos(p.Stats, searchshared.RepoStatusCloning),
		LimitHit:            p.Stats.IsLimitHit,
		SuggestedLimit:      suggestedLimit,
		Trace:               p.Trace,
		DisplayLimit:        p.DisplayLimit,
	}
}

// Current returns the current progress event.
func (p *ProgressAggregator) Current() api.Progress {
	p.Dirty = false

	return api.BuildProgressEvent(p.currentStats(), p.RepoNamer)
}

// Final returns the current progress event, but with final fields set to
// indicate it is the last progress event.
func (p *ProgressAggregator) Final() api.Progress {
	p.Dirty = false

	s := p.currentStats()

	// We only send RepositoriesCount at the end because the number is
	// confusing to users to see while searching.
	if c := len(p.Stats.Repos); c > 0 {
		s.RepositoriesCount = pointers.Ptr(c)
	}

	event := api.BuildProgressEvent(s, p.RepoNamer)
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
	slices.Sort(repos)
	return repos
}
