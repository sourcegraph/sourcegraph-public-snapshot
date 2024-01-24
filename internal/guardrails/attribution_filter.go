package guardrails

import (
	"context"
	"errors"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"
)

// AttributionTest is a predicate that tells whether given snippet can be freely used.
// Returning false indicates that some attribution of snippet was found, and the snippet
// should be used with caution. The operation can be long-running, and should respect
// context cancellation.
type AttributionTest func (context.Context, string) bool

// CompletionEventSink is where selected events end up being streamed to.
type CompletionEventSink func (types.CompletionResponse) error

// CompletionsFilter is used to filter out the completions with potential risk
// due to attribution.
//
// Example usage:
// ```
// sink := func (e types.CompletionResponse) error {
//   // stream completions response back to the client
// }
// test := func (ctx context.Context, snippet string) bool {
//   // execute attribution search and return true/false.
// }
// filter := NewCompletionsFilter(sink, test)
// // as completion events arrive:
// err := filter.Send(ctx, event)
// // all send operations finished:
// err := filter.WaitDone(ctx)
// ```
type CompletionsFilter struct {
	// sink where completion stream events end up if they can
	// get passed along to the user.
	sink CompletionEventSink
	// filter exposes a long-running attribution operation
	test AttributionTest
	// attributionRun synchronizes on attribution to run only once.
	attributionRun sync.Once
	// attributionFinished has to have len=1 and stores error from attribution
	// run once it's finished.
	attributionFinished chan error
	// mu guards all the following fields
	mu sync.Mutex
	// sendingBlocked = true indicates that eventWriter should not be used anymore
	sendingBlocked bool
	// attributionResult is nil if attribution search has not finished yet,
	// true if it finished, and result was permissive, and false otherwise
	attributionResult *bool
	// mostRecentCompletion memoized. We keep the most recent
	// completion passed over to Send while attribution is running. Once attribution
	// finishes, we carry on sending from the most recent completion.
	mostRecentCompletion types.CompletionResponse
}

// NewCompletionsFilter returns a fully initialized streaming filter.
// This filter should be used for only a single code completions streaming
// since it keeps internal state. Public methods are synchronized.
func NewCompletionsFilter(sink CompletionEventSink, test AttributionTest) *CompletionsFilter {
	return &CompletionsFilter{
		sink: sink,
		test: test,
		attributionFinished: make(chan error, 1),
	}
}

// Send is invoked each time new completion prefix arrives as the completion
// is being yielded by the LLM.
func (a *CompletionsFilter) Send(ctx context.Context, e types.CompletionResponse) error {
	if err := ctx.Err(); err != nil && errors.Is(err, context.Canceled) {
		a.blockSending()
	}
	if a.attributionResultPermissive() {
		return a.send(e)
	}
	if a.smallEnough(e) {
		return a.send(e)
	}
	a.setMostRecentCompletion(e)
	a.attributionRun.Do(func () {
		go a.runAttribution(ctx, e)
	})
	return nil
}

// WaitDone is called after all the completions have finished arriving
// in order to await on async attribution within time limits.
// In other words all calls to `Send` will have finished. Then `WaitDone`
// is called and it will wait for remaining attribution search if context
// time limits permit.
func (a *CompletionsFilter) WaitDone(ctx context.Context) error {
	select {
	case <-ctx.Done():
		a.blockSending()
		return nil
	case err := <-a.attributionFinished:
		return err
	}
}

// attributionResultPermissive states whether attribution search
// finished and is permissive - that is no attribution was found.
func (a *CompletionsFilter) attributionResultPermissive() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.attributionResult != nil && *a.attributionResult
}

// setAttributionResult is invoked to note the result of attribution search.
// If true, attribution is emtpy, and the snippet is safe to be used.
// If false, attribution is not empty, and the snippet should be used with caution.
func (a *CompletionsFilter) setAttributionResult(attributionResult bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.attributionResult = &attributionResult
}

// send invokes eventWriter if available. Sending is only enabled
// as long as the http stream is still being served - in other words
// call to `WaitDone` has not returned.
func (a *CompletionsFilter) send(e types.CompletionResponse) error {
	if !a.canSend() {
		return nil
	}
	return a.sink(e)
}

// smallEnough indicates whether snippet size is small enough not to consider
// it for attribution search. At this point we run attribution search for
// snippets 10 lines long or longer.
func (a *CompletionsFilter) smallEnough(e types.CompletionResponse) bool {
	return len(strings.Split(e.Completion, "\n")) < 10
}

// getMostRecentCompletion returns the last completion event to be fed to `Send`.
func (a *CompletionsFilter) getMostRecentCompletion() types.CompletionResponse {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.mostRecentCompletion
}

// setMostRecentCompletion overwrites the last completion event passed to `Send`.
func (a *CompletionsFilter) setMostRecentCompletion(e types.CompletionResponse) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.mostRecentCompletion = e
}

// runAttribution performs attribution search and updates the state of the object
// after finishing. It's run within `attributionRun` to ensure it only runs once.
func (a *CompletionsFilter) runAttribution(ctx context.Context, e types.CompletionResponse) {
	result := a.test(ctx, e.Completion)
	a.setAttributionResult(result)
	if result {
		err := a.send(a.getMostRecentCompletion())
		a.attributionFinished <- err
	}
	close(a.attributionFinished)
}

// blockSending prevents any operations on the event writer.
func (a *CompletionsFilter) blockSending() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.sendingBlocked = true
}

// canSend states whether event writer can be used.
func (a *CompletionsFilter) canSend() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return !a.sendingBlocked
}
