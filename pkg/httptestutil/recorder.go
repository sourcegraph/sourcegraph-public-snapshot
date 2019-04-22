package httptestutil

import (
	"net/http"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
)

// NewRecorder returns an HTTP interaction recorder
func NewRecorder(file string, record bool, filters ...cassette.Filter) (*recorder.Recorder, error) {
	mode := recorder.ModeReplaying
	if record {
		mode = recorder.ModeRecording
	}

	rec, err := recorder.NewAsMode(file, mode, nil)
	if err != nil {
		return nil, err
	}

	filters = append(filters, func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "Authorization")
		delete(i.Response.Headers, "Set-Cookie")
		return nil
	})

	for _, f := range filters {
		rec.AddFilter(f)
	}

	return rec, nil
}

// NewRecorderOpt returns an httpcli.Opt that wraps the Transport
// of an http.Client with the given recorder.
func NewRecorderOpt(rec *recorder.Recorder) httpcli.Opt {
	return func(c *http.Client) error {
		tr := c.Transport
		if tr == nil {
			tr = http.DefaultTransport
		}

		rec.SetTransport(tr)
		c.Transport = rec

		return nil
	}
}
