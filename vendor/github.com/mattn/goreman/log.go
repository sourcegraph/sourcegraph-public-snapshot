package main

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/daviddengcn/go-colortext"
)

type clogger struct {
	idx  int
	proc string
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

// write handler of logger.
func (l *clogger) Write(p []byte) (int, error) {
	buf := bytes.NewBuffer(p)
	wrote := 0
	for {
		line, err := buf.ReadBytes('\n')
		if len(line) > 1 {
			now := time.Now().Format("15:04:05")
			format := fmt.Sprintf("%%s %%%ds | ", maxProcNameLength)
			s := string(line)

			mutex.Lock()
			ct.ChangeColor(colors[l.idx], false, ct.None, false)
			fmt.Printf(format, now, l.proc)
			ct.ResetColor()
			fmt.Print(s)
			mutex.Unlock()

			wrote += len(line)
		}
		if err != nil {
			break
		}
	}
	if len(p) > 0 && p[len(p)-1] != '\n' {
		fmt.Println()
	}
	return len(p), nil
}

// create logger instance.
func createLogger(proc string) *clogger {
	mutex.Lock()
	defer mutex.Unlock()
	l := &clogger{ci, proc}
	ci++
	if ci >= len(colors) {
		ci = 0
	}
	return l
}
