package process

import (
	"bytes"
	"context"
	"time"
)

// tickDuration is the time to wait before writing the buffer contents
// without having received a newline.
var tickDuration = 2 * time.Millisecond

// Logger is a simplified version of goreman's logger:
// https://github.com/mattn/goreman/blob/master/log.go
type Logger struct {
	sink func(string)

	// buf is used to keep partial lines buffered before flushing them (either
	// on the next newline or after tickDuration)
	buf    *bytes.Buffer
	writes chan []byte
	done   chan struct{}
}

// NewLogger returns a new Logger instance and spawns a goroutine in the
// background that regularily flushed the logged output to the given sink.
//
// If the passed in ctx is canceled the goroutine will exit.
func NewLogger(ctx context.Context, sink func(string)) *Logger {
	l := &Logger{
		sink: sink,

		writes: make(chan []byte),
		done:   make(chan struct{}),
		buf:    &bytes.Buffer{},
	}

	go l.writeLines(ctx)

	return l
}

func (l *Logger) bufLine(line []byte) error {
	_, err := l.buf.Write(line)
	return err
}

func (l *Logger) flush() {
	if l.buf.Len() == 0 {
		return
	}
	l.sink(l.buf.String())
	l.buf.Reset()
}

// Write handler of logger.
func (l *Logger) Write(p []byte) (int, error) {
	l.writes <- p
	<-l.done
	return len(p), nil
}

func (l *Logger) writeLines(ctx context.Context) {
	tick := time.NewTicker(tickDuration)
	for {
		select {
		case <-ctx.Done():
			l.flush()
			return
		case w, ok := <-l.writes:
			if !ok {
				l.flush()
				return
			}

			buf := bytes.NewBuffer(w)
			for {
				line, err := buf.ReadBytes('\n')
				if len(line) > 0 {
					if line[len(line)-1] == '\n' {
						// TODO: We currently add a newline in flush(), see comment there
						line = line[0 : len(line)-1]

						// But since there *was* a newline, we need to flush,
						// but only if there is more than a newline or there
						// was already content.
						if len(line) != 0 || l.buf.Len() > 0 {
							if err := l.bufLine(line); err != nil {
								break
							}
							l.flush()
						}
						tick.Stop()
					} else {
						if err := l.bufLine(line); err != nil {
							break
						}
						tick.Reset(tickDuration)
					}
				}
				if err != nil {
					break
				}
			}
			l.done <- struct{}{}
		case <-tick.C:
			l.flush()
			tick.Stop()
		}
	}
}
