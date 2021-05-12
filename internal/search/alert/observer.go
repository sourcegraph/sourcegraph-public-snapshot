package alert

import (
	"context"
	"sync"

	"github.com/hashicorp/go-multierror"
	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/comby"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

type Observer struct {
	// Inputs are used to generate alert messages based on the query.
	Inputs *run.SearchInputs

	// Update state.
	hasResults bool

	// Error state. Can be called concurrently.
	mu    sync.Mutex
	alert *Alert
	err   error
}

// Update AlertObserver's state based on event.
func (o *Observer) Update(event streaming.SearchEvent) {
	if len(event.Results) > 0 {
		o.hasResults = true
	}
}

func (o *Observer) Error(ctx context.Context, err error) {
	// Timeouts are reported through Stats so don't report an error for them.
	if err == nil || isContextError(ctx, err) {
		return
	}

	// We can compute the alert outside of the critical section.
	alert := FromError(err)

	o.mu.Lock()
	defer o.mu.Unlock()

	// The error can be converted into an alert.
	if alert != nil {
		o.update(alert)
		return
	}

	// Track the unexpected error for reporting when calling Done.
	o.err = multierror.Append(o.err, err)
}

// update to alert if it is more important than our current alert.
func (o *Observer) update(alert *Alert) {
	if o.alert == nil || alert.Priority > o.alert.Priority {
		o.alert = alert
	}
}

//  Done returns the highest priority alert and a multierror.Error containing
//  all errors that could not be converted to alerts.
func (o *Observer) Done(stats *streaming.Stats) (*Alert, error) {
	if !o.hasResults && o.Inputs.PatternType != query.SearchTypeStructural && comby.MatchHoleRegexp.MatchString(o.Inputs.OriginalQuery) {
		o.update(alertForStructuralSearchNotSet(o.Inputs.OriginalQuery))
	}

	if o.hasResults && o.err != nil {
		log15.Error("Errors during search", "error", o.err)
		return o.alert, nil
	}

	return o.alert, o.err
}

func alertForStructuralSearchNotSet(queryString string) *Alert {
	return &Alert{
		PrometheusType: "structural_search_not_set",
		Title:          "No results",
		Description:    "It looks like you may have meant to run a structural search, but it is not toggled.",
		ProposedQueries: []ProposedQuery{{
			Description: "Activate structural search",
			Query:       queryString,
			PatternType: query.SearchTypeStructural,
		}},
	}
}

// isContextError returns true if ctx.Err() is not nil or if err
// is an error caused by context cancelation or timeout.
func isContextError(ctx context.Context, err error) bool {
	return ctx.Err() != nil || err == context.Canceled || err == context.DeadlineExceeded
}
