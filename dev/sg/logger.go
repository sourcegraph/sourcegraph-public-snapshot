package main

import (
	"bytes"
	"io"
	"time"

	"github.com/sourcegraph/batch-change-utils/output"
)

type buffers [][]byte

func (v *buffers) consume(n int64) {
	for len(*v) > 0 {
		ln0 := int64(len((*v)[0]))
		if ln0 > n {
			(*v)[0] = (*v)[0][n:]
			return
		}
		n -= ln0
		*v = (*v)[1:]
	}
}

func (v *buffers) WriteTo(w io.Writer) (n int64, err error) {
	for _, b := range *v {
		nb, err := w.Write(b)
		n += int64(nb)
		if err != nil {
			v.consume(n)
			return n, err
		}
	}
	v.consume(n)
	return n, nil
}

// tickDuration is the time to wait before writing the buffer contents
// without having received a newline.
var tickDuration time.Duration = 2 * time.Millisecond

// cmdLogger is a simplified version of goreman's logger:
// https://github.com/mattn/goreman/blob/master/log.go
type cmdLogger struct {
	out *output.Output

	name   string
	writes chan []byte
	done   chan struct{}

	buffers buffers
}

// newCmdLogger returns a new cmdLogger instance and spawns a goroutine in the
// background that regularily flushed the logged output to the given output..
func newCmdLogger(name string, out *output.Output) *cmdLogger {
	l := &cmdLogger{
		name:   name,
		out:    out,
		writes: make(chan []byte),
		done:   make(chan struct{}),
	}

	go l.writeLines()
	return l
}

func (l *cmdLogger) appendAndFlush(line []byte) {
	l.buffers = append(l.buffers, line)
	l.flush()
}

func (l *cmdLogger) flush() {
	if len(l.buffers) == 0 {
		return
	}

	var outBuf bytes.Buffer

	l.buffers.WriteTo(&outBuf)
	l.buffers = l.buffers[0:0]

	// TODO: This always adds a newline, which is not always what we want. When
	// we flush partial lines, we don't want to add a newline character. What
	// we need to do: extend the `*output.Output` type to have a
	// `WritefNoNewline` (yes, bad name) method.
	l.out.Writef("%s[%s]%s %s", output.StyleBold, l.name, output.StyleReset, &outBuf)
}

// Write handler of logger.
func (l *cmdLogger) Write(p []byte) (int, error) {
	l.writes <- p
	<-l.done
	return len(p), nil
}

func (l *cmdLogger) writeLines() {
	var tick <-chan time.Time
	for {
		select {
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
						if len(line) != 1 || len(l.buffers) > 0 {
							// We add our own newline.
							l.appendAndFlush(line[0 : len(line)-1])
						}
						tick = nil
					} else {
						l.buffers = append(l.buffers, line)
						tick = time.After(tickDuration)
					}
				}
				if err != nil {
					break
				}
			}
			l.done <- struct{}{}
		case <-tick:
			l.flush()
			tick = nil
		}
	}
}
