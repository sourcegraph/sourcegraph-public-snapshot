package authz

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/enterprise"
	eauthz "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/authz"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/authz/resolvers"
	eiauthz "github.com/sourcegraph/sourcegraph/enterprise/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
)

func Init(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner, enterpriseServices *enterprise.Services) error {
	eauthz.Init(db, timeutil.Now)

	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			allowAccessByDefault, authzProviders, _, _ :=
				eiauthz.ProvidersFromConfig(ctx, conf.Get(), database.GlobalExternalServices)
			authz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()

	enterpriseServices.AuthzResolver = resolvers.NewResolver(db, timeutil.Now)

	return nil
}
