package ui

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
	"time"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/golang/groupcache/lru"
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/appdash"
	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/ui/reactbridge"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/synclru"
	"sourcegraph.com/sourcegraph/sourcegraph/util/traceutil"
)

const maxEntries = 10000

var (
	rendererMu       sync.Mutex       // guards renderer
	rendererCacheKey string           // used to evict singleton renderer when bundle.js changes
	rendererOnce     *sync.Once       // ensures max 1 in-flight call to newCachingRenderer
	rendererInFlight string           // cache key of in-flight call to newCachingRenderer
	rendererInit     time.Time        // when the renderer creation began
	renderer         *cachingRenderer // singleton renderer (pooled)
	rendererErr      error            // error from last attempt to create renderer

	renderPoolSize = runtime.GOMAXPROCS(0)
)

type createRendererEvent struct {
	S, E time.Time
}

func init() {
	appdash.RegisterEvent(createRendererEvent{})
}

func (createRendererEvent) Schema() string     { return "createRenderer" }
func (e createRendererEvent) Start() time.Time { return e.S }
func (e createRendererEvent) End() time.Time   { return e.E }

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

// getRenderer gets (creating if needed) a JS renderer. It returns the
// error errRendererCreationTimedOut if the renderer could not be
// created within the ctx' deadline.
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
	if cacheKey == rendererCacheKey {
		rendererMu.Unlock()
		return renderer, rendererErr
	}

	start := time.Now()
	ctx = traceutil.NewContext(ctx, appdash.NewSpanID(traceutil.SpanIDFromContext(ctx)))
	defer func() {
		rec := traceutil.Recorder(ctx)
		rec.Name("ui.getRenderer")
		rec.Event(&createRendererEvent{S: start, E: time.Now()})
	}()

	// We need to create a new renderer. Only allow 1 at a time because
	// this is an expensive operation and all operations will return the
	// same results (and overwrite each other anyway).
	if rendererInFlight != cacheKey {
		rendererInFlight = cacheKey
		rendererOnce = new(sync.Once)
		rendererInit = time.Now()
	}
	rendererMu.Unlock()

	done := make(chan struct{})
	go func() {
		// Obtain the ptr with the lock to avoid a race.
		rendererMu.Lock()
		tmp := rendererOnce
		rendererMu.Unlock()

		tmp.Do(func() {
			r, err := newCachingRenderer(js)

			rendererMu.Lock()
			rendererCacheKey = cacheKey
			renderer = r
			rendererErr = err
			rendererMu.Unlock()
		})
		close(done)
	}()
	select {
	case <-done:
		rendererMu.Lock()
		defer rendererMu.Unlock()
		return renderer, rendererErr

	case <-ctx.Done():
		if ctx.Err() == context.DeadlineExceeded {
			// Warn if taking longer than it should.
			rendererMu.Lock()
			elapsed := time.Since(rendererInit)
			rendererMu.Unlock()
			if threshold := 15 * time.Second; elapsed > threshold {
				log15.Warn("JS renderer creation is taking longer than expected", "elapsed", elapsed, "threshold", threshold)
			}

			return nil, errRendererCreationTimedOut
		}
		return nil, err
	}
}

var errRendererCreationTimedOut = errors.New("JS renderer creation timed out")

func newCachingRenderer(js string) (*cachingRenderer, error) {
	errCh := make(chan error)
	r := &cachingRenderer{
		bridge: reactbridge.New(js, renderPoolSize, errCh),
		cache:  synclru.New(lru.New(maxEntries)),
	}

	// See if there's a SyntaxError. If there are none, this channel
	// will be closed, and err will be nil.
	//
	// This only waits for 1 VM to load, not the whole pool, so we
	// get results more quickly.
	if err := <-errCh; err != nil {
		return nil, err
	}

	return r, nil
}

type cachingRenderer struct {
	bridge *reactbridge.Bridge
	cache  *synclru.Cache
}

type prerenderEvent struct {
	Props    string
	CacheHit bool
	S, E     time.Time
}

func init() {
	appdash.RegisterEvent(prerenderEvent{})
}

func (prerenderEvent) Schema() string     { return "prerenderReactComponent" }
func (e prerenderEvent) Start() time.Time { return e.S }
func (e prerenderEvent) End() time.Time   { return e.E }

// callMain calls r.bridge.CallMain with caching.
func (r *cachingRenderer) callMain(ctx context.Context, arg interface{}) (string, error) {
	argJSON, err := json.Marshal(arg)
	if err != nil {
		return "", err
	}

	// Construct cache key.
	keyArray := sha256.Sum256(argJSON)
	key := string(keyArray[:])

	// Get from cache.
	cachedRes, cacheHit := r.cache.Get(key)

	// Log in Appdash.
	start := time.Now()
	ctx = traceutil.NewContext(ctx, appdash.NewSpanID(traceutil.SpanIDFromContext(ctx)))
	defer func() {
		truncatedProps := argJSON
		if max := 300; len(truncatedProps) > max {
			truncatedProps = truncatedProps[:max]
		}

		rec := traceutil.Recorder(ctx)
		rec.Name("prerender React component")
		rec.Event(&prerenderEvent{
			Props:    string(truncatedProps),
			CacheHit: cacheHit,
			S:        start,
			E:        time.Now(),
		})
	}()

	// Return cache hit if present.
	if cacheHit {
		return cachedRes.(string), nil
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
	r, err := getRenderer(ctx)
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

// reactCompatibleHTMLEscape is like template.HTMLEscape, but it uses
// the same HTML entities that React does (in
// escapeTextContentForBrowser). This ensures that the HTML rendered
// by Go (e.g., in blob.render) is identical to that rendered by React
// (which is necessary for server-side rendering).
func reactCompatibleHTMLEscape(w io.Writer, b []byte) {
	last := 0
	for i, c := range b {
		var html []byte
		switch c {
		case '"':
			html = htmlQuot
		case '\'':
			html = htmlApos
		case '&':
			html = htmlAmp
		case '<':
			html = htmlLt
		case '>':
			html = htmlGt
		default:
			continue
		}
		w.Write(b[last:i])
		w.Write(html)
		last = i + 1
	}
	w.Write(b[last:])
}

var (
	// Used by reactCompatibleHTMLEscape.
	htmlQuot = []byte("&quot;")
	htmlApos = []byte("&#x27;")
	htmlAmp  = []byte("&amp;")
	htmlLt   = []byte("&lt;")
	htmlGt   = []byte("&gt;")

	// Go-style, not React-style, HTML escapes.
	htmlApos2 = []byte("&#39;")
	htmlQuot2 = []byte("&#34;")
)

// convertToReactHTMLEscapeStyle uses React-style
// (escapeTextContentForBrowser) escapes instead of
// golang.org/x/net/html-style escapes.
func convertToReactHTMLEscapeStyle(escapedHTML []byte) []byte {
	escapedHTML = bytes.Replace(escapedHTML, htmlApos2, htmlApos, -1)
	escapedHTML = bytes.Replace(escapedHTML, htmlQuot2, htmlQuot, -1)
	return escapedHTML
}
