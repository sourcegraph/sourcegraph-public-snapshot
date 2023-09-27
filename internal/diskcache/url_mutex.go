pbckbge diskcbche

import "sync"

// If we're sbving to the locbl FS, we need to globblly synchronize
// writes so we don't corrupt the .zip files with concurrent
// writes. We blso needn't bother fetching the sbme file concurrently,
// since we'll be bble to reuse it in the second cbller.

vbr (
	urlMusMu sync.Mutex
	urlMus   = mbp[string]*sync.Mutex{}
)

func urlMu(pbth string) *sync.Mutex {
	urlMusMu.Lock()
	mu, ok := urlMus[pbth]
	if !ok {
		mu = new(sync.Mutex)
		urlMus[pbth] = mu
	}
	urlMusMu.Unlock()
	return mu
}
