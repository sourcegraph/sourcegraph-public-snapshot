package telemetry

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

// Samples returns samples of recent telemetry payloads, which we let the site
// admins view so they know what telemetry consists of.
func Samples() []string {
	telemetrySamplesMu.Lock()
	defer telemetrySamplesMu.Unlock()

	samples := make([]string, 0, telemetrySampleCount)
	for i := 0; i < telemetrySampleCount; i++ {
		p := (telemetrySampleCursor - 1 + i + /* ensure nonnegative */ telemetrySampleCount) % telemetrySampleCount
		s := telemetrySamples[p]
		if s == "" {
			break // haven't collected the i'th sample yet
		}
		samples = append(samples, format(s))
	}
	return samples
}

// format indents the JSON of the HTTP request body.
func format(s string) string {
	req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(s)))
	if err != nil {
		return s
	}
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return s
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, body, "", "  "); err != nil {
		return s
	}
	req.Body = ioutil.NopCloser(&buf)
	req.ContentLength = int64(buf.Len())
	req.URL.Scheme, req.URL.Host = "https", "example.com" // needed for DumpRequestOut,
	out, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return s
	}
	return string(out)
}

const telemetrySampleCount = 7

var (
	telemetrySamplesMu    sync.Mutex
	telemetrySamples      [telemetrySampleCount]string
	telemetrySampleCursor int
)

// Sample records a telemetry sample from the given HTTP request to the
// telemetry service.
func Sample(req *http.Request) {
	telemetrySamplesMu.Lock()
	defer telemetrySamplesMu.Unlock()

	// Only sample a subset once we've filled our samples array.
	if telemetrySampleCursor > telemetrySampleCount {
		if rand.Intn(100) != 0 {
			return
		}
	}

	var sample string
	out, err := httputil.DumpRequestOut(req, true)
	if err == nil {
		sample = string(out)
	} else {
		sample = fmt.Sprintf("Error collecting telemetry sample\n%s", err)
	}
	telemetrySamples[telemetrySampleCursor%telemetrySampleCount] = sample
	telemetrySampleCursor++
}
