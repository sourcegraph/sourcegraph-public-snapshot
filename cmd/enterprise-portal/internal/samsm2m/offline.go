package samsm2m

import (
	"context"
	"time"

	"github.com/sourcegraph/log"

	sams "github.com/sourcegraph/sourcegraph-accounts-sdk-go"
	"github.com/sourcegraph/sourcegraph-accounts-sdk-go/scopes"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

// OfflineTokenInstrospector is a TokenIntrospector that always returns a valid
// introspection response with the configured client ID and scopes.
//
// 🚨 Only use for local development!
type OfflineTokenInstrospector struct {
	Logger   log.Logger
	ClientID string
	Scopes   scopes.Scopes
}

func (i OfflineTokenInstrospector) IntrospectToken(context.Context, string) (*sams.IntrospectTokenResponse, error) {
	if !env.InsecureDev {
		i.Logger.Fatal("only use OfflineTokenInstrospector in development with 'INSECURE_DEV=true'")
	}
	i.Logger.Debug("IntrospectToken request",
		log.Strings("scopes", scopes.ToStrings(i.Scopes)))
	return &sams.IntrospectTokenResponse{
		Active:    true,
		ClientID:  i.ClientID,
		Scopes:    i.Scopes,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}, nil
}
