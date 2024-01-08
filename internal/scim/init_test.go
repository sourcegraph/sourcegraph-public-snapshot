package scim

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/license"
	"github.com/sourcegraph/sourcegraph/internal/licensing"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestAuthMiddleware(t *testing.T) {
	testHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	testHandler := scimAuthMiddleware(http.HandlerFunc(testHandlerFunc))

	testCases := []struct {
		name     string
		token    string
		config   *conf.Unified
		testFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:   "matching token",
			config: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimAuthToken: "testtoken"}},
			token:  "Bearer testtoken",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusOK, "request should be ok if tokens match")
			},
		},
		{
			name:   "no token",
			config: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimAuthToken: "testtoken"}},
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusUnauthorized, "unauthorized if no token")
			},
		},
		{
			name:   "not configured fails auth check",
			config: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{}},
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusUnauthorized, "unauthorized if not configured")
			},
		},
		{
			name:   "not matching token",
			config: &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimAuthToken: "testtoken"}},
			token:  "Bearer idontmatch",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusUnauthorized, "unauthorized if token doesnt match")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(tc.config)
			defer conf.Mock(nil)
			req := &http.Request{Header: make(http.Header)}
			if tc.token != "" {
				req.Header.Set("Authorization", tc.token)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)
			tc.testFunc(t, rr)
		})
	}

}

func TestLicenseMiddleware(t *testing.T) {
	testHandlerFunc := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
	testHandler := scimLicenseCheckMiddleware(http.HandlerFunc(testHandlerFunc))

	licensingInfo := func(tags ...string) *license.Info {
		return &license.Info{Tags: tags, ExpiresAt: time.Now().Add(1 * time.Hour)}
	}

	testCases := []struct {
		name     string
		license  *license.Info
		testFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:    "starter license should not have access",
			license: licensingInfo("starter"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusForbidden)
			},
		},
		{
			name:    "plan:free-1 should not have access",
			license: licensingInfo("plan:free-1"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusForbidden, "not allowed for free plan")
			},
		},
		{
			name:    "plan:business-0 should have access",
			license: licensingInfo("plan:business-0"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusOK, "allowed for business-0 plan")
			},
		},
		{
			name:    "plan:enterprise-1 should have access",
			license: licensingInfo("plan:enterprise-1"),
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusOK, "allowed for enterprise-1 plan")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return tc.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			req := &http.Request{Header: make(http.Header)}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)
			tc.testFunc(t, rr)
		})
	}

}

func TestHandler(t *testing.T) {
	db := getMockDB([]*types.UserForSCIM{}, map[int32][]*database.UserEmail{})
	testHandler := NewHandler(context.Background(), db, &observation.TestContext)

	licensingInfo := func(tags ...string) *license.Info {
		return &license.Info{Tags: tags, ExpiresAt: time.Now().Add(1 * time.Hour)}
	}

	testCases := []struct {
		name     string
		config   *conf.Unified
		token    string
		license  *license.Info
		testFunc func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name:    "auth should fail first",
			license: licensingInfo("starter"),
			config:  &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimAuthToken: "testtoken"}},
			token:   "Bearer idontmatch",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusUnauthorized)
			},
		},
		{
			name:    "license check should fail after auth passes",
			license: licensingInfo("starter"),
			config:  &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimAuthToken: "testtoken"}},
			token:   "Bearer testtoken",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusForbidden)
			},
		},
		{
			name:    "Auth & license check passes it runs",
			license: licensingInfo("plan:enterprise-1"),
			config:  &conf.Unified{SiteConfiguration: schema.SiteConfiguration{ScimAuthToken: "testtoken"}},
			token:   "Bearer testtoken",
			testFunc: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, rr.Result().StatusCode, http.StatusOK)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			conf.Mock(tc.config)
			defer conf.Mock(nil)
			licensing.MockGetConfiguredProductLicenseInfo = func() (*license.Info, string, error) {
				return tc.license, "test-signature", nil
			}
			defer func() { licensing.MockGetConfiguredProductLicenseInfo = nil }()
			req := &http.Request{Header: make(http.Header), Method: http.MethodGet, URL: &url.URL{Path: "/.api/scim/v2/Schemas"}}
			if tc.token != "" {
				req.Header.Set("Authorization", tc.token)
			}
			rr := httptest.NewRecorder()
			testHandler.ServeHTTP(rr, req)
			tc.testFunc(t, rr)
		})
	}

}
