package samsflags

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/secrets"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
)

// Memoize so that we don't generate a lot of noise when printing output for
// creating multiple clients.
var localDevClientIDOnce sync.Once
var localDevClientID string
var localDevClientIDError error

func defaultLocalDevClientID(ctx context.Context) (string, error) {
	localDevClientIDOnce.Do(func() {
		ss, err := secrets.FromContext(ctx)
		if err != nil {
			localDevClientIDError = err
			return
		}
		std.Out.WriteSuggestionf("Using default local dev client ID for %s", SAMSDevURL)
		localDevClientID, err = ss.GetExternal(ctx, secrets.ExternalSecret{
			Project: "sourcegraph-local-dev",
			Name:    "SG_LOCAL_DEV_SAMS_CLIENT_ID",
		})
		if err != nil {
			localDevClientIDError = err
		}
	})
	return localDevClientID, localDevClientIDError
}

var localDevClientSecretOnce sync.Once
var localDevClientSecret string
var localDevClientSecretError error

func defaultLocalDevClientSecret(ctx context.Context) (string, error) {
	localDevClientSecretOnce.Do(func() {
		ss, err := secrets.FromContext(ctx)
		if err != nil {
			localDevClientSecretError = err
			return
		}
		std.Out.WriteSuggestionf("Using default local dev client secret for %s", SAMSDevURL)
		localDevClientSecret, err = ss.GetExternal(ctx, secrets.ExternalSecret{
			Project: "sourcegraph-local-dev",
			Name:    "SG_LOCAL_DEV_SAMS_CLIENT_SECRET",
		})
		if err != nil {
			localDevClientSecretError = err
		}
	})
	return localDevClientSecret, localDevClientSecretError
}
