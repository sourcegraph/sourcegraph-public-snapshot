package auth

import (
	"net/http"
	"slices"
	"strings"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/cody-gateway/internal/response"
	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/internal/sams"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SAMSAuthenticator provides an auth middleware that uses SAMS service-to-service tokens to authenticate the requests.
type SAMSAuthenticator struct {
	Logger     log.Logger
	SAMSClient sams.Client
}

func (a *SAMSAuthenticator) Middleware(requiredScopes []sams.Scope, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger := trace.Logger(r.Context(), a.Logger)
		token, err := authbearer.ExtractBearer(r.Header)
		if err != nil {
			response.JSONError(logger, w, http.StatusBadRequest, err)
			return
		}

		introspectionResponse, err := a.SAMSClient.IntrospectToken(r.Context(), token)
		if err != nil {
			logger.Error("error introspecting token", log.Error(err))
			response.JSONError(logger, w, http.StatusInternalServerError, err)
			return
		}

		if !introspectionResponse.Active {
			logger.Error(
				"attempt to authenticate with inactive SAMS token",
				log.String("client", introspectionResponse.ClientID))
			response.JSONError(logger, w, http.StatusUnauthorized, errors.New("inactive token"))
			return
		}

		gotScopes := strings.Split(introspectionResponse.Scope, " ")
		for _, requiredScope := range requiredScopes {
			if !slices.Contains(gotScopes, requiredScope) {
				logger.Error(
					"attempt to authenticate using SAMS token without required scope",
					log.Strings("gotScopes", gotScopes),
					log.String("requiredScope", requiredScope))
				response.JSONError(logger, w, http.StatusForbidden, errors.New("missing required scope"))
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
