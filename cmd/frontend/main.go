package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/auth/providers/httpauth"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

// Note: All frontend code should be added to shared.Main, not here. See that
// function for details.

func main() {
	// Set dummy authz provider to unblock channel for checking permissions in GraphQL APIs.
	// See https://github.com/sourcegraph/sourcegraph/issues/3847 for details.
	authz.SetProviders(true, []authz.Provider{})

	env.Lock()
	env.HandleHelpFlag()

	shared.Main(func(db database.DB, _ conftypes.UnifiedWatchable) enterprise.Services {
		if envvar.OAuth2ProxyMode() {
			httpauth.Init(db)
		}
		return enterprise.DefaultServices()
	})
}
