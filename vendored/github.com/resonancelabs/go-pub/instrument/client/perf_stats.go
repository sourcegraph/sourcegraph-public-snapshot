package client

import (
	"runtime"
	"sync"

	"src.sourcegraph.com/sourcegraph/vendored/github.com/resonancelabs/go-pub/base"
)

var (
	// protects gPerfStats
	gPerfLock sync.RWMutex

	// gPerfStats isn't part of the Runtime struct because it is, by
	// definition, process-global.
	//
	// TODO: If/when we do proper per-runtime timeseries monitoring, it would
	// be reasonable to move this to the Runtime, though the proxy RPCs would
	// need to include that info as well.
	gPerfStats perfStats
)

const kPerfMaxStalenessMicros = 1 * base.MICROS_PER_SECOND

type perfStats struct {
	// The timestamp that applies to the rest of this struct.
	PerfSampleMicros base.Micros

	// General memory stuff.
	BytesInUse      uint64 // MemStats.Alloc
	BytesFromSystem uint64 // MemStats.Sys

	// Heap stuff.
	HeapInUse       uint64      // MemStats.HeapAlloc
	LastGCMicros    base.Micros // MemStats.LastGC / 1000
	NextGCHeapInUse uint64      // MemStats.NextGC

	// CPU stuff: golang does not make this portable && cheap. TODO?

	// IO stuff: golang does not make this easy, either... also TODO.
}

// Caller must not hold r.lock!
func maybeRefreshPerfStats() {
	now := base.NowMicros()

	// Fast path.
	gPerfLock.RLock()
	if now-gPerfStats.PerfSampleMicros < kPerfMaxStalenessMicros {
		gPerfLock.RUnlock()
		return
	}
	gPerfLock.RUnlock()

	// Slow path.
	gPerfLock.Lock()
	// (We have to check again per the usual fast/slow pattern)
	if now-gPerfStats.PerfSampleMicros < kPerfMaxStalenessMicros {
		gPerfLock.Unlock()
		return
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	gPerfStats.PerfSampleMicros = now
	gPerfStats.BytesInUse = m.Alloc
	gPerfStats.BytesFromSystem = m.Sys
	gPerfStats.HeapInUse = m.HeapAlloc
	gPerfStats.LastGCMicros = base.Micros(m.LastGC / base.NS_PER_MICRO)
	gPerfStats.NextGCHeapInUse = m.NextGC

	gPerfLock.Unlock()
}
