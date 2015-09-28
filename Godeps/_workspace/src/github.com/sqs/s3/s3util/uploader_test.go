package s3util

import (
	"io"
	"io/ioutil"
	"net/http"
	"runtime"
	"strings"
	"testing"
)

func runUpload(t *testing.T, makeCloser func(io.Reader) io.ReadCloser) *uploader {
	c := *DefaultConfig
	c.Client = &http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			var s string
			switch q := req.URL.Query(); {
			case req.Method == "PUT":
			case req.Method == "POST" && q["uploads"] != nil:
				s = `<UploadId>foo</UploadId>`
			case req.Method == "POST" && q["uploadId"] != nil:
				b, _ := ioutil.ReadAll(req.Body)
				if !strings.Contains(string(b), "<ETag>foo</ETag>") {
					t.Error("missing ETag")
				}
			default:
				t.Fatal("unexpected request", req)
			}
			resp := &http.Response{
				StatusCode: 200,
				Body:       makeCloser(strings.NewReader(s)),
				Header: http.Header{
					"Etag": {`"foo"`},
				},
			}
			return resp, nil
		}),
	}
	u, err := newUploader("https://s3.amazonaws.com/foo/bar", nil, &c)
	if err != nil {
		t.Fatal("unexpected err", err)
	}
	const size = minPartSize + minPartSize/3
	n, err := io.Copy(u, io.LimitReader(devZero, size))
	if err != nil {
		t.Fatal("unexpected err", err)
	}
	if n != size {
		t.Fatal("wrote %d bytes want %d", n, size)
	}
	err = u.Close()
	if err != nil {
		t.Fatal("unexpected err", err)
	}
	return u
}

func TestUploaderCloseRespBody(t *testing.T) {
	want := make(chan int, 100)
	got := make(closeCounter, 100)
	f := func(r io.Reader) io.ReadCloser {
		want <- 1
		return readClose{r, got}
	}
	runUpload(t, f)
	if len(want) != len(got) {
		t.Errorf("closes = %d want %d", len(got), len(want))
	}
}

// Used in TestUploaderFreesBuffers to force liveness.
var DummyUploader *uploader

func TestUploaderFreesBuffers(t *testing.T) {
	var m0, m1 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m0)
	u := runUpload(t, ioutil.NopCloser)
	runtime.GC()
	runtime.ReadMemStats(&m1)

	// Make sure everything reachable from u is still live while reading m1.
	// (Very aggressive cross-package optimization could hypothetically
	// break this, rendering the test ineffective.)
	DummyUploader = u

	// The uploader never allocates buffers smaller than minPartSize,
	// so if the increase is < minPartSize we know none are reachable.
	inc := m1.Alloc - m0.Alloc
	if m1.Alloc > m0.Alloc && inc >= minPartSize {
		t.Errorf("inc = %d want <%d", inc, minPartSize)
	}
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (f RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type closeCounter chan int

func (c closeCounter) Close() error {
	c <- 1
	return nil
}

type readClose struct {
	io.Reader
	io.Closer
}

var devZero io.Reader = repeatReader(0)

type repeatReader byte

func (r repeatReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = byte(r)
	}
	return len(p), nil
}

func TestEmptyEtag(t *testing.T) {
	defer func() {
		if err := recover(); err != nil {
			t.Fatal("uploader panic", err)
		}
	}()
	body := readClose{
		strings.NewReader(`<UploadId>foo</UploadId>`),
		make(closeCounter, 100),
	}
	c := *DefaultConfig
	c.Client = &http.Client{
		Transport: RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp := &http.Response{
				StatusCode: 200,
				Body:       body,
				Header: http.Header{
					"Etag": {""},
				},
			}
			return resp, nil
		}),
	}
	u, err := newUploader("https://s3.amazonaws.com/foo/bar", nil, &c)
	if err != nil {
		t.Fatal("unexpected err", err)
	}
	const size = minPartSize + minPartSize/3
	n, err := io.Copy(u, io.LimitReader(devZero, size))
	if err == nil || err.Error() != `received invalid etag ""` {
		t.Fatalf("expected err: %q", err)
	}
	if n != minPartSize {
		t.Fatalf("wrote %d bytes want %d", n, minPartSize)
	}
	err = u.Close()
	if err == nil || err.Error() != `received invalid etag ""` {
		t.Fatalf("expected err: %q", err)
	}
}
