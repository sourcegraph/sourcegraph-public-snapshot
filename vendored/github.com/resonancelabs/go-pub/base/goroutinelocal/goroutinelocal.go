// Provide a sane-looking (emphasis on "looking") goroutine-local storage API.
// See ./README.md for more on this. This implementation is alarmingly fragile
// but the golang authors give us no non-fragile alternatives.
package goroutinelocal

import (
	"bytes"
	"math"
	"runtime"
	"strconv"
	"sync"
)

type GoroutineId uint64

type genericMap map[string]interface{}

var globalLock sync.Mutex
var globalMap map[GoroutineId]genericMap

func init() {
	globalMap = make(map[GoroutineId]genericMap)
}

var smallBufPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 64)
		return &buf
	},
}

var kGoroutinePrefix = []byte("goroutine ")

// Return the currently active GoroutineId.
func CurrGoroutineId() GoroutineId {
	bufferPtr := smallBufPool.Get().(*[]byte)
	defer smallBufPool.Put(bufferPtr)
	buffer := *bufferPtr
	buffer = buffer[:runtime.Stack(buffer, false)]
	// Parse the 4707 out of "goroutine 4707 ["
	buffer = bytes.TrimPrefix(buffer, kGoroutinePrefix)
	i := bytes.IndexByte(buffer, ' ')
	if i < 0 {
		// XXX: panic'ing is no-kay, but we'd love to log this somewhere?
		// panic(fmt.Sprintf("No space found in %q", buffer))
		return math.MaxUint64
	}
	s := string(buffer[:i]) // Convert to string for ParseUint.
	n, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		// XXX: panic'ing is no-kay, but we'd love to log this somewhere?
		// panic(fmt.Sprintf("Failed to parse goroutine ID out of %q: %v", s, err))
		return math.MaxUint64
	}
	return GoroutineId(n)
}

// Set key:val in goroutine-local storage. Also see GetWithDefault.
func Set(key string, val interface{}) {
	grid := CurrGoroutineId()

	globalLock.Lock()
	defer globalLock.Unlock()

	gmap, found := globalMap[grid]
	if !found {
		gmap = make(genericMap)
		globalMap[grid] = gmap
	}
	gmap[key] = val
}

// Get the value associated (via Set) with `key`, or nil if no such key is
// found in goroutine-local storage.
func Get(key string) interface{} {
	grid := CurrGoroutineId()

	globalLock.Lock()
	defer globalLock.Unlock()

	if gmap, found := globalMap[grid]; found {
		return gmap[key]
	} else {
		return nil
	}
}

// Get the value associated (via Set) with `key`, or insert and return `def` if
// no such key is found in goroutine-local storage.
func GetWithDefault(key string, def interface{}) interface{} {
	grid := CurrGoroutineId()

	globalLock.Lock()
	defer globalLock.Unlock()

	gmap, found := globalMap[grid]
	if !found {
		gmap = make(genericMap)
		globalMap[grid] = gmap
	}
	val, found := gmap[key]
	if !found {
		val = def
		gmap[key] = val
	}
	return val
}

// Clear any value associated with `key` in goroutine-local storage.
func Clear(key string) {
	grid := CurrGoroutineId()

	globalLock.Lock()
	defer globalLock.Unlock()

	if gmap, found := globalMap[grid]; found {
		delete(gmap, key)
		if len(gmap) == 0 {
			delete(globalMap, grid)
		}
	}
}
