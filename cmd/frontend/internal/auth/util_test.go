package auth

import (
	"net/http"
	"reflect"
	"testing"
	"time"
)

// appURL is the URL of our fake OIDC-enabled app
const appURL = "http://my-app.com"

// check checks if condition is true and errors with errMsg if false.
func check(t *testing.T, condition bool, errMsg string) {
	if !condition {
		t.Error(errMsg)
	}
}

// checkEq checks for equality *if* the expected value is non-zero.
func checkEq(t *testing.T, expected, actual interface{}, errMsg string) {
	if expected == nil {
		return
	}
	if exp, ok := expected.(int); ok && exp == 0 {
		return
	}
	if exp, ok := expected.(string); ok && exp == "" {
		return
	}
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected %v, but got %v (%s)", expected, actual, errMsg)
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
