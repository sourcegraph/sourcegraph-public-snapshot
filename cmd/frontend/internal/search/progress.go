package search

import (
	"fmt"
	"time"

	sgapi "github.com/sourcegraph/sourcegraph/internal/api"
	searchshared "github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
)

type progressAggregator struct {
	Start        time.Time
	MatchCount   int
	Stats        streaming.Stats
	Limit        int
	DisplayLimit int
	Trace        string // may be empty

	// Dirty is true if p has changed since the last call to Current.
	Dirty bool
}

func (p *progressAggregator) Update(event streaming.SearchEvent) {
	if len(event.Results) == 0 && event.Stats.Zero() {
		return
	}

	p.Dirty = true
	p.Stats.Update(&event.Stats)
	for _, match := range event.Results {
		p.MatchCount += match.ResultCount()
	}

	if p.MatchCount > p.Limit {
		p.MatchCount = p.Limit
		p.Stats.IsLimitHit = true
	}
}

func (p *progressAggregator) currentStats() api.ProgressStats {
	// Suggest the next 1000 after rounding off.
	suggestedLimit := (p.Limit + 1500) / 1000 * 1000

	return api.ProgressStats{
		MatchCount:          p.MatchCount,
		ElapsedMilliseconds: int(time.Since(p.Start).Milliseconds()),
		ExcludedArchived:    p.Stats.ExcludedArchived,
		ExcludedForks:       p.Stats.ExcludedForks,
		Timedout:            getNames(p.Stats, searchshared.RepoStatusTimedout),
		Missing:             getNames(p.Stats, searchshared.RepoStatusMissing),
		Cloning:             getNames(p.Stats, searchshared.RepoStatusCloning),
		LimitHit:            p.Stats.IsLimitHit,
		SuggestedLimit:      suggestedLimit,
		Trace:               p.Trace,
		DisplayLimit:        p.DisplayLimit,
	}
}

// Current returns the current progress event.
func (p *progressAggregator) Current() api.Progress {
	p.Dirty = false

	return api.BuildProgressEvent(p.currentStats())
}

// Final returns the current progress event, but with final fields set to
// indicate it is the last progress event.
func (p *progressAggregator) Final() api.Progress {
	p.Dirty = false

	s := p.currentStats()

	// We only send RepositoriesCount at the end because the number is
	// confusing to users to see while searching.
	if c := len(p.Stats.Repos); c > 0 {
		s.RepositoriesCount = intPtr(c)
	}

	event := api.BuildProgressEvent(s)
	event.Done = true
	return event
}

type namerFunc string

func (n namerFunc) Name() string {
	return string(n)
}

func getNames(stats streaming.Stats, status searchshared.RepoStatus) []api.Namer {
	var names []api.Namer
	stats.Status.Filter(status, func(id sgapi.RepoID) {
		if name, ok := stats.Repos[id]; ok {
			names = append(names, namerFunc(name.Name))
		} else {
			names = append(names, namerFunc(fmt.Sprintf("UNKNOWN{ID=%d}", id)))
		}
	})
	return names
}

func intPtr(i int) *int {
	return &i
}
