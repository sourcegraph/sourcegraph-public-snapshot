package streaming

import (
	"context"
	"sync"
	"time"

	"go.uber.org/atomic"

	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
)

type SearchEvent struct {
	Results result.Matches
	Stats   Stats
}

type Sender interface {
	Send(SearchEvent)
}

type LimitStream struct {
	s         Sender
	cancel    context.CancelFunc
	remaining atomic.Int64
}

func (s *LimitStream) Send(event SearchEvent) {
	count := int64(event.Results.ResultCount())

	// Avoid limit checks if no change to result count.
	if count == 0 {
		s.s.Send(event)
		return
	}

	// Get the remaining count before and after sending this event
	after := s.remaining.Sub(count)
	before := after + count

	// Check if the event needs truncating before being sent
	if after < 0 {
		limit := before
		if before < 0 {
			limit = 0
		}
		event.Results.Limit(int(limit))
	}

	// Send the maybe-truncated event. We want to always send the event
	// even if we truncate it to zero results in case it has stats on it
	// that we care about it.
	s.s.Send(event)

	// Send the IsLimitHit event and call cancel exactly once. This will
	// only trigger when the result count of an event causes us to cross
	// the zero-remaining threshold.
	if after <= 0 && before > 0 {
		s.s.Send(SearchEvent{Stats: Stats{IsLimitHit: true}})
		s.cancel()
	}
}

// WithLimit returns a child Stream of parent as well as a child Context of
// ctx. The child stream passes on all events to parent. Once more than limit
// ResultCount are sent on the child stream the context is canceled and an
// IsLimitHit event is sent.
//
// Canceling this context releases resources associated with it, so code
// should call cancel as soon as the operations running in this Context and
// Stream are complete.
func WithLimit(ctx context.Context, parent Sender, limit int) (context.Context, Sender, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	stream := &LimitStream{cancel: cancel, s: parent}
	stream.remaining.Store(int64(limit))
	return ctx, stream, cancel
}

// WithSelect returns a child Stream of parent that runs the select operation
// on each event, deduplicating where possible.
func WithSelect(parent Sender, s filter.SelectPath) Sender {
	var mux sync.Mutex
	dedup := result.NewDeduper()

	return StreamFunc(func(e SearchEvent) {
		if parent == nil {
			return
		}
		mux.Lock()

		selected := e.Results[:0]
		for _, match := range e.Results {
			current := match.Select(s)
			if current == nil {
				continue
			}

			// If the selected file is a file match send it unconditionally
			// to ensure we get all line matches for a file. One exception:
			// if we are only interested in the path (via `select:file`),
			// we only send the result once.
			seen := dedup.Seen(current)
			fm, isFileMatch := current.(*result.FileMatch)
			if seen && !isFileMatch {
				continue
			}
			if seen && isFileMatch && fm.IsPathMatch() {
				continue
			}

			dedup.Add(current)
			selected = append(selected, current)
		}
		e.Results = selected

		mux.Unlock()
		parent.Send(e)
	})
}

type StreamFunc func(SearchEvent)

func (f StreamFunc) Send(se SearchEvent) {
	f(se)
}

// NewAggregatingStream returns a stream that collects all the events
// sent to it. The aggregated event can be retrieved with Get().
func NewAggregatingStream() *aggregatingStream {
	return &aggregatingStream{}
}

type aggregatingStream struct {
	sync.Mutex
	SearchEvent
}

func (c *aggregatingStream) Send(event SearchEvent) {
	c.Lock()
	c.Results = append(c.Results, event.Results...)
	c.Stats.Update(&event.Stats)
	c.Unlock()
}

func NewNullStream() Sender {
	return StreamFunc(func(SearchEvent) {})
}

func NewStatsObservingStream(s Sender) *statsObservingStream {
	return &statsObservingStream{
		parent: s,
	}
}

type statsObservingStream struct {
	parent Sender

	sync.Mutex
	Stats
}

func (s *statsObservingStream) Send(event SearchEvent) {
	s.Lock()
	s.Stats.Update(&event.Stats)
	s.Unlock()
	s.parent.Send(event)
}

func NewResultCountingStream(s Sender) *resultCountingStream {
	return &resultCountingStream{
		parent: s,
	}
}

type resultCountingStream struct {
	parent Sender
	count  atomic.Int64
}

func (c *resultCountingStream) Send(event SearchEvent) {
	c.count.Add(int64(event.Results.ResultCount()))
	c.parent.Send(event)
}

func (c *resultCountingStream) Count() int {
	return int(c.count.Load())
}

// NewBatchingStream returns a stream that batches results sent to it, holding
// delaying their forwarding by a max time of maxDelay, then sending the batched
// event to the parent stream. The first event is passed through without delay.
// When there will be no more events sent on the batching stream, Done() must be
// called to flush the remaining batched events.
func NewBatchingStream(maxDelay time.Duration, parent Sender) *batchingStream {
	return &batchingStream{
		parent:   parent,
		maxDelay: maxDelay,
	}
}

type batchingStream struct {
	parent   Sender
	maxDelay time.Duration

	mu             sync.Mutex
	sentFirstEvent bool
	dirty          bool
	batch          SearchEvent
	timer          *time.Timer
	flushScheduled bool
}

func (s *batchingStream) Send(event SearchEvent) {
	s.mu.Lock()

	// Update the batch
	s.batch.Results = append(s.batch.Results, event.Results...)
	s.batch.Stats.Update(&event.Stats)
	s.dirty = true

	// If this is our first event with results, flush immediately
	if !s.sentFirstEvent && len(event.Results) > 0 {
		s.sentFirstEvent = true
		s.flush()
		s.mu.Unlock()
		return
	}

	if s.timer == nil {
		// Create a timer and schedule a flush
		s.timer = time.AfterFunc(s.maxDelay, func() {
			s.mu.Lock()
			s.flush()
			s.flushScheduled = false
			s.mu.Unlock()
		})
		s.flushScheduled = true
	} else if !s.flushScheduled {
		// Reuse the timer, scheduling a new flush
		s.timer.Reset(s.maxDelay)
		s.flushScheduled = true
	}
	// If neither of those conditions is true,
	// a flush has already been scheduled and
	// we're good to go.
	s.mu.Unlock()
}

// Done should be called when no more events will be sent down
// the stream. It flushes any events that are currently batched
// and cancels any scheduled flush.
func (s *batchingStream) Done() {
	s.mu.Lock()
	// Cancel any scheduled flush
	if s.timer != nil {
		s.timer.Stop()
	}

	s.flush()
	s.mu.Unlock()
}

// flush sends the currently batched events to the parent stream. The caller must hold
// a lock on the batching stream.
func (s *batchingStream) flush() {
	if s.dirty {
		s.parent.Send(s.batch)
		s.batch = SearchEvent{}
		s.dirty = false
	}
}
