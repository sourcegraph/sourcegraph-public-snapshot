package eventsutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func TestAgentMiddleware(t *testing.T) {
	var called bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpctx.SetForRequest(r, context.Background())
		AgentMiddleware(w, r, func(w http.ResponseWriter, r *http.Request) {
			called = true
			ctx := httpctx.FromRequest(r)

			userAgent := UserAgentFromContext(ctx)
			if want := "sourcegraphbot"; userAgent != want {
				t.Errorf("got User-Agent %q, want %q", userAgent, want)
			}
		})
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	req.Header.Set("User-Agent", "sourcegraphbot")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !called {
		t.Error("!called")
	}
}

type cookieJar struct {
	Cs []*http.Cookie
}

func (jar *cookieJar) Cookies(u *url.URL) []*http.Cookie {
	return jar.Cs
}

func (jar *cookieJar) SetCookies(u *url.URL, cookies []*http.Cookie) {} // do nothing

func TestDeviceIdMiddleware(t *testing.T) {
	var called bool
	var deviceId string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpctx.SetForRequest(r, context.Background())
		DeviceIdMiddleware(w, r, func(w http.ResponseWriter, r *http.Request) {
			called = true
			ctx := httpctx.FromRequest(r)
			deviceId = DeviceIdFromContext(ctx)
		})
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if !called {
		t.Error("!called")
	}
	if deviceId == "" {
		t.Error("!deviceId")
	}

	oldDeviceId := deviceId

	// Make another request to verify the device id is the same as before
	jar := cookieJar{Cs: resp.Cookies()}
	c := http.Client{Jar: &jar}
	resp, err = c.Get(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if oldDeviceId != deviceId {
		t.Error("oldDeviceId != deviceId")
	}
}
