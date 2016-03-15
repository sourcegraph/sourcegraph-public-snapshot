package ui

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"runtime"
	"sync"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/golang/groupcache/lru"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/reactbridge"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/synclru"
)

const maxEntries = 10000

var (
	rendererMu       sync.Mutex       // guards renderer
	rendererCacheKey string           // used to evict singleton renderer when bundle.js changes
	renderer         *cachingRenderer // singleton renderer (pooled)

	renderPoolSize = runtime.GOMAXPROCS(0)
)

func readBundleJS() (string, error) {
	f, err := assets.Assets.Open("/bundle.js")
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func getRenderer() (*cachingRenderer, error) {
	rendererMu.Lock()
	defer rendererMu.Unlock()

	js, cacheKey, err := getBundleJS()
	if err != nil {
		return nil, err
	}

	if cacheKey == rendererCacheKey {
		log15.Info("Renderer cache HIT")
		return renderer, nil
	}

	log15.Info("Renderer cache MISS")
	rendererCacheKey = cacheKey
	renderer = newCachingRenderer(js)
	return renderer, nil
}

func newCachingRenderer(js string) *cachingRenderer {
	errCh := make(chan error)
	loadedOneVM := make(chan struct{})
	r := &cachingRenderer{
		bridge:      reactbridge.New(js, renderPoolSize, errCh),
		cache:       synclru.New(lru.New(maxEntries)),
		loadedOneVM: loadedOneVM,
	}

	go func() {
		// See if there's a SyntaxError. If there are none, this channel
		// will be closed, and r.err will be nil.
		r.err = <-errCh
		close(loadedOneVM)
	}()

	return r
}

type cachingRenderer struct {
	bridge *reactbridge.Bridge
	cache  *synclru.Cache

	loadedOneVM <-chan struct{} // closed after at least 1 JS VM loads (and been checked for SyntaxError)
	err         error           // init error (likely a SyntaxError)
}

// callMain calls r.bridge.CallMain with caching.
func (r *cachingRenderer) callMain(ctx context.Context, arg interface{}) (string, error) {
	// Don't proceed if init failed (likely a SyntaxError).
	<-r.loadedOneVM
	if r.err != nil {
		return "", r.err
	}

	argJSON, err := json.Marshal(arg)
	if err != nil {
		return "", err
	}

	// Construct cache key.
	keyArray := sha256.Sum256(argJSON)
	key := string(keyArray[:])

	hit, ok := r.cache.Get(key)
	if ok {
		return hit.(string), nil
	}

	// Optimization: pass the already-JSON-marshaled argJSON instead of
	// having CallMain re-marshal arg.
	arg = (*json.RawMessage)(&argJSON)

	res, err := r.bridge.CallMain(ctx, arg)
	if err == nil {
		r.cache.Add(key, res)
	}
	return res, err

}

// renderReactComponent renders the React component exported (as
// default) by the named JavaScript module. It passes the given props
// as the component's props. If there is a prop named "component",
// then its value resolved to the default-exported component at the
// given path. The provided store data is preloaded into the stores
// prior to rendering the components.
//
// NOTE: React 15 contains a syntactically incorrect RegExp that is
// accepted by V8 but not Duktape. (It's the "\uB7" at
// https://github.com/facebook/react/blob/f8046f2dc22e669e300d2d9a967e5c5bfa1b105b/src/renderers/dom/shared/DOMProperty.js#L169.)
// Be sure that the app/node_modules/react/lib/DOMProperty.js file has
// a manual edit to make this "\u00B7".
func renderReactComponent(ctx context.Context, componentModule string, props interface{}, stores *StoreData) (string, error) {
	r, err := getRenderer()
	if err != nil {
		return "", err
	}

	data := struct {
		ComponentModule string
		Props           interface{}
		Stores          *StoreData
	}{
		ComponentModule: componentModule,
		Props:           props,
		Stores:          stores,
	}
	return r.callMain(ctx, data)
}
