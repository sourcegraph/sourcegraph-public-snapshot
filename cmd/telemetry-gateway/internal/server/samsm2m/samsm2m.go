package samsm2m

import (
	"context"
	"slices"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/authbearer"
	"github.com/sourcegraph/sourcegraph/internal/sams"
)

const requiredSamsScope = "telemetry_gateway::events::write"

// CheckWriteEventsScope ensures the request context has a valid SAMS token with requiredSamsScope.
// It returns a gRPC status error suitable to be returned directly from an RPC implementation.
//
// See: go/sams-m2m
func CheckWriteEventsScope(ctx context.Context, logger log.Logger, samsClient sams.Client) error {
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
	result, err := samsClient.IntrospectToken(ctx, token)
	if err != nil {
		logger.Error("samsClient.IntrospectToken failed", log.Error(err))
		return status.Error(codes.PermissionDenied, "unable to validate token")
	}
	if !result.Active {
		return status.Error(codes.PermissionDenied, "token is inactive")
	}

	gotScopes := strings.Split(result.Scope, " ")
	if !slices.Contains(gotScopes, requiredSamsScope) {
		logger.Error(
			"attempt to authenticate using SAMS token without required scope",
			log.Strings("gotScopes", gotScopes),
			log.String("requiredScope", requiredSamsScope))
		return status.Error(codes.PermissionDenied, "token does not have the required scopes")
	}

	return nil
}
