package auth

// pre-commit:ignore_sourcegraph_token
import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	mockrequire "github.com/derision-test/go-mockgen/v2/testutil/require"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor/anonymous"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor/productsubscription"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/actor/productsubscription/productsubscriptiontest"
	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/events"
	"github.com/sourcegraph/sourcegraph/internal/codygateway/codygatewayactor"
	codyaccessv1 "github.com/sourcegraph/sourcegraph/lib/enterpriseportal/codyaccess/v1"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TODO(@bobheadxi): Try to rewrite this as a table-driven test for less copy-pasta.
func TestAuthenticatorMiddleware(t *testing.T) {
	logger := logtest.Scoped(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })

	concurrencyConfig := codygatewayactor.ActorConcurrencyLimitConfig{Percentage: 50, Interval: time.Hour}

	t.Run("unauthenticated and allow anonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(anonymous.NewSource(true, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("unauthenticated but disallow anonymous", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(anonymous.NewSource(false, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("authenticated without cache hit", func(t *testing.T) {
		cache := NewMockListingCache()
		client := productsubscriptiontest.NewMockEnterprisePortalClient()
		client.GetCodyGatewayAccessFunc.PushReturn(
			&codyaccessv1.GetCodyGatewayAccessResponse{
				Access: &codyaccessv1.CodyGatewayAccess{
					SubscriptionId: "es_6452a8fc-e650-45a7-a0a2-357f776b3b46",
					Enabled:        true,
					ChatCompletionsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
						Limit:            10,
						IntervalDuration: durationpb.New(10 * time.Second),
					},
					CodeCompletionsRateLimit: &codyaccessv1.CodyGatewayRateLimit{
						Limit:            10,
						IntervalDuration: durationpb.New(10 * time.Second),
					},
				},
			},
			nil,
		)

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, actor.FromContext(r.Context()))
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockrequire.Called(t, client.GetCodyGatewayAccessFunc)
	})

	t.Run("authenticated with cache hit", func(t *testing.T) {
		cache := NewMockListingCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":true,"rateLimit":null}`),
			true,
		)
		client := productsubscriptiontest.NewMockEnterprisePortalClient()
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.NotNil(t, actor.FromContext(r.Context()))
			w.WriteHeader(http.StatusOK)
		})

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
		mockrequire.NotCalled(t, client.GetCodyGatewayAccessFunc)
	})

	t.Run("authenticated but not enabled", func(t *testing.T) {
		cache := NewMockListingCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":false,"rateLimit":null}`),
			true,
		)
		client := productsubscriptiontest.NewMockEnterprisePortalClient()

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("authenticated bypassing attribution while not enabled", func(t *testing.T) {
		cache := NewMockListingCache()
		cache.GetFunc.SetDefaultReturn(
			[]byte(`{"id":"UHJvZHVjdFN1YnNjcmlwdGlvbjoiNjQ1MmE4ZmMtZTY1MC00NWE3LWEwYTItMzU3Zjc3NmIzYjQ2Ig==","accessEnabled":false,"endpointAccess":{"/v1/attribution":true},"rateLimit":null}`),
			true,
		)
		client := productsubscriptiontest.NewMockEnterprisePortalClient()

		t.Run("bypass works for attribution", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/v1/attribution", strings.NewReader(`{}`))
			r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
			(&Authenticator{
				Logger:      logger,
				EventLogger: events.NewStdoutLogger(logger),
				Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
			}).Middleware(next).ServeHTTP(w, r)
			assert.Equal(t, http.StatusOK, w.Code)
		})

		t.Run("bypass does not work for other endpoints", func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/v1/completions", strings.NewReader(`{}`))
			r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
			(&Authenticator{
				Logger:      logger,
				EventLogger: events.NewStdoutLogger(logger),
				Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
			}).Middleware(next).ServeHTTP(w, r)
			assert.Equal(t, http.StatusForbidden, w.Code)
		})
	})

	t.Run("access token denied from sources", func(t *testing.T) {
		cache := NewMockListingCache()
		client := productsubscriptiontest.NewMockEnterprisePortalClient()
		client.GetCodyGatewayAccessFunc.PushReturn(
			nil,
			status.Error(codes.NotFound, "not found"),
		)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockrequire.Called(t, client.GetCodyGatewayAccessFunc)
	})

	t.Run("server error from sources", func(t *testing.T) {
		cache := NewMockListingCache()
		client := productsubscriptiontest.NewMockEnterprisePortalClient()
		client.GetCodyGatewayAccessFunc.SetDefaultReturn(
			nil,
			errors.New("server error"),
		)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{}`))
		r.Header.Set("Authorization", "Bearer sgs_abc1228e23e789431f08cd15e9be20e69b8694c2dff701b81d16250a4a861f37")
		(&Authenticator{
			Logger:      logger,
			EventLogger: events.NewStdoutLogger(logger),
			Sources:     actor.NewSources(productsubscription.NewSource(logger, cache, client, concurrencyConfig)),
		}).Middleware(next).ServeHTTP(w, r)
		assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	})
}
