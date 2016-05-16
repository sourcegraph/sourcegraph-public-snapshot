// Package experiment provides a simple abstractions for running A/B
// tests. The APIs are targetted at infrastructure needs, rather than for
// tests on users.
package experiment

import (
	"time"

	"gopkg.in/inconshreveable/log15.v2"
)

// Perf is an A/B test where we want to decide on which version is faster. A
// is the current version and B is the new version
type Perf struct {
	// Name is passed to Report. Used for identifying an experiment
	Name string

	// B is the new version to test against. Its return values are ignored
	// and it is run purely to record its execution time.
	B func()

	// Report is an optional function to set which can override the
	// behaviour of reporting the durations. By default we use log15.Debug
	Report func(name string, aDur, bDur time.Duration)
}

// StartA indicates that you are starting the current version A. It returns a
// function done which should be called when A is finished. StartA will also
// start an execution of B in another goroutine and record its run time.
//
// Note: Calling done() will not block on B
func (p *Perf) StartA() (done func()) {
	var aDur, bDur time.Duration
	start := time.Now()
	aDone := make(chan struct{}, 1)
	go func() {
		p.B()
		bDur = time.Since(start)

		<-aDone
		r := p.Report
		if r == nil {
			r = defaultPerfReport
		}
		r(p.Name, aDur, bDur)
	}()
	return func() {
		aDur = time.Since(start)
		aDone <- struct{}{}
	}
}

func defaultPerfReport(name string, aDur, bDur time.Duration) {
	faster := "a"
	if bDur < aDur {
		faster = "b"
	}
	log15.Debug("experiment", "name", name, "faster", faster, "aDur", aDur, "bDur", bDur)
}
