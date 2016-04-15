package oauth2util

import (
	"strings"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"sourcegraph.com/sourcegraph/sourcegraph/auth"
	"sourcegraph.com/sourcegraph/sourcegraph/auth/accesstoken"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
)

// GRPCMiddleware reads the OAuth2 access token from the gRPC call's
// metadata. If present and valid, its information is added to the
// context.
//
// Lack of authentication is not an error, but a failed authentication
// attempt does result in a non-nil error.
func GRPCMiddleware(ctx context.Context) (context.Context, error) {
	md, ok := metadata.FromContext(ctx)
	if !ok {
		return ctx, nil
	}

	authMD, ok := md["authorization"]
	if !ok || len(authMD) == 0 {
		return ctx, nil
	}

	// This is for backwards compatibility with client instances that are running older versions
	// of sourcegraph (< v0.7.22).
	// TODO: remove this hack once clients upgrade to binaries having the new grpc-go API.
	authToken := authMD[len(authMD)-1]

	parts := strings.SplitN(authToken, " ", 2)
	if len(parts) != 2 {
		return nil, grpc.Errorf(codes.InvalidArgument, "invalid authorization metadata")
	}
	if !strings.EqualFold(parts[0], "bearer") {
		return ctx, nil
	}

	tokStr := parts[1]

	// Elevate authorization level (using elevatedActor) to allow
	// looking up registered clients' public keys.
	actor, _, err := accesstoken.ParseAndVerify(elevatedActor(ctx), tokStr)
	if err != nil {
		return nil, grpc.Errorf(codes.Unauthenticated, "access token middleware failed to parse/verify token: %s", err)
	}

	// Make future calls use this access token.
	ctx = sourcegraph.WithCredentials(ctx, oauth2.StaticTokenSource(&oauth2.Token{TokenType: "Bearer", AccessToken: tokStr}))

	// Set actor in context.
	if actor != nil {
		ctx = auth.WithActor(ctx, *actor)
	}

	return ctx, nil
}

func elevatedActor(ctx context.Context) context.Context {
	return auth.WithActor(ctx, auth.Actor{Scope: map[string]bool{"internal:tmp": true}})
}
