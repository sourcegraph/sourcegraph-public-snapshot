// Package snapshotprof creates snapshots of pprof data
package snapshotprof

import (
	"bytes"
	"encoding/json"
	"io"
	"runtime/pprof"
	"time"
)

// Snapshot stores profiler data at a point in time
type Snapshot struct {
	// Time is the time the Snapshot was created.
	Time time.Time
	// Since is the duration of Time since the program started
	Since time.Duration
	// Duration is how long the Snapshot took to create
	Duration time.Duration
	// Profiles contains the output of pprof.Profile.WriteTo for each
	// registered profiler
	Profiles map[string]string
}

// Capture returns a snapshot of all the profiler data at the current point in
// time.
func Capture() *Snapshot {
	s := &Snapshot{
		Time:     time.Now(),
		Profiles: map[string]string{},
	}
	s.Since = s.Time.Sub(programStart)
	b := &bytes.Buffer{}
	for _, p := range pprof.Profiles() {
		if blacklist[p.Name()] {
			continue
		}
		b.Reset()
		err := p.WriteTo(b, 2)
		if err != nil {
			s.Profiles[p.Name()] = "Error: " + err.Error()
			continue
		}
		s.Profiles[p.Name()] = b.String()
	}
	s.Duration = time.Since(s.Time)
	return s
}

// Run continuously writes to w the current Snapshot as a line of
// JSON. interval is how regularly the write occurs.
func Run(w io.Writer, interval time.Duration) error {
	e := json.NewEncoder(w)
	for {
		time.Sleep(interval)
		err := e.Encode(Capture())
		if err != nil {
			return err
		}
	}
}

// blacklist stores profiles that are too expensive or large to calculate
// regularly
var blacklist = map[string]bool{
	"heap": true,
}

var programStart = time.Now()
