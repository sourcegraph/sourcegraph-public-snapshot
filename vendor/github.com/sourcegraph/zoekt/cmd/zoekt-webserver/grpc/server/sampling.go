package server

import (
	"math"

	"github.com/sourcegraph/zoekt"
)

// newSamplingSender is a zoekt.Sender that samples stats events to avoid
// sending many empty stats events over the wire.
func newSamplingSender(next zoekt.Sender) *samplingSender {
	return &samplingSender{next: next}
}

type samplingSender struct {
	next     zoekt.Sender
	agg      zoekt.SearchResult
	aggCount int
}

func (s *samplingSender) Send(event *zoekt.SearchResult) {
	// We don't want to send events over the wire if they don't contain file
	// matches. Hence, in case we didn't find any results, we aggregate the stats
	// and send them out in regular intervals.
	if len(event.Files) == 0 {
		s.aggCount++

		s.agg.Stats.Add(event.Stats)
		s.agg.Progress = event.Progress

		if s.aggCount%100 == 0 && !s.agg.Stats.Zero() {
			s.next.Send(&s.agg)
			s.agg = zoekt.SearchResult{}
		}

		return
	}

	// If we have aggregate stats, we merge them with the new event before sending
	// it. We drop agg.Progress, because we assume that event.Progress reflects the
	// latest status.
	if !s.agg.Stats.Zero() {
		event.Stats.Add(s.agg.Stats)
		s.agg = zoekt.SearchResult{}
	}

	s.next.Send(event)
}

// Flush sends any aggregated stats that we haven't sent yet
func (s *samplingSender) Flush() {
	if !s.agg.Stats.Zero() {
		s.next.Send(&zoekt.SearchResult{
			Stats: s.agg.Stats,
			Progress: zoekt.Progress{
				Priority:           math.Inf(-1),
				MaxPendingPriority: math.Inf(-1),
			},
		})
	}
}
