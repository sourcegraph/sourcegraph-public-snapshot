package graphqlbackend

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func NewAndJob(children ...run.Job) run.Job {
	if len(children) == 0 {
		return run.NewNoopJob()
	} else if len(children) == 1 {
		return children[0]
	}
	return &AndJob{children: children}
}

type AndJob struct {
	children []run.Job
}

func (a *AndJob) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "AndJob", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	var (
		g           multierror.Group
		maxAlerter  search.MaxAlerter
		intersector = newLiveIntersector(len(a.children))
	)
	for i, child := range a.children {
		i, child := i, child
		g.Go(func() error {
			intersectingStream := streaming.StreamFunc(func(event streaming.SearchEvent) {
				streamableResults := intersector.addMatches(event.Results, i)
				if len(streamableResults) > 0 {
					event.Results = streamableResults
				} else {
					event.Results = nil
				}
				stream.Send(event)
			})

			alert, err := child.Run(ctx, db, intersectingStream)
			maxAlerter.Add(alert)
			return err
		})
	}

	return maxAlerter.Alert, g.Wait().ErrorOrNil()
}

func (a *AndJob) Name() string {
	return "AndJob"
}

func newLiveIntersector(numSources int) *liveIntersector {
	return &liveIntersector{
		numSources: numSources,
		matches:    make(map[result.Key]intersectorVal, 100),
	}
}

type liveIntersector struct {
	mu         sync.Mutex
	numSources int
	matches    map[result.Key]intersectorVal
}

type intersectorVal struct {
	match       result.Match
	sourceMarks []bool
	sent        bool
}

func (s *liveIntersector) addMatches(matches result.Matches, source int) result.Matches {
	s.mu.Lock()
	defer s.mu.Unlock()

	var streamableMatches result.Matches
	for _, match := range matches {
		streamableMatch := s.addMatch(match, source)
		if streamableMatch != nil {
			streamableMatches = append(streamableMatches, streamableMatch)
		}
	}
	return streamableMatches
}

func (s *liveIntersector) addMatch(m result.Match, source int) result.Match {
	// Check if we've seen the match before
	key := m.Key()
	prev, ok := s.matches[key]
	if prev.sent {
		return nil
	}
	if !ok {
		// If not, track it and continue
		newVal := intersectorVal{
			match:       m,
			sourceMarks: make([]bool, s.numSources),
		}
		newVal.sourceMarks[source] = true
		s.matches[key] = newVal
		return nil
	}

	// If we have seen it, merge any mergeable types
	switch v := m.(type) {
	case *result.FileMatch:
		prev.match.(*result.FileMatch).AppendMatches(v)
	case *result.CommitMatch:
		prev.match.(*result.CommitMatch).AppendMatches(v)
	}

	// Mark the key as seen by this source
	prev.sourceMarks[source] = true

	// Check if the match is in the result set of all sources
	if all(prev.sourceMarks) {
		prev.sent = true
		s.matches[key] = prev
		return prev.match
	}

	s.matches[key] = prev
	return nil
}

func all(b []bool) bool {
	for _, val := range b {
		if !val {
			return false
		}
	}
	return true
}
