package bg

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/pkg/conf"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// MigrateExternalAccounts finishes the user_multiple_external_accounts migration, which adds a
// new table user_external_accounts so users can have multiple external accounts. The previous users
// table omitted the current provider type (e.g., "saml" or "openidconnect"), so the SQL migration
// script just sets the new user_external_accounts.service_type column to 'migration_in_progress'
// (for this func to fill in).
func MigrateExternalAccounts(ctx context.Context) {
	// Assume that the first currently configured auth provider is the one used for all
	// 'migration_in_progress' user_external_accounts rows. Users for whom this assumption is
	// invalid wouldn't have been able to sign in anyway.
	var serviceType string
	if ps := conf.AuthProviders(); len(ps) >= 1 {
		serviceType = conf.AuthProviderType(ps[0])
	}
	if serviceType == "builtin" {
		// "builtin" means the external accounts were linked when another auth provider was used,
		// but we can't determine which auth provider it was anymore, so delete the external
		// accounts.
		serviceType = ""
	}

	if err := db.ExternalAccounts.TmpMigrate(ctx, serviceType); err != nil {
		log15.Error("Error migrating user external accounts.", "err", err)
		return
	}
}
