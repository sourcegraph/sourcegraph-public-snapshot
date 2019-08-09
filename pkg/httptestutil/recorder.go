package httptestutil

import (
	"net/http"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
)

// NewRecorder returns an HTTP interaction recorder with the given record mode and filters. It strips away the HTTP Authorization and Set-Cookie headers.
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_837(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
