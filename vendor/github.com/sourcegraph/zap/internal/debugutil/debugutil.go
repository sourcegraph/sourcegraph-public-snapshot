// Package debugutil contains utilities for debugging Zap.
package debugutil

import (
	"math"
	"math/rand"
	"os"
	"strconv"
	"sync"
	"time"
)

var (
	Mu sync.Mutex // guards all vars in this block

	SimulatedLatency, _          = time.ParseDuration(os.Getenv("SIMULATED_LATENCY"))
	RandomizeSimulatedLatency, _ = strconv.ParseBool(os.Getenv("RANDOMIZE_SIMULATED_LATENCY"))
)

func SimulateLatency() {
	Mu.Lock()
	if SimulatedLatency == 0 && !RandomizeSimulatedLatency {
		Mu.Unlock()
		return
	}
	d := SimulatedLatency
	if RandomizeSimulatedLatency {
		if d == 0 {
			d = 10 * time.Millisecond // default
		}
		x := math.Abs(rand.NormFloat64()*0.6 + 1)
		if x < 0.1 || x > 3 {
			x = 1
		}
		d = time.Duration(float64(d) * x)
	}
	Mu.Unlock()
	time.Sleep(d)
}
