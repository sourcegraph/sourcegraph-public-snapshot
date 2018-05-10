package handlerutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPSRedirectLoadBalanced(t *testing.T) {
	cases := []struct {
		URL         string
		HeaderKey   string
		HeaderValue string
		Redirect    bool
	}{
		{
			URL:      "http://sourcegraph.com/foo",
			Redirect: true,
		},
		{
			URL:      "https://sourcegraph.com/foo",
			Redirect: false,
		},
		{
			URL:         "http://sourcegraph.com/foo",
			HeaderKey:   "X-Forwarded-Proto",
			HeaderValue: "http",
			Redirect:    true,
		},
		{
			URL:         "http://sourcegraph.com/foo",
			HeaderKey:   "X-Forwarded-Proto",
			HeaderValue: "https",
			Redirect:    false,
		},
		{
			URL:         "https://sourcegraph.com/foo",
			HeaderKey:   "X-Forwarded-Proto",
			HeaderValue: "http",
			Redirect:    true,
		},
		{
			URL:         "https://sourcegraph.com/foo",
			HeaderKey:   "X-Forwarded-Proto",
			HeaderValue: "https",
			Redirect:    false,
		},
	}

	for _, cc := range cases {
		req, err := http.NewRequest("GET", cc.URL, nil)
		if err != nil {
			t.Fatal(err)
		}
		if cc.HeaderKey != "" {
			req.Header.Set(cc.HeaderKey, cc.HeaderValue)
		}

		redirect := true
		h := HTTPSRedirectLoadBalanced(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			redirect = false
		}))
		h.ServeHTTP(httptest.NewRecorder(), req)

		if redirect != cc.Redirect {
			t.Errorf("got redirect=%v for %+v", redirect, cc)
		}
	}
}
