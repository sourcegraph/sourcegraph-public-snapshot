package debugproxies

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/router"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestReverseProxyRequestPaths(t *testing.T) {
	var rph ReverseProxyHandler

	proxiedServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Path))
	}))
	defer proxiedServer.Close()

	proxiedURL, err := url.Parse(proxiedServer.URL)
	if err != nil {
		t.Errorf("setup error %v", err)
		return
	}

	featureFlags := dbmocks.NewMockFeatureFlagStore()
	featureFlags.GetFeatureFlagFunc.SetDefaultReturn(nil, sql.ErrNoRows)

	db := dbmocks.NewStrictMockDB()
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

	ep := Endpoint{Service: "gitserver", Addr: proxiedURL.Host}
	displayName := displayNameFromEndpoint(ep)
	rph.Populate(db, []Endpoint{ep})

	ctx := actor.WithInternalActor(context.Background())

	link := fmt.Sprintf("%s/-/debug/proxies/%s/metrics", proxiedServer.URL, displayName)
	req := httptest.NewRequest("GET", link, nil)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	rtr := mux.NewRouter()
	rtr.PathPrefix("/-/debug").Name(router.Debug)
	rph.AddToRouter(rtr.Get(router.Debug).Subrouter(), db)

	rtr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if string(body) != "/metrics" {
		t.Errorf("expected /metrics to be passed to reverse proxy, got %s", body)
	}
}

func TestIndexLinks(t *testing.T) {
	var rph ReverseProxyHandler

	proxiedServer := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, _ = writer.Write([]byte(request.URL.Path))
	}))
	defer proxiedServer.Close()

	proxiedURL, err := url.Parse(proxiedServer.URL)
	if err != nil {
		t.Errorf("setup error %v", err)
		return
	}

	ep := Endpoint{Service: "gitserver", Addr: proxiedURL.Host}
	displayName := displayNameFromEndpoint(ep)
	rph.Populate(dbmocks.NewMockDB(), []Endpoint{ep})

	ctx := actor.WithInternalActor(context.Background())

	link := fmt.Sprintf("%s/-/debug/", proxiedServer.URL)
	req := httptest.NewRequest("GET", link, nil)

	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	featureFlags := dbmocks.NewMockFeatureFlagStore()
	featureFlags.GetFeatureFlagFunc.SetDefaultReturn(nil, sql.ErrNoRows)

	db := dbmocks.NewStrictMockDB()
	db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

	rtr := mux.NewRouter()
	rtr.PathPrefix("/-/debug").Name(router.Debug)
	rph.AddToRouter(rtr.Get(router.Debug).Subrouter(), db)

	rtr.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedContent := fmt.Sprintf("<a href=\"proxies/%s/\">%s</a><br>", displayName, displayName)

	if !strings.Contains(string(body), expectedContent) {
		t.Errorf("expected %s, got %s", expectedContent, body)
	}
}

func TestDisplayNameFromEndpoint(t *testing.T) {
	cases := []struct {
		Service, Addr, Hostname string
		Want                    string
	}{{
		Service:  "gitserver",
		Addr:     "192.168.10.0:2323",
		Hostname: "gitserver-0",
		Want:     "gitserver-0",
	}, {
		Service: "searcher",
		Addr:    "192.168.10.3:2323",
		Want:    "searcher-192.168.10.3",
	}, {
		Service: "no-port",
		Addr:    "192.168.10.1",
		Want:    "no-port-192.168.10.1",
	}}

	for _, c := range cases {
		got := displayNameFromEndpoint(Endpoint{
			Service:  c.Service,
			Addr:     c.Addr,
			Hostname: c.Hostname,
		})
		if got != c.Want {
			t.Errorf("displayNameFromEndpoint(%q, %q) mismatch (-want +got):\n%s", c.Service, c.Addr, cmp.Diff(c.Want, got))
		}
	}
}

func TestAdminOnly(t *testing.T) {
	tests := []struct {
		name             string
		mockUsers        func(users *dbmocks.MockUserStore)
		mockFeatureFlags func(featureFlags *dbmocks.MockFeatureFlagStore)
		mockActor        *actor.Actor
		wantStatus       int
	}{
		{
			name: "not an admin",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1}, nil)
			},
			mockFeatureFlags: func(featureFlags *dbmocks.MockFeatureFlagStore) {
				featureFlags.GetFeatureFlagFunc.SetDefaultReturn(nil, sql.ErrNoRows)
			},
			mockActor:  &actor.Actor{},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "no feature flag",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFeatureFlags: func(featureFlags *dbmocks.MockFeatureFlagStore) {
				featureFlags.GetFeatureFlagFunc.SetDefaultReturn(nil, sql.ErrNoRows)
			},
			mockActor:  &actor.Actor{},
			wantStatus: http.StatusOK,
		},
		{
			name: "has feature flag but not enabled",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFeatureFlags: func(featureFlags *dbmocks.MockFeatureFlagStore) {
				featureFlags.GetFeatureFlagFunc.SetDefaultReturn(&featureflag.FeatureFlag{Bool: &featureflag.FeatureFlagBool{Value: false}}, nil)
			},
			mockActor:  &actor.Actor{},
			wantStatus: http.StatusOK,
		},
		{
			name: "feature flag enabled but not Sourcegraph operator",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFeatureFlags: func(featureFlags *dbmocks.MockFeatureFlagStore) {
				featureFlags.GetFeatureFlagFunc.SetDefaultReturn(&featureflag.FeatureFlag{Bool: &featureflag.FeatureFlagBool{Value: true}}, nil)
			},
			mockActor:  &actor.Actor{},
			wantStatus: http.StatusForbidden,
		},
		{
			name: "feature flag enabled and Sourcegraph operator",
			mockUsers: func(users *dbmocks.MockUserStore) {
				users.GetByCurrentAuthUserFunc.SetDefaultReturn(&types.User{ID: 1, SiteAdmin: true}, nil)
			},
			mockFeatureFlags: func(featureFlags *dbmocks.MockFeatureFlagStore) {
				featureFlags.GetFeatureFlagFunc.SetDefaultReturn(&featureflag.FeatureFlag{Bool: &featureflag.FeatureFlagBool{Value: true}}, nil)
			},
			mockActor:  &actor.Actor{SourcegraphOperator: true},
			wantStatus: http.StatusOK,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			users := dbmocks.NewMockUserStore()
			test.mockUsers(users)

			featureFlags := dbmocks.NewMockFeatureFlagStore()
			test.mockFeatureFlags(featureFlags)

			db := dbmocks.NewMockDB()
			db.UsersFunc.SetDefaultReturn(users)
			db.FeatureFlagsFunc.SetDefaultReturn(featureFlags)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/-/debug", nil)
			r = r.WithContext(actor.WithActor(r.Context(), test.mockActor))
			AdminOnly(
				db,
				http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
			).ServeHTTP(w, r)

			assert.Equal(t, test.wantStatus, w.Code)
		})
	}
}
