package vfsutil

import "sync"

// If we're saving to the local FS, we need to globally synchronize
// writes so we don't corrupt the .zip files with concurrent
// writes. We also needn't bother fetching the same file concurrently,
// since we'll be able to reuse it in the second caller.
//
// This URL mutex is shared among multiple VFS implementations in this
// package.

var (
	urlMusMu sync.Mutex
	urlMus   = map[string]*sync.Mutex{}
)

func urlMu(path string) *sync.Mutex {
	urlMusMu.Lock()
	mu, ok := urlMus[path]
	if !ok {
		mu = new(sync.Mutex)
		urlMus[path] = mu
	}
	urlMusMu.Unlock()
	return mu
}
