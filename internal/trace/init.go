package trace

type URLRenderer func(traceID string) string

var urlRenderer URLRenderer

// Init initializes the trace package with a URL renderer. The passed function
// will be used to take a traceID and turn it into an environment specific URL
// that can be used to make debugging a little easier.
func Init(ur URLRenderer) {
	urlRenderer = ur
}
