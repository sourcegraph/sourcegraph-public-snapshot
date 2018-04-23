package auth

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/session"
	"github.com/sourcegraph/sourcegraph/pkg/actor"
)

func TestAllowAnonymousRequest(t *testing.T) {
	req := func(method, urlStr string) *http.Request {
		r, err := http.NewRequest(method, urlStr, nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	tests := []struct {
		req  *http.Request
		want bool
	}{
		{req: req("GET", "/"), want: false},
		{req: req("POST", "/"), want: false},
		{req: req("POST", "/-/sign-in"), want: true},
		{req: req("GET", "/sign-in"), want: true},
		{req: req("GET", "/doesntexist"), want: false},
		{req: req("POST", "/doesntexist"), want: false},
		{req: req("GET", "/doesnt/exist"), want: false},
		{req: req("POST", "/doesnt/exist"), want: false},
		{req: req("POST", "/.api/telemetry/log/v1/production"), want: true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.req.Method, test.req.URL), func(t *testing.T) {
			got := allowAnonymousRequest(test.req)
			if got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestNewUserRequiredAuthzHandler(t *testing.T) {
	cleanup := session.ResetMockSessionStore(t)
	defer cleanup()

	withAuth := func(r *http.Request) *http.Request {
		w := httptest.NewRecorder()
		if err := session.StartNewSession(w, httptest.NewRequest("GET", "/", nil), &actor.Actor{UID: 123}, time.Hour); err != nil {
			t.Fatal(err)
		}
		for _, cookie := range w.Result().Cookies() {
			if cookie.Expires.After(time.Now()) || cookie.MaxAge > 0 {
				r.AddCookie(cookie)
			}
		}
		return r
	}

	testcases := []struct {
		name       string
		req        *http.Request
		allowed    bool
		wantStatus int
		location   string
	}{
		{
			name:       "no_auth__private_route",
			req:        httptest.NewRequest("GET", "/", nil),
			allowed:    false,
			wantStatus: http.StatusFound,
			location:   "/sign-in?returnTo=%2F",
		},
		{
			name:       "no_auth__public_route",
			req:        httptest.NewRequest("GET", "/sign-in", nil),
			allowed:    true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "auth__private_route",
			req:        withAuth(httptest.NewRequest("GET", "/", nil)),
			allowed:    true,
			wantStatus: http.StatusOK,
		},
		{
			name:       "auth__public_route",
			req:        withAuth(httptest.NewRequest("GET", "/sign-in", nil)),
			allowed:    true,
			wantStatus: http.StatusOK,
		},
	}
	for _, tst := range testcases {
		t.Run(tst.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			allowed := false
			newUserRequiredAuthzHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				allowed = true
			})).ServeHTTP(rec, tst.req)
			if allowed != tst.allowed {
				t.Fatalf("got request allowed %v want %v", allowed, tst.allowed)
			}
			if status := rec.Result().StatusCode; status != tst.wantStatus {
				t.Fatalf("got status code %v want %v", status, tst.wantStatus)
			}
			loc := rec.Result().Header.Get("Location")
			if loc != tst.location {
				t.Fatalf("got location %q want %q", loc, tst.location)
			}
		})
	}
}
