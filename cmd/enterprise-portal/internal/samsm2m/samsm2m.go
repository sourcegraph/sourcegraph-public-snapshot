package samsm2m

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/connectutil"
	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// EnterprisePortalScope returns the Enterprise Portal service scope for the
// given permission and action.
func EnterprisePortalScope(permission scopes.Permission, action scopes.Action) scopes.Scope {
	return scopes.ToScope("enterprise_portal", // TODO use scopes.ServiceEnterprisePortal
		permission, action)
}

var tracer = otel.GetTracerProvider().Tracer("telemetry-gateway/samsm2m")

type TokenIntrospector interface {
	IntrospectToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error)
}

type Request interface {
	Header() http.Header
}

// RequireScope ensures the request context has a valid SAMS M2M token
// with requiredScope. It returns a ConnectRPC status error suitable to be
// returned directly from a ConnectRPC implementation.
//
// See: go/sams-m2m
func RequireScope(ctx context.Context, logger log.Logger, tokens TokenIntrospector, requiredScope scopes.Scope, req Request) (err error) {
	logger = logger.Scoped("samsm2m")

	var span trace.Span
	ctx, span = tracer.Start(ctx, "RequireScope")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, "check failed")
		}
		span.End()
	}()

	var token string
	if v := req.Header().Values("authorization"); len(v) == 1 && v[0] != "" {
		var err error
		token, err = authbearer.ExtractBearerContents(v[0])
		if err != nil {
			return connect.NewError(connect.CodeUnauthenticated,
				errors.Wrap(err, "invalid authorization header"))
		}
	} else {
		return connect.NewError(connect.CodeUnauthenticated,
			errors.New("no authorization header"))
	}

	// TODO: as part of go/sams-m2m we need to build out a SDK for SAMS M2M
	// consumers that has a recommended short-caching mechanism. Avoid doing it
	// for now until we have a concerted effort.
	result, err := tokens.IntrospectToken(ctx, token)
	if err != nil {
		return connectutil.InternalError(ctx, logger, err, "unable to validate token")
	}
	span.SetAttributes(attribute.String("client_id", result.ClientID))

	// Active encapsulates whether the token is active, including expiration.
	if !result.Active {
		// Record detailed error in span, and return an opaque one
		span.SetAttributes(attribute.String("full_error", "inactive token"))
		return connect.NewError(connect.CodePermissionDenied, errors.New("permission denied"))
	}

	// Check for our required scope.
	if !result.Scopes.Match(requiredScope) {
		// Record detailed error in span and logs
		err = errors.Newf("got scopes %+v, required: %+v", result.Scopes, requiredScope)
		span.SetAttributes(attribute.String("full_error", err.Error()))
		logger.Error("attempt to authenticate using SAMS token without required scope",
			log.String("clientID", result.ClientID),
			log.Error(err))
		// Return an opaque error
		return connect.NewError(connect.CodePermissionDenied, errors.New("insufficient scope"))
	}

	return nil
}
