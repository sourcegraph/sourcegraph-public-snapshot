package httputil

import (
	"bytes"
	"compress/gzip"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

// gzip middleware taken from https://github.com/PuerkitoBio/ghost.
//
// See gzip.go for the license.

func TestGzipped(t *testing.T) {
	body := "This is the body"
	headers := []string{"gzip", "*", "gzip, deflate, sdch"}

	h := Gzip(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	for _, hdr := range headers {
		t.Logf("running with Accept-Encoding header %s", hdr)
		req, err := http.NewRequest("GET", s.URL, nil)
		if err != nil {
			panic(err)
		}
		req.Header.Set("Accept-Encoding", hdr)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			panic(err)
		}
		assertStatus(http.StatusOK, res.StatusCode, t)
		assertHeader("Content-Encoding", "gzip", res, t)
		assertGzippedBody([]byte(body), res, t)
	}
}

func TestNoGzip(t *testing.T) {
	body := "This is the body"

	h := Gzip(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}

func TestNoGzipOnFilter(t *testing.T) {
	body := "This is the body"

	h := Gzip(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "x/x")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), nil)
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}

func TestNoGzipOnCustomFilter(t *testing.T) {
	body := "This is the body"

	h := Gzip(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), func(w http.ResponseWriter, r *http.Request) bool {
		return false
	})
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "", res, t)
	assertBody([]byte(body), res, t)
}

func TestGzipOnCustomFilter(t *testing.T) {
	body := "This is the body"

	h := Gzip(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "x/x")
			_, err := w.Write([]byte(body))
			if err != nil {
				panic(err)
			}
		}), func(w http.ResponseWriter, r *http.Request) bool {
		return true
	})
	s := httptest.NewServer(h)
	defer s.Close()

	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept-Encoding", "gzip")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	assertStatus(http.StatusOK, res.StatusCode, t)
	assertHeader("Content-Encoding", "gzip", res, t)
	assertGzippedBody([]byte(body), res, t)
}

func assertStatus(ex, ac int, t *testing.T) {
	if ex != ac {
		t.Errorf("expected status code to be %d, got %d", ex, ac)
	}
}

func assertBody(ex []byte, res *http.Response, t *testing.T) {
	buf, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	if !bytes.Equal(ex, buf) {
		t.Errorf("expected body to be '%s' (%d), got '%s' (%d)", ex, len(ex), buf, len(buf))
	}
}

func assertGzippedBody(ex []byte, res *http.Response, t *testing.T) {
	gr, err := gzip.NewReader(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, gr)
	if err != nil {
		panic(err)
	}
	if !bytes.Equal(ex, buf.Bytes()) {
		t.Errorf("expected unzipped body to be '%s' (%d), got '%s' (%d)", ex, len(ex), buf.Bytes(), buf.Len())
	}
}

func assertHeader(hName, ex string, res *http.Response, t *testing.T) {
	hVal, ok := res.Header[hName]
	if (!ok || len(hVal) == 0) && len(ex) > 0 {
		t.Errorf("expected header %s to be %s, was not set", hName, ex)
	} else if len(hVal) > 0 && hVal[0] != ex {
		t.Errorf("expected header %s to be %s, got %s", hName, ex, hVal)
	}
}
