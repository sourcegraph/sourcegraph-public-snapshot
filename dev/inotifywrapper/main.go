package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type lineWatcher struct {
	timeout time.Duration
	buffers [][]byte
	events  chan string
}

// do the actual notification dance.
func (l *lineWatcher) notify(events map[string]struct{}) {
	args := make([]string, 0, len(events))
	for k := range events {
		for _, r := range matchRegexes {
			if r.re.MatchString(k) {
				args = append(args, k)
				break
			}
		}
		delete(events, k)
	}
	if len(args) > 0 {
		cmd := exec.Command(watchCmd, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			log.Printf("failure running '%s': %s", watchCmd, err)
		}
	}
}

// when an event comes in, start a timer. collect events until the timer
// runs out or the event source closes, then notify them.
func (l *lineWatcher) combineEvents() {
	recent := make(map[string]struct{})
	var tick <-chan time.Time
	for {
		select {
		case f, ok := <-l.events:
			if !ok {
				l.notify(recent)
				return
			}
			recent[f] = struct{}{}
			if tick == nil {
				tick = time.After(l.timeout)
			}
		case <-tick:
			l.notify(recent)
			tick = nil
		}
	}
}

func (l *lineWatcher) handleLine() {
	if len(l.buffers) < 1 {
		return
	}
	line := bytes.Join(l.buffers, []byte(""))
	l.events <- string(line)
	l.buffers = l.buffers[0:0]
}

// Write() separates out lines from input and forwards them to the
// events channel, saving partial lines for later combining.
//
// Note: If a trailing line of input has no newline, and no following
// writes happen, that data gets lost. This is okay, since (1) it can't
// happen, (2) this is best-effort.
func (l *lineWatcher) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer(p)
	for {
		line, err := buf.ReadBytes('\n')
		if len(line) > 0 {
			if line[len(line)-1] == '\n' {
				l.buffers = append(l.buffers, line[:len(line)-1])
				l.handleLine()
			} else {
				l.buffers = append(l.buffers, line)
			}
		}
		if err != nil {
			break
		}
	}
	return len(p), nil
}

func newLineWatcher() *lineWatcher {
	l := &lineWatcher{events: make(chan string), timeout: 1 * time.Second}
	go l.combineEvents()
	return l
}

type pathRegex struct {
	raw string
	re  *regexp.Regexp
}

type pathRegexes []pathRegex

var matchRegexes pathRegexes

func (p *pathRegexes) String() string {
	if p == nil {
		return "nil pathRegexes"
	}
	raws := make([]string, 0, len(*p))
	for _, e := range *p {
		raws = append(raws, e.raw)
	}
	return strings.Join(raws, ",")
}

func (p *pathRegexes) Set(s string) error {
	re, err := regexp.Compile(s)
	if err != nil {
		return err
	}
	pre := pathRegex{raw: s, re: re}
	*p = append(*p, pre)
	return nil
}

type pathList []string

var watchPaths pathList

func (p *pathList) String() string {
	if p == nil {
		return "nil pathList"
	}
	return strings.Join(*p, ",")
}

func (p *pathList) Set(s string) error {
	*p = append(*p, s)
	return nil
}

var watchCmd string

func main() {
	flag.Var(&matchRegexes, "match", "path regexps to match")
	flag.Var(&watchPaths, "path", "paths to watch")
	flag.StringVar(&watchCmd, "cmd", "./dev/handle-change.sh", "command to run with matched paths")
	flag.Parse()
	if len(watchPaths) < 0 {
		log.Fatal("must specify at least one path to watch [-path foo]")
	}
	inotifyArgs := []string{
		"inotifywait", "-mqr", "--format", "%w%f", "-e", "modify", "-e", "create", "-e", "close_write", "-e", "delete", "-e", "move",
	}
	inotifyArgs = append(inotifyArgs, watchPaths...)
	cmd := exec.Command(inotifyArgs[0], inotifyArgs[1:]...)
	lines := newLineWatcher()
	cmd.Stdout = lines
	cmd.Stderr = lines
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}
