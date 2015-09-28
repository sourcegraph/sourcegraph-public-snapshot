package browser

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"sourcegraph.com/sourcegraph/surf/agent"
	"sourcegraph.com/sourcegraph/surf/jar"
)

func newDefaultTestBrowser() *Browser {
	bow := &Browser{}
	bow.SetUserAgent(agent.Create())
	bow.SetState(&jar.State{})
	bow.SetCookieJar(jar.NewMemoryCookies())
	bow.SetBookmarksJar(jar.NewMemoryBookmarks())
	bow.SetHistoryJar(jar.NewMemoryHistory())
	bow.SetHeadersJar(jar.NewMemoryHeaders())
	bow.SetAttributes(AttributeMap{
		SendReferer:         true,
		MetaRefreshHandling: true,
		FollowRedirects:     true,
	})
	return bow
}

// TestRedirect
// See: https://github.com/headzoo/surf/pull/18
func TestRedirect(t *testing.T) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			io.WriteString(w, `<html><body>
<form class="foo-form" action="/foo" method="POST">
	<input type="text" name="bar">
</form>
</body></html>`)
		case "/foo":
			http.Error(w, "Unimplemented", 500)
			return
		}
	}))
	defer ts0.Close()

	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			t.Logf("Redirecting to %s", ts0.URL)
			w.Header().Set("Location", ts0.URL)
			w.WriteHeader(302)
		default:
			http.Error(w, "Not found", 404)
			return
		}
	}))
	defer ts1.Close()

	// First, a sanity check using the default http.Client
	client := http.Client{
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return nil
		},
	}
	resp, err := client.Get(ts1.URL)
	if err != nil {
		t.Errorf("(sanity) Failed to open url: %s", ts1.URL)
		return
	}

	if resp.Request.URL.String() != ts0.URL {
		t.Errorf("Expected redirect to have happened")
		return
	}
	// If we got here, then we know that redirects are correctly working

	// Alright, now let's see if the browser does the same thing
	b := newDefaultTestBrowser()

	if err := b.Open(ts1.URL); err != nil {
		t.Errorf("Failed to open url: %s", ts1.URL)
		return
	}

	if b.Url().String() != ts0.URL {
		t.Errorf("Error: Redirects are not being recorded?")
		return
	}
}

// TestCookieHeader ensures that headers are not shared/merged across
// requests.
func TestCookieHeader(t *testing.T) {
	calls := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++

		cookie, err := r.Cookie("c")
		if err == http.ErrNoCookie {
			err = nil
		}
		if err != nil {
			t.Fatal(err)
		}

		switch r.URL.Path {
		case "/cookie":
			http.SetCookie(w, &http.Cookie{
				Name: "c", Value: "v",
				Expires: time.Now().Add(time.Hour),
				Path:    "/cookie",
			})
			if r.URL.Query().Get("check") != "" && cookie == nil {
				t.Errorf("got no cookie")
			}

		case "/":
			if cookie != nil {
				t.Errorf("got cookie %v, want no cookie", cookie)
			}
		}
	}))
	defer ts.Close()

	b := newDefaultTestBrowser()
	if err := b.Open(ts.URL + "/cookie"); err != nil {
		t.Fatal(err)
	}
	if err := b.Open(ts.URL + "/cookie?check=1"); err != nil {
		t.Fatal(err)
	}
	if err := b.Open(ts.URL + "/"); err != nil {
		t.Fatal(err)
	}

	if want := 3; calls != want {
		t.Errorf("got %d calls, want %d", calls, want)
	}
}
