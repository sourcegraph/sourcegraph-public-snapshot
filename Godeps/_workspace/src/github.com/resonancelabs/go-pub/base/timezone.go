package base

import (
	"sync"
	"time"
)

var (
	tzCacheLock = new(sync.RWMutex)
	tzCache     = make(map[string]*time.Location)
)

// Provides a cache around time.LoadLocation.  Will return nil on any error.
// time.LoadLocation by default does not cache Location objects, and searches around on disk/reads contents
// from files every time they are asked for.
func LoadLocation(name string) *time.Location {
	// Don't cache the local tz in case it changes
	if name == "Local" {
		return time.Local
	}

	tzCacheLock.RLock()
	location, ok := tzCache[name]
	tzCacheLock.RUnlock()
	if !ok {
		location, _ = time.LoadLocation(name)
		tzCacheLock.Lock()
		tzCache[name] = location
		tzCacheLock.Unlock()
	}

	return location
}
