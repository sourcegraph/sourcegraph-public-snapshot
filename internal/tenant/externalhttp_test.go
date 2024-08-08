package tenant

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestExternalTenantFromHostnameMiddleware(t *testing.T) {
	t.Run("valid tenant", func(t *testing.T) {
		mapper := func(ctx context.Context, host string) (int, error) {
			return 42, nil
		}

		handler := ExternalTenantFromHostnameMiddleware(mapper, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, err := FromContext(r.Context())
			require.NoError(t, err)
			require.Equal(t, 42, tenant.ID())
		}))

		req := httptest.NewRequest("GET", "http://example.com", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})

	t.Run("tenant not found", func(t *testing.T) {
		mapper := func(ctx context.Context, host string) (int, error) {
			return 0, &notFoundError{}
		}

		handler := ExternalTenantFromHostnameMiddleware(mapper, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("This handler should not be called")
		}))

		req := httptest.NewRequest("GET", "http://unknown.com", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
		require.Contains(t, rr.Body.String(), "tenant \"unknown.com\" not known")
	})

	t.Run("internal error", func(t *testing.T) {
		mapper := func(ctx context.Context, host string) (int, error) {
			return 0, errors.New("database error")
		}

		handler := ExternalTenantFromHostnameMiddleware(mapper, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Fatal("This handler should not be called")
		}))

		req := httptest.NewRequest("GET", "http://example.com", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusInternalServerError, rr.Code)
		require.Contains(t, rr.Body.String(), "failed to fetch tenant")
	})

	t.Run("host with port", func(t *testing.T) {
		mapper := func(ctx context.Context, host string) (int, error) {
			require.Equal(t, "example.com", host)
			return 42, nil
		}

		handler := ExternalTenantFromHostnameMiddleware(mapper, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenant, err := FromContext(r.Context())
			require.NoError(t, err)
			require.Equal(t, 42, tenant.ID())
		}))

		req := httptest.NewRequest("GET", "http://example.com:8080", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
	})
}

type notFoundError struct{}

func (e notFoundError) Error() string {
	return "not found"
}

func (e notFoundError) NotFound() bool {
	return true
}

func TestExtractHost(t *testing.T) {
	t.Run("host without port", func(t *testing.T) {
		host := extractHost("example.com")
		require.Equal(t, "example.com", host)
	})

	t.Run("host with port", func(t *testing.T) {
		host := extractHost("example.com:8080")
		require.Equal(t, "example.com", host)
	})

	t.Run("empty string", func(t *testing.T) {
		host := extractHost("")
		require.Equal(t, "", host)
	})

	t.Run("multiple colons", func(t *testing.T) {
		host := extractHost("a:b:c:8080")
		require.Equal(t, "a:b:c", host)
	})
}
