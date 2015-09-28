// Copyright © 2013 Steve Francia <spf@spf13.com>.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Quick and Easy Performance Analyzer
// Useful for comparing A/B against different drafts of functions or different functions
// Loosely inspired by the go benchmark package
//
// Example:
//	import "github.com/spf13/nitro"
//	timer := nitro.Initialize()
//	prepTemplates()
//	timer.Step("initialize & template prep")
//	CreatePages()
//	timer.Step("import pages")
package nitro

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"time"
)

// Used for every benchmark for measuring memory.
var memStats runtime.MemStats

var AnalysisOn = false

type B struct {
	initialTime time.Time // Time entire process started
	start       time.Time // Time step started
	duration    time.Duration
	timerOn     bool
	result      R
	// The initial states of memStats.Mallocs and memStats.TotalAlloc.
	startAllocs uint64
	startBytes  uint64
	// The net total of this test after being run.
	netAllocs uint64
	netBytes  uint64
}

func (b *B) startTimer() {
	if b == nil {
		fmt.Println("ERROR: can't call startTimer on a nil value")
		os.Exit(-1)
	}
	if !b.timerOn {
		runtime.ReadMemStats(&memStats)
		b.startAllocs = memStats.Mallocs
		b.startBytes = memStats.TotalAlloc
		b.start = time.Now()
		b.timerOn = true
	}
}

func (b *B) stopTimer() {
	if b == nil {
		fmt.Println("ERROR: can't call stopTimer on a nil value")
		os.Exit(-1)
	}
	if b.timerOn {
		b.duration += time.Since(b.start)
		runtime.ReadMemStats(&memStats)
		b.netAllocs += memStats.Mallocs - b.startAllocs
		b.netBytes += memStats.TotalAlloc - b.startBytes
		b.timerOn = false
	}
}

// ResetTimer sets the elapsed benchmark time to zero.
// It does not affect whether the timer is running.
func (b *B) resetTimer() {
	if b.timerOn {
		runtime.ReadMemStats(&memStats)
		b.startAllocs = memStats.Mallocs
		b.startBytes = memStats.TotalAlloc
		b.start = time.Now()
	}
	b.duration = 0
	b.netAllocs = 0
	b.netBytes = 0
}

// Call this first to get the performance object
// Should be called at the top of your function.
func Initialize() *B {
	flag.BoolVar(&AnalysisOn, "stepAnalysis", false, "display memory and timing of different steps of the program")

	b := &B{}
	b.initialTime = time.Now()
	runtime.GC()
	b.resetTimer()
	b.startTimer()
	return b
}

// Simple wrapper for Initialize
// Maintain for legacy purposes
func Initalize() *B {
	return Initialize()
}

// Call perf.Step("step name") at each step in your
// application you want to benchmark
// Measures time spent since last Step call.
func (b *B) Step(str string) {
	if !AnalysisOn {
		return
	}

	b.stopTimer()
	fmt.Println(str + ":")
	fmt.Println(b.results().toString())

	b.resetTimer()
	b.startTimer()
}

func (b *B) results() R {
	return R{time.Since(b.initialTime), b.duration, b.netAllocs, b.netBytes}
}

type R struct {
	C         time.Duration // Cumulative time taken
	T         time.Duration // The total time taken.
	MemAllocs uint64        // The total number of memory allocations.
	MemBytes  uint64        // The total number of bytes allocated.
}

func (r R) mbPerSec() float64 {
	if r.MemBytes <= 0 || r.T <= 0 {
		return 0
	}

	return byteToMb(r.MemBytes) / r.T.Seconds()
}

func byteToMb(b uint64) float64 {
	if b <= 0 {
		return 0
	}
	return float64(b) / 1e6
}

func (r R) toString() string {
	time := fmt.Sprintf("%v (%5v)\t", r.T, r.C)
	mem := fmt.Sprintf("%7.2f MB \t%v Allocs", byteToMb(r.MemBytes), r.MemAllocs)
	return fmt.Sprintf("\t%s %s", time, mem)
}
