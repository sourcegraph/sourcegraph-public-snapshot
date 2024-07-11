package appliance

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log"
)

var appliance = &Appliance{
	jwtSecret: []byte("a-jwt-secret"),
	logger:    log.NoOp(),
}

func TestCheckAuthorization_CallsNextHandlerWhenValidJWTSupplied(t *testing.T) {
	validUntil := time.Now().Add(time.Hour).UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		jwtClaimsValidUntilKey: validUntil.Format(time.RFC3339),
	})
	tokenStr, err := token.SignedString(appliance.jwtSecret)
	require.NoError(t, err)

	req, err := http.NewRequest("GET", "example.com", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:    authCookieName,
		Value:   tokenStr,
		Expires: validUntil,
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	respSpy := httptest.NewRecorder()
	appliance.CheckAuthorization(handler).ServeHTTP(respSpy, req)

	require.Equal(t, http.StatusAccepted, respSpy.Code)
}

func TestCheckAuthorization_RedirectsToErrorPageWhenNoCookieSupplied(t *testing.T) {
	req, err := http.NewRequest("GET", "example.com", nil)
	require.NoError(t, err)
	assertDirectAndHandlerNotCalled(t, req)
}

func TestCheckAuthorization_RedirectsToErrorPageWhenCookieContainsInvalidJWT(t *testing.T) {
	req, err := http.NewRequest("GET", "example.com", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:    authCookieName,
		Value:   "not-a-jwt",
		Expires: time.Now().Add(time.Hour),
	})
	assertDirectAndHandlerNotCalled(t, req)
}

func TestCheckAuthorization_RedirectsToErrorPageWhenCookieContainsJWTWithIncorrectSignature(t *testing.T) {
	validUntil := time.Now().Add(time.Hour).UTC()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		jwtClaimsValidUntilKey: validUntil.Format(time.RFC3339),
	})
	tokenStr, err := token.SignedString([]byte("wrong-key!"))

	require.NoError(t, err)
	req, err := http.NewRequest("GET", "example.com", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:    authCookieName,
		Value:   tokenStr,
		Expires: validUntil,
	})
	assertDirectAndHandlerNotCalled(t, req)
}

func TestCheckAuthorization_RedirectsToErrorPageWhenCookieContainsJWTWithMalformedClaims(t *testing.T) {
	validUntil := time.Now().Add(time.Hour).UTC()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"wrong-key": validUntil.Format(time.RFC3339),
	})
	tokenStr, err := token.SignedString(appliance.jwtSecret)

	require.NoError(t, err)
	req, err := http.NewRequest("GET", "example.com", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:    authCookieName,
		Value:   tokenStr,
		Expires: validUntil,
	})
	assertDirectAndHandlerNotCalled(t, req)
}

func TestCheckAuthorization_RedirectsToErrorPageWhenCookieContainsJWTWithExpiredValidity(t *testing.T) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		jwtClaimsValidUntilKey: time.Now().Add(-1 * time.Hour).Format(time.RFC3339),
	})
	tokenStr, err := token.SignedString(appliance.jwtSecret)

	require.NoError(t, err)
	req, err := http.NewRequest("GET", "example.com", nil)
	require.NoError(t, err)
	req.AddCookie(&http.Cookie{
		Name:    authCookieName,
		Value:   tokenStr,
		Expires: time.Now().Add(time.Hour),
	})
	assertDirectAndHandlerNotCalled(t, req)
}

func assertDirectAndHandlerNotCalled(t *testing.T, req *http.Request) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		require.Fail(t, "next handler should not be called")
	})
	respSpy := httptest.NewRecorder()
	appliance.CheckAuthorization(handler).ServeHTTP(respSpy, req)

	require.Equal(t, http.StatusFound, respSpy.Code)
}
