package guardrails

import (
	"context"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/completions/types"

	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
)

// AttributionTest is a predicate that tells whether given snippet can be freely used.
// Returning false indicates that some attribution of snippet was found, and the snippet
// should be used with caution. The operation can be long-running, and should respect
// context cancellation.
type AttributionTest func (context.Context, string) bool

// CompletionsFilter is used to filter out the completions with potential risk
// due to attribution.
//
// Example usage:
// ```
// filter := NewCompletionsFilter(w, func (ctx context.Context, snippet string) bool {
//   // execute attribution search and return true/false.
// })
// // as completion events arrive:
// err := filter.Send(ctx, event)
// // all send operations finished:
// err := filter.WaitDone(ctx)
// ```
type CompletionsFilter struct {
	// eventWriter is the sink where completion stream events end up if they can
	// get passed along to the user.
	eventWriter func () *streamhttp.Writer
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
	// canUse is nil if attribution has not finished yet,
	// true if it finished, and result was permissive, and false otherwise
	canUse *bool
	// last is the most recent completion response. We keep the most recent
	// completion passed over to Send while attribution is running. Once attribution
	// finishes, we carry on sending from the most recent completion.
	last types.CompletionResponse
}

// NewCompletionsFilter returns a fully initialized streaming filter.
// This filter should be used for only a single code completions streaming
// since it keeps internal state. Public methods are synchronized.
func NewCompletionsFilter(w func () *streamhttp.Writer, t AttributionTest) *CompletionsFilter {
	return &CompletionsFilter{
		eventWriter: w,
		test: t,
		attributionFinished: make(chan error, 1),
	}
}

// Send is invoked each time new completion prefix arrives as the completion
// is being yielded by the LLM.
func (a *CompletionsFilter) Send(ctx context.Context, e types.CompletionResponse) error {
	if u := a.getCanUse(); u != nil && *u {
		return a.send(e)
	}
	a.setLast(e)
	if a.smallEnough(e) {
		return a.send(e)
	}
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

// getCanUse returns the attribution search result.
// It is true if attribution was empty and snippet is free to be used.
// It is false if attribution was not empty and caution should be used proceeding.
// It is nil if attribution search did not finish (or run) yet.
func (a *CompletionsFilter) getCanUse() *bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.canUse
}

// setCanUse is invoked to note the result of attribution search.
// If true, attribution is emtpy, and the snippet is safe to be used.
// If false, attribution is not empty, and the snippet should be used with caution.
func (a *CompletionsFilter) setCanUse(canUse bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.canUse = &canUse
}

// send invokes eventWriter if available. Sending is only enabled
// as long as the http stream is still being served - in other words
// call to `WaitDone` has not returned.
func (a *CompletionsFilter) send(e types.CompletionResponse) error {
	if !a.canSend() {
		return nil
	}
	if w := a.eventWriter(); w != nil {
		return w.Event("completion", e)
	}
	return nil
}

// smallEnough indicates whether snippet size is small enough not to consider
// it for attribution search. At this point we run attribution search for
// snippets 10 lines long or longer.
func (a *CompletionsFilter) smallEnough(e types.CompletionResponse) bool {
	return len(strings.Split(e.Completion, "\n")) < 10
}

// getLast returns the last completion event to be fed to `Send`.
func (a *CompletionsFilter) getLast() types.CompletionResponse {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.last
}

// setLast overwrites the last completion event passed to `Send`.
func (a *CompletionsFilter) setLast(e types.CompletionResponse) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.last = e
}

// runAttribution performs attribution search and updates the state of the object
// after finishing. It's run within `attributionRun` to ensure it only runs once.
func (a *CompletionsFilter) runAttribution(ctx context.Context, e types.CompletionResponse) {
	canUse := a.test(ctx, e.Completion)
	a.setCanUse(canUse)
	err := a.send(a.getLast())
	a.attributionFinished <- err
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
