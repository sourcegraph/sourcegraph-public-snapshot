package samsm2m

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	otelcodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var requiredSamsScope = scopes.ToScope(scopes.ServiceTelemetryGateway, "events", scopes.ActionWrite)

var tracer = otel.GetTracerProvider().Tracer("telemetry-gateway/samsm2m")

type TokenIntrospector interface {
	IntrospectToken(ctx context.Context, token string) (*sams.IntrospectTokenResponse, error)
}

// CheckWriteEventsScope ensures the request context has a valid SAMS MSM token
// with requiredSamsScope. It returns a gRPC status error suitable to be returned
// directly from an RPC implementation.
//
// See: go/sams-m2m
func CheckWriteEventsScope(ctx context.Context, logger log.Logger, tokens TokenIntrospector) (err error) {
	var span trace.Span
	ctx, span = tracer.Start(ctx, "CheckWriteEventsScope")
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(otelcodes.Error, "check failed")
		}
		span.End()
	}()

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "no token header")
	}

	var token string
	if v := md.Get("authorization"); len(v) == 1 && v[0] != "" {
		var err error
		token, err = authbearer.ExtractBearerContents(v[0])
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "invalid token header: %v", err)
		}
	} else {
		return status.Error(codes.Unauthenticated, "no token header value")
	}

	// TODO: as part of go/sams-m2m we need to build out a SDK for SAMS M2M
	// consumers that has a recommended short-caching mechanism. Avoid doing it
	// for now until we have a concerted effort.
	result, err := tokens.IntrospectToken(ctx, token)
	if err != nil {
		if errors.IsContextCanceled(err) {
			return status.Error(codes.Canceled, "request canceled")
		} else {
			logger.Error("samsClient.IntrospectToken failed", log.Error(err))
			return status.Error(codes.Internal, "unable to validate token")
		}
	}
	span.SetAttributes(attribute.String("client_id", result.ClientID))

	// Active encapsulates whether the token is active, including expiration.
	if !result.Active {
		// Record detailed error in span, and return an opaque one
		span.RecordError(errors.New("inactive token"))
		return status.Error(codes.PermissionDenied, "permission denied")
	}

	// Check for our required scope.
	if !result.Scopes.Match(requiredSamsScope) {
		// Record detailed error in span and logs, and return an opaque one
		err = errors.Newf("got scopes %+v, required: %+v", result.Scopes, requiredSamsScope)
		span.RecordError(err)
		logger.Error("attempt to authenticate using SAMS token without required scope",
			log.String("clientID", result.ClientID),
			log.Error(err))
		return status.Error(codes.PermissionDenied, "permission denied")
	}

	return nil
}
