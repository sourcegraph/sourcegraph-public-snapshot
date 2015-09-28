package nosurf

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
)

// A reader that always fails on Read()
// Suitable for testing the case of crypto/rand unavailability
type failReader struct{}

func (f failReader) Read(p []byte) (n int, err error) {
	err = errors.New("dummy error")
	return
}

func dummyGet() *http.Request {
	req, err := http.NewRequest("GET", "http://dum.my/", nil)
	if err != nil {
		panic(err)
	}
	return req
}

func succHand(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("success"))
}

// Returns a HandlerFunc
// that tests for the correct failure reason
func correctReason(t *testing.T, reason error) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		got := Reason(r)
		if got != reason {
			t.Errorf("CSRF check should have failed with the reason %#v,"+
				" but it failed with the reason %#v", reason, got)
		}
		// Writes the default failure code
		w.WriteHeader(FailureCode)
	}

	return http.HandlerFunc(fn)
}

// Gets a cookie with the specified name from the Response
// Returns nil on not finding a suitable cookie
func getRespCookie(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// Encodes a slice of key-value pairs to a form value string
func formBody(pairs [][]string) string {
	vals := url.Values{}
	for _, pair := range pairs {
		vals.Add(pair[0], pair[1])
	}

	return vals.Encode()
}

// The same as formBody(), but wraps the string in a Reader
func formBodyR(pairs [][]string) *strings.Reader {
	return strings.NewReader(formBody(pairs))
}
