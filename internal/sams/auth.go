package sams

import (
	"net/http"
	"slices"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

// Authenticator provides an auth middleware that uses SAMS service-to-service tokens to authenticate the requests.
type Authenticator struct {
	Logger     log.Logger
	SAMSClient Client
}

// RequireScopes performs an authorization check on the incoming HTTP request.
// It will return a 401 if the request does not have a valid SAMS access token,
// or a 403 if the token is valid but is missing ANY of the required scopes.
func (a *Authenticator) RequireScopes(requiredScopes []Scope, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := trace.Logger(r.Context(), a.Logger)
		token, err := authbearer.ExtractBearer(r.Header)
		if err != nil || token == "" {
			logger.Error("error extracting bearer token", log.Error(err))
			const unauthorized = http.StatusUnauthorized
			http.Error(w, http.StatusText(unauthorized), unauthorized)
			return
		}

		introspectionResponse, err := a.SAMSClient.IntrospectToken(r.Context(), token)
		if err != nil || introspectionResponse == nil {
			logger.Error("error introspecting token", log.Error(err))
			const ise = http.StatusInternalServerError
			http.Error(w, http.StatusText(ise), ise)
			return
		}

		if !introspectionResponse.Active {
			logger.Error(
				"attempt to authenticate with inactive SAMS token",
				log.String("client", introspectionResponse.ClientID))
			const unauthorized = http.StatusUnauthorized
			http.Error(w, "Unauthorized: Inactive token", unauthorized)
			return
		}

		gotScopes := strings.Split(introspectionResponse.Scope, " ")
		for _, requiredScope := range requiredScopes {
			if !slices.Contains(gotScopes, string(requiredScope)) {
				logger.Error(
					"attempt to authenticate using SAMS token without required scope",
					log.Strings("gotScopes", gotScopes),
					log.String("requiredScope", string(requiredScope)))
				http.Error(w, "Forbidden: Missing required scope", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
