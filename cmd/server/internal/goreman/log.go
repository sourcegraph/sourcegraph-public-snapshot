package goreman

import (
	"bytes"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	ct "github.com/daviddengcn/go-colortext"
)

type clogger struct {
	idx     int
	proc    string
	writes  chan []byte
	done    chan struct{}
	timeout time.Duration // how long to wait before printing partial lines
	buffers net.Buffers   // partial lines awaiting printing
}

var colors = []ct.Color{
	ct.Green,
	ct.Cyan,
	ct.Magenta,
	ct.Yellow,
	ct.Blue,
	ct.Red,
}
var ci int

var mutex = new(sync.Mutex)

// write any stored buffers, plus the given line, then empty out
// the buffers.
func (l *clogger) writeBuffers(line []byte) {
	now := time.Now().Format("15:04:05")
	mutex.Lock()
	ct.ChangeColor(colors[l.idx], false, ct.None, false)
	fmt.Printf("%s %*s | ", now, maxProcNameLength, l.proc)
	ct.ResetColor()
	l.buffers = append(l.buffers, line)
	_, _ = l.buffers.WriteTo(os.Stdout)
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
func createLogger(proc string) *clogger {
	mutex.Lock()
	defer mutex.Unlock()
	l := &clogger{idx: ci, proc: proc, writes: make(chan []byte), done: make(chan struct{}), timeout: 2 * time.Millisecond}
	go l.writeLines()
	ci++
	if ci >= len(colors) {
		ci = 0
	}
	return l
}
