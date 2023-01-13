package ui

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/derision-test/glock"
)

// IntervalProcessWriter accepts stdout/stderr writes from processes, prefixed
// them accordingly, and flushes to the given sink on the given interval.
type IntervalProcessWriter struct {
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

func newIntervalProcessWriter(ctx context.Context, ticker glock.Ticker, sink func(string)) *IntervalProcessWriter {
	l := &IntervalProcessWriter{
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

// NewIntervalProcessWriter returns a new IntervalProcessWriter instance and
// spawns a goroutine in the background that regularily flushed the logged
// output to the given sink.
//
// If the passed in ctx is canceled the goroutine will exit.
func NewIntervalProcessWriter(ctx context.Context, interval time.Duration, sink func(string)) *IntervalProcessWriter {
	return newIntervalProcessWriter(ctx, glock.NewRealTicker(interval), sink)
}

// StdoutWriter returns an io.Writer that prefixes every line with "stdout: "
func (l *IntervalProcessWriter) StdoutWriter() io.Writer {
	return &prefixedWriter{writes: l.writes, writesDone: l.writesDone, prefix: "stdout: "}
}

// SterrWriter returns an io.Writer that prefixes every line with "stderr: "
func (l *IntervalProcessWriter) StderrWriter() io.Writer {
	return &prefixedWriter{writes: l.writes, writesDone: l.writesDone, prefix: "stderr: "}
}

// Close blocks until all pending writes have been flushed to the buffer. It
// then causes the underlying goroutine to exit.
func (l *IntervalProcessWriter) Close() error {
	l.closed <- struct{}{}
	<-l.done
	return nil
}

func (l *IntervalProcessWriter) flush() {
	if l.buf.Len() == 0 {
		return
	}
	l.sink(l.buf.String())
	l.buf.Reset()
}

func (l *IntervalProcessWriter) writeLines(ctx context.Context) {
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

type prefixedWriter struct {
	writes     chan []byte
	writesDone chan struct{}
	prefix     string
}

var newLineByteSlice = []byte("\n")

// Write is only ever called with a single line. That line may or may not end with a newline character.
// It then writes the content as a single line to the inner writer, regardless of if the provided line
// had a newline or not. That is, because our encoding requires exactly one line per formatted line.
func (w *prefixedWriter) Write(p []byte) (int, error) {
	prefixedLine := append([]byte(w.prefix), p...)
	if !bytes.HasSuffix(prefixedLine, newLineByteSlice) {
		prefixedLine = append(prefixedLine, newLineByteSlice...)
	}
	w.writes <- prefixedLine
	<-w.writesDone
	return len(p), nil
}
