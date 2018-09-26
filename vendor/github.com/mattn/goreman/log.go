package main

import (
	"bytes"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/mattn/go-colorable"
)

type clogger struct {
	idx     int
	proc    string
	writes  chan []byte
	done    chan struct{}
	timeout time.Duration // how long to wait before printing partial lines
	buffers net.Buffers   // partial lines awaiting printing
}

var colors = []int{
	32, // green
	36, // cyan
	35, // magenta
	33, // yellow
	34, // blue
	31, // red
}
var mutex = new(sync.Mutex)

var out = colorable.NewColorableStdout()

// write any stored buffers, plus the given line, then empty out
// the buffers.
func (l *clogger) writeBuffers(line []byte) {
	now := time.Now().Format("15:04:05")
	mutex.Lock()
	fmt.Fprintf(out, "\x1b[%dm", colors[l.idx])
	fmt.Fprintf(out, "%s %*s | ", now, maxProcNameLength, l.proc)
	fmt.Fprintf(out, "\x1b[m")
	l.buffers = append(l.buffers, line)
	l.buffers.WriteTo(out)
	l.buffers = l.buffers[0:0]
	mutex.Unlock()
}

// bundle writes into lines, waiting briefly for completion of lines
func (l *clogger) writeLines() {
	var tick <-chan time.Time
	for {
		select {
		case w, ok := <-l.writes:
			if !ok {
				if len(l.buffers) > 0 {
					l.writeBuffers([]byte("\n"))
				}
				return
			}
			buf := bytes.NewBuffer(w)
			for {
				line, err := buf.ReadBytes('\n')
				if len(line) > 0 {
					if line[len(line)-1] == '\n' {
						// any text followed by a newline should flush
						// existing buffers. a bare newline should flush
						// existing buffers, but only if there are any.
						if len(line) != 1 || len(l.buffers) > 0 {
							l.writeBuffers(line)
						}
						tick = nil
					} else {
						l.buffers = append(l.buffers, line)
						tick = time.After(l.timeout)
					}
				}
				if err != nil {
					break
				}
			}
			l.done <- struct{}{}
		case <-tick:
			if len(l.buffers) > 0 {
				l.writeBuffers([]byte("\n"))
			}
			tick = nil
		}
	}

}

// write handler of logger.
func (l *clogger) Write(p []byte) (int, error) {
	l.writes <- p
	<-l.done
	return len(p), nil
}

// create logger instance.
func createLogger(proc string, colorIndex int) *clogger {
	mutex.Lock()
	defer mutex.Unlock()
	l := &clogger{idx: colorIndex, proc: proc, writes: make(chan []byte), done: make(chan struct{}), timeout: 2 * time.Millisecond}
	go l.writeLines()
	return l
}
