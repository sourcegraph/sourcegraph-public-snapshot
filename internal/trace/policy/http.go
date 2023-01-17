package policy

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
)

const (
	traceHeader = "X-Sourcegraph-Should-Trace"
	traceQuery  = "trace"
)

// Transport wraps an underlying HTTP RoundTripper, injecting the X-Sourcegraph-Should-Trace header
// into outgoing requests whenever the shouldTraceKey context value is true.
type Transport struct {
	http.RoundTripper
}

func (r *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set(traceHeader, strconv.FormatBool(ShouldTrace(req.Context())))
	t := nethttp.Transport{RoundTripper: r.RoundTripper}
	return t.RoundTrip(req)
}

// requestWantsTrace returns true if a request is opting into tracing either
// via our HTTP Header or our URL Query.
func RequestWantsTracing(r *http.Request) bool {
	// Prefer header over query param.
	if v := r.Header.Get(traceHeader); v != "" {
		b, _ := strconv.ParseBool(v)
		return b
	}
	// PERF: Avoid parsing RawQuery if "trace=" is not present
	if strings.Contains(r.URL.RawQuery, "trace=") {
		v := r.URL.Query().Get(traceQuery)
		b, _ := strconv.ParseBool(v)
		return b
	}
	return false
}
