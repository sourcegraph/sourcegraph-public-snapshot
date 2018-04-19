package auth

import (
	"net/http"
	"testing"
	"time"
)

// check checks if condition is true and errors with errMsg if false.
func check(t *testing.T, condition bool, errMsg string) {
	t.Helper()
	if !condition {
		t.Error(errMsg)
	}
}

// unexpiredCookies returns the list of unexpired cookies set by the response
func unexpiredCookies(resp *http.Response) (cookies []*http.Cookie) {
	for _, cookie := range resp.Cookies() {
		if cookie.RawExpires == "" || cookie.Expires.After(time.Now()) {
			cookies = append(cookies, cookie)
		}
	}
	return
}
