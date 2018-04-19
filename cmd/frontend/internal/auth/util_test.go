package auth

import (
	"net/http"
	"time"
)

// unexpiredCookies returns the list of unexpired cookies set by the response
func unexpiredCookies(resp *http.Response) (cookies []*http.Cookie) {
	for _, cookie := range resp.Cookies() {
		if cookie.RawExpires == "" || cookie.Expires.After(time.Now()) {
			cookies = append(cookies, cookie)
		}
	}
	return
}
