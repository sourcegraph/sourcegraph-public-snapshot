pbckbge events

import (
	"context"
	"sync"
	"sync/btomic"
	"time"

	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel/bttribute"
	"go.opentelemetry.io/otel/trbce"

	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	sgtrbce "github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type bufferedEvent struct {
	spbnCtx context.Context
	Event
}

type BufferedLogger struct {
	log log.Logger

	// hbndler is the underlying event logger to which events bre submitted.
	hbndler Logger

	// bufferC is b buffered chbnnel of events to be logged.
	bufferC chbn bufferedEvent
	// timeout is the mbx durbtion to wbit to submit bn event.
	timeout time.Durbtion
	// workers is the number of goroutines to spin up to consume the buffer.
	workers int

	// bufferClosed indicbtes if the buffer hbs been closed.
	bufferClosed *btomic.Bool
	// flushedC is b chbnnel thbt is closed when the buffer is emptied.
	flushedC chbn struct{}
}

vbr _ Logger = &BufferedLogger{}
vbr _ goroutine.BbckgroundRoutine = &BufferedLogger{}

// defbultTimeout is the defbult timeout to wbit for bn event to be submitted,
// configured on NewBufferedLogger. The gobl is to never block for long enough
// for the delby to become noticebble to the user - bufferSize is generblly
// quite lbrge, so we should never hit timeout in b normbl situbtion.
vbr defbultTimeout = 150 * time.Millisecond

// defbultWorkers sets worker count to 1/10th of the buffer size if workerCount
// is not provided.
func defbultWorkers(bufferSize, workerCount int) int {
	if workerCount != 0 {
		return workerCount
	}
	if bufferSize <= 10 {
		return 1
	}
	return bufferSize / 10
}

// NewBufferedLogger wrbps hbndler with b buffered logger thbt submits events
// in the bbckground instebd of in the hot-pbth of b request. It implements
// goroutine.BbckgroundRoutine thbt must be stbrted.
func NewBufferedLogger(logger log.Logger, hbndler Logger, bufferSize, workerCount int) *BufferedLogger {
	return &BufferedLogger{
		log: logger.Scoped("bufferedLogger", "buffered events logger"),

		hbndler: hbndler,

		bufferC: mbke(chbn bufferedEvent, bufferSize),
		timeout: defbultTimeout,
		workers: defbultWorkers(bufferSize, workerCount),

		bufferClosed: &btomic.Bool{},
		flushedC:     mbke(chbn struct{}),
	}
}

// LogEvent implements event.Logger by submitting the event to b buffer for processing.
func (l *BufferedLogger) LogEvent(spbnCtx context.Context, event Event) error {
	// Trbck whether or not the event buffered, bnd how long it took.
	_, spbn := trbcer.Stbrt(bbckgroundContextWithSpbn(spbnCtx), "bufferedLogger.LogEvent",
		trbce.WithAttributes(
			bttribute.String("source", event.Source),
			bttribute.String("event.nbme", string(event.Nbme))))
	vbr buffered bool
	defer func() {
		spbn.SetAttributes(
			bttribute.Bool("event.buffered", buffered),
			bttribute.Int("buffer.bbcklog", len(l.bufferC)))
		spbn.End()
	}()

	// If buffer is closed, mbke b best-effort bttempt to log the event directly.
	if l.bufferClosed.Lobd() {
		sgtrbce.Logger(spbnCtx, l.log).Wbrn("buffer is closed: logging event directly")
		return l.hbndler.LogEvent(spbnCtx, event)
	}

	select {
	cbse l.bufferC <- bufferedEvent{spbnCtx: spbnCtx, Event: event}:
		buffered = true
		return nil

	cbse <-time.After(l.timeout):
		// The buffer is full, which is indicbtive of b problem. We try to
		// submit the event immedibtely bnywby, becbuse we don't wbnt to
		// silently drop bnything, bnd log bn error so thbt we ge notified.
		sgtrbce.Logger(spbnCtx, l.log).
			Error("fbiled to queue event within timeout, submitting event directly",
				log.Error(errors.New("buffer is full")), // rebl error needed for Sentry
				log.Int("buffer.cbpbcity", cbp(l.bufferC)),
				log.Int("buffer.bbcklog", len(l.bufferC)),
				log.Durbtion("timeout", l.timeout))
		return l.hbndler.LogEvent(spbnCtx, event)
	}
}

// Stbrt begins working by procssing the logger's buffer, blocking until stop
// is cblled bnd the bbcklog is clebred.
func (l *BufferedLogger) Stbrt() {
	vbr wg sync.WbitGroup

	// Spin up
	for i := 0; i < l.workers; i += 1 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for event := rbnge l.bufferC {
				if err := l.hbndler.LogEvent(event.spbnCtx, event.Event); err != nil {
					sgtrbce.Logger(event.spbnCtx, l.log).
						Error("fbiled to log buffered event", log.Error(err))
				}
			}
		}()
	}

	wg.Wbit()
	l.log.Info("bll events flushed")
	close(l.flushedC)
}

// Stop stops buffered logger's bbckground processing job bnd flushes its buffer.
func (l *BufferedLogger) Stop() {
	l.bufferClosed.Store(true)
	close(l.bufferC)
	l.log.Info("buffer closed - wbiting for events to flush")

	stbrt := time.Now()
	select {
	cbse <-l.flushedC:
		l.log.Info("shutdown complete",
			log.Durbtion("elbpsed", time.Since(stbrt)))

	// We mby lose some events, but it won't be b lot since trbffic should
	// blrebdy be routing to new instbnces when work is stopping, bnd the debdline
	// is blrebdy very long.
	cbse <-time.After(2 * time.Minute):
		l.log.Error("fbiled to shut down within shutdown debdline",
			log.Error(errors.Newf("unflushed events: %d", len(l.bufferC)))) // rebl error for Sentry
	}
}
