package ui

import (
	"crypto/sha256"
	"encoding/json"
	"io/ioutil"
	"runtime"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/golang/groupcache/lru"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/jsserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/synclru"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

const maxEntries = 10000

var (
	rendererMu       sync.Mutex       // guards renderer
	rendererCacheKey string           // used to evict singleton renderer when bundle JS changes
	renderer         *cachingRenderer // singleton renderer for this cache key
	rendererErr      error            // error from generating renderer for this cache key

	renderPoolSize = runtime.GOMAXPROCS(0)
)

func readBundleJS() ([]byte, error) {
	f, err := assets.Assets.Open("/bundle.js")
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// getRenderer gets (creating if needed) a JS renderer.
func getRenderer(ctx context.Context) (*cachingRenderer, error) {
	// Don't respect the ctx timeout for getting the bundle JS, since
	// that is only an async operation in dev. In production, it is
	// read from the in-memory bundled asset data structure, so it
	// will never block (in practice; obviously memory access takes
	// some nanoseconds).
	js, cacheKey, err := getBundleJS()
	if err != nil {
		return nil, err
	}

	// Fastpath for when the renderer has already been created and the
	// bundle JS hasn't changed.
	rendererMu.Lock()
	defer rendererMu.Unlock()
	if cacheKey == rendererCacheKey {
		return renderer, rendererErr
	}

	// Need to create a new renderer.
	if renderer != nil {
		if err := renderer.Close(); err != nil {
			log15.Warn("Error closing existing JS renderer.", "err", err)
		}
	}

	renderer = newCachingRenderer(js)
	rendererCacheKey = cacheKey
	return renderer, rendererErr
}

func newCachingRenderer(js []byte) *cachingRenderer {
	return &cachingRenderer{
		s:     jsserver.NewPool(js, renderPoolSize),
		cache: synclru.New(lru.New(maxEntries)),
	}
}

type cachingRenderer struct {
	s     jsserver.Server
	cache *synclru.Cache
}

type prerenderEvent struct {
	Arg      string
	CacheHit bool
	S, E     time.Time
}

func init() {
	appdash.RegisterEvent(prerenderEvent{})
}

func (prerenderEvent) Schema() string     { return "prerenderReactComponent" }
func (e prerenderEvent) Start() time.Time { return e.S }
func (e prerenderEvent) End() time.Time   { return e.E }

// call calls r.s.Call with caching.
func (r *cachingRenderer) Call(ctx context.Context, arg json.RawMessage) ([]byte, error) {
	// Construct cache key.
	keyArray := sha256.Sum256(arg)
	key := string(keyArray[:])

	// Get from cache.
	cachedRes, cacheHit := r.cache.Get(key)

	// Log in Appdash.
	start := time.Now()
	rec := traceutil.Recorder(ctx)
	rec.Name("prerender React component")
	defer func() {
		truncatedArg := arg
		if max := 300; len(truncatedArg) > max {
			truncatedArg = truncatedArg[:max]
		}
		rec.Event(&prerenderEvent{
			Arg:      string(truncatedArg),
			CacheHit: cacheHit,
			S:        start,
			E:        time.Now(),
		})
	}()

	// Return cache hit if present.
	if cacheHit {
		return cachedRes.([]byte), nil
	}

	res, err := r.s.Call(ctx, arg)
	if err == nil {
		r.cache.Add(key, res)
	}
	return res, err

}

func (r *cachingRenderer) Close() error {
	return r.s.Close()
}

func renderReactComponent(ctx context.Context, componentModule string, props interface{}, stores *StoreData) ([]byte, error) {
	r, err := getRenderer(ctx)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(struct {
		ComponentModule string
		Props           interface{}
		Stores          *StoreData
	}{
		ComponentModule: componentModule,
		Props:           props,
		Stores:          stores,
	})
	if err != nil {
		return nil, err
	}

	return r.Call(ctx, data)
}

type contextKey int

const (
	dontPrerenderReactComponents contextKey = iota
)

// DisabledReactPrerendering disables server-side prerendering of React
// components within this context.
func DisabledReactPrerendering(ctx context.Context) context.Context {
	return context.WithValue(ctx, dontPrerenderReactComponents, struct{}{})
}

func shouldPrerenderReact(ctx context.Context) bool {
	return ctx.Value(dontPrerenderReactComponents) == nil
}
