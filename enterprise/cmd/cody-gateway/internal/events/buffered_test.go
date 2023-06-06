package events_test

import (
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/hexops/autogold/v2"
	"github.com/sourcegraph/conc"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-gateway/internal/events"
)

func TestBufferedLogger(t *testing.T) {
	// Disabled because of race condition, see https://sourcegraph.slack.com/archives/C05497E9MDW/p1685955562994269
	// Skipping because it enables to re-enable the race condition detector and allow the team to fix this in a different PR.
	t.Skip()

	t.Parallel()
	ctx := context.Background()

	t.Run("passes events to handler", func(t *testing.T) {
		t.Parallel()

		handler := &mockLogger{}

		// Test with a buffer size of 0, which should immediately submit events
		b := events.NewBufferedLogger(logtest.Scoped(t), handler, 0)
		wg := conc.NewWaitGroup()
		wg.Go(b.Start)

		assert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "foo"}))
		assert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "bar"}))
		assert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "baz"}))

		// Stop the worker and wait for it to finish so that events flush before
		// making any assertions
		b.Stop()
		wg.Wait()

		autogold.Expect([]string{"foo", "bar", "baz"}).Equal(t, asIdentifiersList(handler.ReceivedEvents))

	})

	t.Run("buffers until full", func(t *testing.T) {
		t.Parallel()

		// blockEventSubmissionC should be closed to unblock event submission,
		// until it is all events to handler blocks indefinitely.
		blockEventSubmissionC := make(chan struct{})
		handler := &mockLogger{
			PreLogEventHook: func() { <-blockEventSubmissionC }, // block until test completion
		}

		// Set up a buffered logger we can fill up
		size := 3
		b := events.NewBufferedLogger(logtest.Scoped(t), handler, size)
		wg := conc.NewWaitGroup()
		wg.Go(b.Start)

		// Fill up the buffer with blocked events
		for i := 0; i <= size; i++ {
			assert.NoErrorf(t, b.LogEvent(ctx, events.Event{Identifier: strconv.Itoa(i)}), "event %d", i)
		}

		// Drop the next event - the queue should be full now
		err := b.LogEvent(ctx, events.Event{Identifier: "blocked"})
		require.Error(t, err)
		autogold.Expect("failed to insert event in 150ms: buffer full: 3 items pending").Equal(t, err.Error())

		// Indicate close and stop the worker so that the buffer can flush
		close(blockEventSubmissionC)
		b.Stop()
		wg.Wait()

		// All backlogged events get submitted, but the blocked event is dropped.
		autogold.Expect([]string{"0", "1", "2", "3"}).Equal(t, asIdentifiersList(handler.ReceivedEvents))
	})

	t.Run("rejects events after stop", func(t *testing.T) {
		t.Parallel()

		handler := &mockLogger{}
		l, exportLogs := logtest.Captured(t)
		b := events.NewBufferedLogger(l, handler, 10)
		wg := conc.NewWaitGroup()
		wg.Go(b.Start)

		assert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "foo"}))

		// Stop the worker and wait for it to finish
		b.Stop()
		wg.Wait()

		// Submit an additional event - this should immediately attempt to
		// submit the event
		assert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "bar"}))

		// Expect all events to still be logged
		autogold.Expect([]string{"foo", "bar"}).Equal(t, asIdentifiersList(handler.ReceivedEvents))
		// Expect log message to indicate send-after-close
		assert.True(t, exportLogs().Contains(func(l logtest.CapturedLog) bool {
			return strings.Contains(l.Message, "buffer is closed")
		}))
	})
}

type mockLogger struct {
	// PreLogEventHook, if set, is called on LogEvent before the event is added
	// to (*mockLogger).Events.
	PreLogEventHook func()
	// ReceivedEvents are all the events submitted to LogEvent. When mockLogger
	// is used as the handler for a BufferedLogger, ReceivedEvents must not be
	// accessed until (*BufferedLogger).Stop() has been called and
	// (*BufferedLogger).Start() has exited.
	ReceivedEvents []events.Event
}

func (m *mockLogger) LogEvent(spanCtx context.Context, event events.Event) error {
	if m.PreLogEventHook != nil {
		m.PreLogEventHook()
	}
	m.ReceivedEvents = append(m.ReceivedEvents, event)
	return nil
}

// asIdentifiersList renders a list of events as a list of their identifiers,
// (events.Event).Identifier, for ease of comparison.
func asIdentifiersList(sources []events.Event) []string {
	var names []string
	for _, s := range sources {
		names = append(names, s.Identifier)
	}
	return names
}
