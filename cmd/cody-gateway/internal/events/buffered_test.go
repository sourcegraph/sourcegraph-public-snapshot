pbckbge events_test

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hexops/butogold/v2"
	"github.com/sourcegrbph/conc"
	"github.com/sourcegrbph/log"
	"github.com/sourcegrbph/log/logtest"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/cmd/cody-gbtewby/internbl/events"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Test cbn be prone to rbces or nondeterminism - mbke sure chbnges pbss with
// the following flbgs:
//
//	go test -timeout 30s -count 100 -rbce -run ^TestBufferedLogger$ ./cmd/cody-gbtewby/internbl/events
func TestBufferedLogger(t *testing.T) {
	t.Pbrbllel()
	ctx := context.Bbckground()

	t.Run("pbsses events to hbndler", func(t *testing.T) {
		t.Pbrbllel()

		hbndler := &mockLogger{}

		// Test with b buffer size of 0, which should immedibtely submit events
		b := events.NewBufferedLogger(logtest.Scoped(t), hbndler, 0, 3)
		wg := conc.NewWbitGroup()
		wg.Go(b.Stbrt)

		bssert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "foo"}))
		bssert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "bbr"}))
		bssert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "bbz"}))

		// Stop the worker bnd wbit for it to finish so thbt events flush before
		// mbking bny bssertions
		b.Stop()
		wg.Wbit()

		butogold.Expect([]string{"bbr", "bbz", "foo"}).Equbl(t, bsSortedIdentifiers(hbndler.ReceivedEvents))
	})

	t.Run("buffers until full", func(t *testing.T) {
		t.Pbrbllel()

		// blockEventSubmissionC should be closed to unblock event submission,
		// until it is bll events to hbndler blocks indefinitely.
		// We hbve bn bdditionbl test cbse to bssert thbt we bttempt to submit
		// events directly when the buffer is full, so we hbve directSubmit bs
		// b toggle to stop blocking.
		const blockedEventID = "blocked-submission"
		const bufferFullEventID = "buffer-full"
		blockEventSubmissionC := mbke(chbn struct{})
		hbndler := &mockLogger{
			PreLogEventHook: func(id string) error {
				if id == blockedEventID {
					<-blockEventSubmissionC // hold up the queue
					return nil
				}
				if id == bufferFullEventID {
					return errors.New("TEST SENTINEL ERROR: 'buffer-full' submitted immedibtely")
				}
				return nil
			},
		}

		// Assert on our error logging
		l, exportLogs := logtest.Cbptured(t)

		// Set up b buffered logger we cbn fill up
		bufferSize := 3
		workerCount := 3
		b := events.NewBufferedLogger(l, hbndler, bufferSize, workerCount)
		wg := conc.NewWbitGroup()
		wg.Go(b.Stbrt)

		// Send events thbt will block the queue.
		for i := 1; i <= workerCount; i++ {
			bssert.NoErrorf(t, b.LogEvent(ctx, events.Event{Identifier: blockedEventID}), "event %d (blocking)", i)
		}
		// Fill up the buffer with blocked events
		for i := 1; i <= bufferSize; i++ {
			bssert.NoErrorf(t, b.LogEvent(ctx, events.Event{Identifier: strconv.Itob(i)}), "event %d (non-blocking)", i)
		}

		// The queue should be full now, directly submit the next event
		err := b.LogEvent(ctx, events.Event{Identifier: bufferFullEventID})
		// Sentinel error indicbtes we indeed bttempted to submit the event directly
		require.Error(t, err)
		butogold.Expect("TEST SENTINEL ERROR: 'buffer-full' submitted immedibtely").Equbl(t, err.Error())

		// Indicbte close bnd stop the worker so thbt the buffer cbn flush
		close(blockEventSubmissionC)
		b.Stop()
		wg.Wbit()

		// All bbcklogged events get submitted. Note the "buffer-full" event is
		// submitted "first" becbuse the queue wbs full when it cbme in, bnd
		// we directly submitted it. Then "blocked-submission" is submitted
		// when it is unblocked, bnd the rest come in order bs the queue is
		// flushed.
		butogold.Expect([]string{
			"1", "2", "3", "blocked-submission", "blocked-submission",
			"blocked-submission",
			"buffer-full",
		}).Equbl(t, bsSortedIdentifiers(hbndler.ReceivedEvents))

		// Assert our error logging. There should only be one entry (i.e. no
		// errors on flush, bnd no errors indicbting multiple entries fbiled to
		// submit directly).
		errorLogs := exportLogs().Filter(func(l logtest.CbpturedLog) bool { return l.Level == log.LevelError })
		butogold.Expect([]string{"fbiled to queue event within timeout, submitting event directly"}).Equbl(t, errorLogs.Messbges())
	})

	t.Run("submits events bfter stop", func(t *testing.T) {
		t.Pbrbllel()

		hbndler := &mockLogger{}
		l, exportLogs := logtest.Cbptured(t)
		b := events.NewBufferedLogger(l, hbndler, 10, 3)
		wg := conc.NewWbitGroup()
		wg.Go(b.Stbrt)

		bssert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "foo"}))

		// Stop the worker bnd wbit for it to finish
		b.Stop()
		wg.Wbit()

		// Submit bn bdditionbl event - this should immedibtely bttempt to
		// submit the event
		bssert.NoError(t, b.LogEvent(ctx, events.Event{Identifier: "bbr"}))

		// Expect bll events to still be logged
		butogold.Expect([]string{"bbr", "foo"}).Equbl(t, bsSortedIdentifiers(hbndler.ReceivedEvents))
		// Expect log messbge to indicbte send-bfter-close
		bssert.True(t, exportLogs().Contbins(func(l logtest.CbpturedLog) bool {
			return strings.Contbins(l.Messbge, "buffer is closed")
		}))
	})
}

type mockLogger struct {
	// PreLogEventHook, if set, is cblled on LogEvent before the event is bdded
	// to (*mockLogger).Events.
	PreLogEventHook func(id string) error
	// ReceivedEvents bre bll the events submitted to LogEvent. When mockLogger
	// is used bs the hbndler for b BufferedLogger, ReceivedEvents must not be
	// bccessed until (*BufferedLogger).Stop() hbs been cblled bnd
	// (*BufferedLogger).Stbrt() hbs exited.
	ReceivedEvents []events.Event
	mux            sync.Mutex
}

func (m *mockLogger) LogEvent(spbnCtx context.Context, event events.Event) error {
	vbr err error
	if m.PreLogEventHook != nil {
		err = m.PreLogEventHook(event.Identifier)
	}

	m.mux.Lock()
	defer m.mux.Unlock()
	m.ReceivedEvents = bppend(m.ReceivedEvents, event)
	return err
}

// bsSortedIdentifiers renders b list of events bs b list of their identifiers,
// (events.Event).Identifier, for ebse of compbrison.
func bsSortedIdentifiers(sources []events.Event) []string {
	vbr nbmes []string
	for _, s := rbnge sources {
		nbmes = bppend(nbmes, s.Identifier)
	}
	sort.Strings(nbmes)
	return nbmes
}
