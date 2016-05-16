package experiment_test

import (
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/experiment"
)

func TestPerf_AFaster(t *testing.T) {
	d := testPerf(time.Millisecond, 5*time.Millisecond)
	if d.bDur < d.aDur {
		t.Errorf("Expected A to be faster. a: %v b: %v", d.aDur, d.bDur)
	}
}

func TestPerf_BFaster(t *testing.T) {
	d := testPerf(5*time.Millisecond, time.Millisecond)
	if d.aDur < d.bDur {
		t.Errorf("Expected B to be faster. a: %v b: %v", d.aDur, d.bDur)
	}
}

func TestPerf_DefaultReporter(t *testing.T) {
	// This just ensures we don't panic if we don't set a reporter
	e := experiment.Perf{
		Name: "test",
		B:    func() {},
	}
	done := e.StartA()
	done()
	// reporting is done in another goroutine, just give it time to panic
	// if it does
	time.Sleep(5 * time.Millisecond)
}

type perfDurations struct{ aDur, bDur time.Duration }

func testPerf(a, b time.Duration) perfDurations {
	reported := make(chan perfDurations)
	e := experiment.Perf{
		Name: "test",
		B:    func() { time.Sleep(b) },
		Report: func(_ string, aDur, bDur time.Duration) {
			reported <- perfDurations{aDur, bDur}
		},
	}
	done := e.StartA()
	time.Sleep(a)
	done()
	return <-reported
}
