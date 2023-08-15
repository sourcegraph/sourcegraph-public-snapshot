package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/ui"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAllowAnonymousRequest(t *testing.T) {
	ui.InitRouter(dbmocks.NewMockDB())
	// Ensure auth.public is false (be robust against some other tests having side effects that
	// change it, or changed defaults).
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthPublic: false, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}})
	defer conf.Mock(nil)

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
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.req.Method, test.req.URL), func(t *testing.T) {
			got := auth.AllowAnonymousRequest(test.req)
			if got != test.want {
				t.Errorf("got %v, want %v", got, test.want)
			}
		})
	}
}

func TestAllowAnonymousRequestWithAdditionalConfig(t *testing.T) {
	ui.InitRouter(dbmocks.NewMockDB())
	// Ensure auth.public is false (be robust against some other tests having side effects that
	// change it, or changed defaults).
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthPublic: false, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}})
	defer conf.Mock(nil)

	req := func(method, urlStr string) *http.Request {
		r, err := http.NewRequest(method, urlStr, nil)
		if err != nil {
			t.Fatal(err)
		}
		return r
	}

	tests := []struct {
		req                      *http.Request
		confAuthPublic           bool
		allowAnonymousContextKey *bool
		want                     bool
	}{
		{req: req("GET", "/"), confAuthPublic: false, allowAnonymousContextKey: nil, want: false},
		{req: req("GET", "/"), confAuthPublic: true, allowAnonymousContextKey: nil, want: false},
		{req: req("GET", "/"), confAuthPublic: false, allowAnonymousContextKey: pointers.Ptr(false), want: false},
		{req: req("GET", "/"), confAuthPublic: true, allowAnonymousContextKey: pointers.Ptr(false), want: false},
		{req: req("GET", "/"), confAuthPublic: false, allowAnonymousContextKey: pointers.Ptr(true), want: false},
		{req: req("GET", "/"), confAuthPublic: true, allowAnonymousContextKey: pointers.Ptr(true), want: true},
		{req: req("POST", "/"), confAuthPublic: false, allowAnonymousContextKey: nil, want: false},
		{req: req("POST", "/"), confAuthPublic: true, allowAnonymousContextKey: nil, want: false},
		{req: req("POST", "/"), confAuthPublic: false, allowAnonymousContextKey: pointers.Ptr(false), want: false},
		{req: req("POST", "/"), confAuthPublic: true, allowAnonymousContextKey: pointers.Ptr(false), want: false},
		{req: req("POST", "/"), confAuthPublic: false, allowAnonymousContextKey: pointers.Ptr(true), want: false},
		{req: req("POST", "/"), confAuthPublic: true, allowAnonymousContextKey: pointers.Ptr(true), want: true},

		{req: req("POST", "/-/sign-in"), confAuthPublic: false, allowAnonymousContextKey: nil, want: true},
		{req: req("POST", "/-/sign-in"), confAuthPublic: true, allowAnonymousContextKey: nil, want: true},
		{req: req("POST", "/-/sign-in"), confAuthPublic: false, allowAnonymousContextKey: pointers.Ptr(true), want: true},
		{req: req("POST", "/-/sign-in"), confAuthPublic: true, allowAnonymousContextKey: pointers.Ptr(true), want: true},
		{req: req("GET", "/sign-in"), confAuthPublic: false, allowAnonymousContextKey: nil, want: true},
		{req: req("GET", "/sign-in"), confAuthPublic: true, allowAnonymousContextKey: nil, want: true},
		{req: req("GET", "/sign-in"), confAuthPublic: false, allowAnonymousContextKey: pointers.Ptr(true), want: true},
		{req: req("GET", "/sign-in"), confAuthPublic: true, allowAnonymousContextKey: pointers.Ptr(true), want: true},
	}
	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s + auth.public=%v, allowAnonymousContext=%v", test.req.Method, test.req.URL, test.confAuthPublic, test.allowAnonymousContextKey), func(t *testing.T) {
			r := test.req
			if test.allowAnonymousContextKey != nil {
				r = r.WithContext(context.WithValue(r.Context(), auth.AllowAnonymousRequestContextKey, *test.allowAnonymousContextKey))
			}
			conf.Get().AuthPublic = test.confAuthPublic
			defer func() { conf.Get().AuthPublic = false }()

			got := auth.AllowAnonymousRequest(r)
			require.Equal(t, test.want, got)
		})
	}
}

func TestNewUserRequiredAuthzMiddleware(t *testing.T) {
	ui.InitRouter(dbmocks.NewMockDB())
	// Ensure auth.public is false (be robust against some other tests having side effects that
	// change it, or changed defaults).
	conf.Mock(&conf.Unified{SiteConfiguration: schema.SiteConfiguration{AuthPublic: false, AuthProviders: []schema.AuthProviders{{Builtin: &schema.BuiltinAuthProvider{}}}}})
	defer conf.Mock(nil)

	withAuth := func(r *http.Request) *http.Request {
		return r.WithContext(actor.WithActor(context.Background(), &actor.Actor{UID: 1}))
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
			name:       "no_auth__raw_route",
			req:        httptest.NewRequest("GET", "/test-repo/-/raw/README.md", nil),
			allowed:    false,
			wantStatus: http.StatusUnauthorized,
			location:   "/sign-in?returnTo=%2Ftest-repo%2F-%2Fraw%2FREADME.md",
		},
		{
			name:       "no_auth__api_route",
			req:        httptest.NewRequest("GET", "/.api/graphql", nil),
			allowed:    false,
			wantStatus: http.StatusUnauthorized,
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
			name:       "auth__api_route",
			req:        withAuth(httptest.NewRequest("GET", "/.api/graphql", nil)),
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
			setAllowedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { allowed = true })

			handler := http.NewServeMux()
			handler.Handle("/.api/", auth.RequireAuthMiddleware.API(setAllowedHandler))
			handler.Handle("/", auth.RequireAuthMiddleware.App(setAllowedHandler))
			handler.ServeHTTP(rec, tst.req)

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
