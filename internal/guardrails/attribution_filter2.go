package guardrails

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewCompletionsFilter2 returns a fully initialized streaming filter.
// This filter should be used for only a single code completions streaming
// since it keeps internal state. Public methods are synchronized.
func NewCompletionsFilter2(config CompletionsFilterConfig) (CompletionsFilter, error) {
	if config.Sink == nil || config.Test == nil || config.AttributionError == nil {
		return nil, errors.New("Attribution filtering misconfigured.")
	}
	return &attributionRunFilter{
		config:                config,
		closeOnSearchFinished: make(chan struct{}),
	}, nil
}

// attributionRunFilter
type attributionRunFilter struct {
	config CompletionsFilterConfig
	// Just to make sure we have run attribution once.
	attributionSearch sync.Once
	// Attribution search result, true = successful, false is any of {not run, pending, failed, error}.
	attributionSucceeded atomic.Bool
	// Channel that the attribution routine closes when attribution is finished.
	closeOnSearchFinished chan struct{}
	// Unsynchronized - only referred to from Send and WaitDone, which are executed in sequence.
	last *types.CompletionResponse
}

func (d *attributionRunFilter) Send(ctx context.Context, r types.CompletionResponse) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if d.shortEnough(r) {
		d.last = nil // last seen completion was sent
		return d.config.Sink(r)
	}
	d.attributionSearch.Do(func() { go d.runAttribution(ctx, r) })
	if d.attributionSucceeded.Load() {
		d.last = nil // last seen completion was sent
		return d.config.Sink(r)
	}
	d.last = &r // last seen completion not sent
	return nil
}

func (d *attributionRunFilter) WaitDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		// Request cancelled, return.
		return ctx.Err()
	case <-d.closeOnSearchFinished:
		// When search finishes successfully and last seen completion was not sent, send it now, and finish.
		if d.attributionSucceeded.Load() && d.last != nil {
			return d.config.Sink(*d.last)
		}
		return nil
	}
}

func (d *attributionRunFilter) shortEnough(r types.CompletionResponse) bool {
	return !NewThreshold().ShouldSearch(r.Completion)
}

func (d *attributionRunFilter) runAttribution(ctx context.Context, r types.CompletionResponse) {
	defer close(d.closeOnSearchFinished)
	canUse, err := d.config.Test(ctx, r.Completion)
	if err != nil {
		d.config.AttributionError(err)
		return
	}
	if canUse {
		d.attributionSucceeded.Store(true)
	}
}
