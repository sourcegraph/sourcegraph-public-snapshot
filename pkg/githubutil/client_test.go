package githubutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil"
)

// Test that we don't use cached responses from requests authenticated
// as user A when making requests authenticated as user B.
func TestGitHubClient_noCacheLeak(t *testing.T) {
	handled := 0
	responseLogin := "a"
	mux := http.NewServeMux()
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		// Mimic the Vary and Cache-Control headers GitHub adds
		w.Header().Add("cache-control", "private, max-age=60, s-maxage=60")
		w.Header().Add("vary", "Accept-Encoding")
		w.Header().Add("vary", "Accept, Authorization, Cookie, X-GitHub-OTP")
		w.Write([]byte(`{"login":"` + responseLogin + `"}`))
		handled++
	})
	s := httptest.NewServer(mux)
	defer s.Close()
	baseURL, _ := url.Parse(s.URL)

	ghconf := &Config{Cache: httputil.Cache}

	ghA := ghconf.AuthedClient("token-for-a")
	ghB := ghconf.AuthedClient("token-for-b")
	ghA.BaseURL = baseURL
	ghB.BaseURL = baseURL

	userA, _, err := ghA.Users.Get("")
	if err != nil {
		t.Fatal(err)
	}
	if handled != 1 {
		t.Fatal("handler not called")
	}
	if want := "a"; *userA.Login != want {
		t.Errorf("got userA.Login == %q, want %q", *userA.Login, want)
	}

	responseLogin = "b"
	userB, _, err := ghB.Users.Get("")
	if err != nil {
		t.Fatal(err)
	}
	if handled != 2 {
		t.Errorf("handler not called")
	}
	if want := "b"; *userB.Login != want {
		t.Errorf("got userB.Login == %q, want %q", *userB.Login, want)
	}

}
