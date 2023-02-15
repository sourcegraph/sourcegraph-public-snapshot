package dbconn

import (
	"time"

	"github.com/jackc/pgx/v4"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database/dbconn/rds"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var (
	pgAuthProvider                       = env.Get("PG_AUTH_PROVIDER", "", "The auth provider to use for connecting to the database.")
	pgAuthProviderRefreshIntervalSeconds = env.MustGetDuration("PG_AUTH_PROVIDER_REFRESH_INTERVAL_SECONDS", 300*time.Second, "The interval at which to check if the auth provider's credentials need to be refreshed.")
)

// AuthProvider is an interface for authenticating with the database
// The methods are expected to mutate the *pgx.ConnConfig
type AuthProvider interface {
	// Apply applies the auth provider credentials to the *pgx.ConnConfig
	// It will be called once during startup and then
	// periodically to ensure that the credentials are still valid in the background
	Apply(log.Logger, *pgx.ConnConfig) error

	// IsRefresh indicates whether the auth provider's credentials need to be refreshed
	IsRefresh(log.Logger, *pgx.ConnConfig) bool
}

var authProviders = map[string]AuthProvider{
	"EC2_ROLE_CREDENTIALS": rds.NewAuth(),
}
