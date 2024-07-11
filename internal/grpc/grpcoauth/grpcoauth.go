package grpcoauth

import (
	"context"

	"golang.org/x/oauth2"
	"google.golang.org/grpc/credentials"

	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// TokenSource supplies PerRPCCredentials from an oauth2.TokenSource.
// It is a fork of the implementation in "google.golang.org/grpc/credentials/oauth",
// but checks "internal/env" to toggle the "must require secure transport"
// behaviour. Without these changes, the upstream token source cannot be used
// in local development.
type TokenSource struct {
	oauth2.TokenSource
}

// GetRequestMetadata gets the request metadata as a map from a TokenSource.
func (ts TokenSource) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	token, err := ts.Token()
	if err != nil {
		return nil, err
	}
	if !env.InsecureDev {
		ri, _ := credentials.RequestInfoFromContext(ctx)
		if err = credentials.CheckSecurityLevel(ri.AuthInfo, credentials.PrivacyAndIntegrity); err != nil {
			return nil, errors.Newf("unable to transfer TokenSource PerRPCCredentials: %w", err)
		}
	}
	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}

// RequireTransportSecurity indicates whether the credentials requires transport security.
// For this implementation, we disable it if the INSECURE_DEV environment variable is set.
func (ts TokenSource) RequireTransportSecurity() bool {
	return !env.InsecureDev
}
