package search

import (
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	sgapi "github.com/sourcegraph/sourcegraph/internal/api"
	searchshared "github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
)

type progressAggregator struct {
	Start      time.Time
	MatchCount int
	Stats      streaming.Stats
}

func (p *progressAggregator) Update(event graphqlbackend.SearchEvent) {
	// TODO implement Update such that we update api.ProgressStats to avoid
	// re-reading the whole of stats.

	p.Stats.Update(&event.Stats)

	for _, result := range event.Results {
		p.MatchCount += int(result.ResultCount())
	}
}

func (p *progressAggregator) Build() api.Progress {
	return api.BuildProgressEvent(api.ProgressStats{
		MatchCount:          p.MatchCount,
		ElapsedMilliseconds: int(time.Since(p.Start).Milliseconds()),
		RepositoriesCount:   len(p.Stats.Repos),
		ExcludedArchived:    p.Stats.ExcludedArchived,
		ExcludedForks:       p.Stats.ExcludedForks,
		Timedout:            getNames(p.Stats, searchshared.RepoStatusTimedout),
		Missing:             getNames(p.Stats, searchshared.RepoStatusMissing),
		Cloning:             getNames(p.Stats, searchshared.RepoStatusCloning),
		LimitHit:            p.Stats.IsLimitHit,
	})
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
