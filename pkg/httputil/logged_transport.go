package httputil

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

func init() {
	if v, _ := strconv.ParseBool(env.Get("SG_HTTP_TRACE", "false", "additional logging for each HTTP request")); v {
		http.DefaultTransport = &LoggedTransport{
			Writer:    os.Stderr,
			Transport: http.DefaultTransport,
		}
	}
}

// A LoggedTransport prints URLs and timings for each HTTP request.
type LoggedTransport struct {
	io.Writer                   // destination of output
	Transport http.RoundTripper // underlying transport (or default if nil)
}

func (t *LoggedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var u http.RoundTripper
	if t.Transport != nil {
		u = t.Transport
	} else {
		u = http.DefaultTransport
	}

	start := time.Now()

	resp, err := u.RoundTrip(req)
	if err != nil {
		fmt.Fprintf(t.Writer, "HTTP %s %s: error: %s\n", req.Method, req.URL, err)
		return nil, err
	}

	fmt.Fprintf(t.Writer, "HTTP %s %s %d [%s rtt]\n", req.Method, req.URL, resp.StatusCode, time.Since(start))

	return resp, nil
}
