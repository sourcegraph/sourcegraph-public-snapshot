package ui

import (
	"bytes"
	"context"
	"time"

	"github.com/derision-test/glock"
)

// IntervalWriter is a io.Writer that flushes to the given sink on the given
// interval.
type IntervalWriter struct {
	sink func(string)

	ticker glock.Ticker

	// buf is used to keep partial lines buffered before flushing them (either
	// on the next newline or after tickDuration)
	buf        *bytes.Buffer
	writes     chan []byte
	writesDone chan struct{}

	closed chan struct{}
	done   chan struct{}
}

func newIntervalWriter(ctx context.Context, ticker glock.Ticker, sink func(string)) *IntervalWriter {
	l := &IntervalWriter{
		sink:   sink,
		ticker: ticker,

		writes:     make(chan []byte),
		writesDone: make(chan struct{}),

		buf: &bytes.Buffer{},

		closed: make(chan struct{}, 1),
		done:   make(chan struct{}, 1),
	}

	go l.writeLines(ctx)

	return l
}

// NewLogger returns a new Logger instance and spawns a goroutine in the
// background that regularily flushed the logged output to the given sink.
//
// If the passed in ctx is canceled the goroutine will exit.
func NewIntervalWriter(ctx context.Context, interval time.Duration, sink func(string)) *IntervalWriter {
	return newIntervalWriter(ctx, glock.NewRealTicker(interval), sink)
}

func (l *IntervalWriter) flush() {
	if l.buf.Len() == 0 {
		return
	}
	l.sink(l.buf.String())
	l.buf.Reset()
}

// Close flushes the
func (l *IntervalWriter) Close() error {
	l.closed <- struct{}{}
	<-l.done
	return nil
}

// Write handler of IntervalWriter.
func (l *IntervalWriter) Write(p []byte) (int, error) {
	l.writes <- p
	<-l.writesDone
	return len(p), nil
}

func (l *IntervalWriter) writeLines(ctx context.Context) {
	defer func() {
		l.flush()
		l.ticker.Stop()
		l.done <- struct{}{}
	}()

	for {
		select {
		case <-ctx.Done():
			return

		case <-l.closed:
			return

		case w, ok := <-l.writes:
			if !ok {
				return
			}

			if _, err := l.buf.Write(w); err != nil {
				break
			}
			l.writesDone <- struct{}{}
		case <-l.ticker.Chan():
			l.flush()
		}
	}
}
