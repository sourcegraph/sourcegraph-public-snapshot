// Package multicache provides a "fallback" cache implementation that
// short-circuits gets and writes/deletes to all underlying caches.
package multicache // import "sourcegraph.com/sourcegraph/multicache"

// Fallback is a cache that wraps a list of caches. Gets read from the caches in
// sequence until a cache entry is found. Sets write to all caches, returning
// after the first WaitNSets Set operations complete. Deletes delete from all
// caches, returning after the first WaitNDeletes Delete operations complete.
type Fallback struct {
	caches       []Underlying
	WaitNSets    int
	WaitNDeletes int
}

func (f *Fallback) Get(key string) (resp []byte, ok bool) {
	for _, c := range f.caches {
		resp, ok = c.Get(key)
		if ok {
			return
		}
	}
	return
}

func (f *Fallback) Set(key string, resp []byte) {
	for i, c := range f.caches {
		if i < f.WaitNSets {
			c.Set(key, resp)
		} else {
			go c.Set(key, resp)
		}
	}
}

func (f *Fallback) Delete(key string) {
	for i, c := range f.caches {
		if i < f.WaitNDeletes {
			c.Delete(key)
		} else {
			go c.Delete(key)
		}
	}
}

// NewFallback returns a new Fallback cache with WaitNSets == WaitNDeletes ==
// len(caches).
func NewFallback(caches ...Underlying) *Fallback {
	return &Fallback{caches: caches, WaitNSets: len(caches), WaitNDeletes: len(caches)}
}
