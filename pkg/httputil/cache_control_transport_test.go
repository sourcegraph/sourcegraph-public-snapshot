package httputil

import (
	"bytes"
	"net/http"
	"testing"
)

type recorderTransport struct {
	req *http.Request
}

func (t *recorderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.req = req
	return nil, nil
}

func TestCacheControlTransport(t *testing.T) {
	url1 := "https://foo.com/bar?baz"
	url2 := "https://bar.com/bar?baz"
	url3 := "https://foo.com/long-lived"

	getReq, _ := http.NewRequest("GET", url1, bytes.NewBuffer(nil))
	getReq2, _ := http.NewRequest("GET", url2, bytes.NewBuffer(nil))
	getReq3, _ := http.NewRequest("GET", url3, bytes.NewBuffer(nil))
	postReq, _ := http.NewRequest("POST", url1, bytes.NewBuffer(nil))

	recorder := &recorderTransport{}
	ccTransport := NewCacheControlTransport("max-age=0", recorder, func(r *http.Request) bool {
		return !(r.URL.String() == url3)
	})

	// Cache-control round-tripper
	ccTransport.RoundTrip(getReq)
	if cc := recorder.req.Header.Get("Cache-Control"); cc != "max-age=0" {
		t.Errorf("expected cache control header on GET, but got none")
	}

	ccTransport.RoundTrip(getReq)
	if cc := recorder.req.Header.Get("Cache-Control"); cc != "" {
		t.Errorf("expected no cache control header on 2nd GET, but got %s", cc)
	}

	ccTransport.RoundTrip(getReq2)
	if cc := recorder.req.Header.Get("Cache-Control"); cc != "max-age=0" {
		t.Errorf("expected cache control header on GET to new URL, but got none")
	}

	ccTransport.RoundTrip(getReq3)
	if cc := recorder.req.Header.Get("Cache-Control"); cc != "" {
		t.Errorf("expected no cache control header on GET to long-lived URL, but got %s", cc)
	}

	ccTransport.RoundTrip(postReq)
	if cc := recorder.req.Header.Get("Cache-Control"); cc != "" {
		t.Errorf("expected no cache control header on POST, but got %s", cc)
	}

	// Pass-through if no cache control set
	noCCTransport := NewCacheControlTransport("", recorder, nil)
	for _, meth := range []string{"GET", "PUT", "PATCH", "DELETE", "POST"} {
		req, _ := http.NewRequest(meth, url1, bytes.NewBuffer(nil))
		noCCTransport.RoundTrip(req)
		if cc := recorder.req.Header.Get("Cache-Control"); cc != "" {
			t.Errorf("expected no cache control on %s (because none set on transport), but got %s", meth, cc)
		}
	}
}
