package trace

import "sync"

type URLRenderer func(traceID string) string

var (
	urlRenderer   URLRenderer
	urlRendererMu sync.Mutex
)

// RegisterURLRenderer configures the trace package with a URL renderer. The passed
// function will be used to take a traceID and turn it into an environment specific URL
// that can be used to make debugging a little easier.
// Passing `nil` will disable the URL renderer.
func RegisterURLRenderer(ur URLRenderer) {
	urlRendererMu.Lock()
	urlRenderer = ur
	urlRendererMu.Unlock()
}
