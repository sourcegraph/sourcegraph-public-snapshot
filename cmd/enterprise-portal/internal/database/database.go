package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

// DB is the database handle for the storage layer.
type DB struct {
	db *pgxpool.Pool
}

// ⚠️ WARNING: This list is meant to be read-only.
var allTables = []any{
	&Subscription{},
	&Permission{},
}

const databaseName = "enterprise-portal"

// NewHandle returns a new database handle with the given configuration. It may
// attempt to auto-migrate the database schema if the application version has
// changed.
func NewHandle(ctx context.Context, logger log.Logger, contract runtime.Contract, redisClient *redis.Client, currentVersion string) (*DB, error) {
	err := maybeMigrate(ctx, logger, contract, redisClient, currentVersion)
	if err != nil {
		return nil, errors.Wrap(err, "maybe migrate")
	}

	pool, err := contract.PostgreSQL.GetConnectionPool(ctx, databaseName)
	if err != nil {
		return nil, errors.Wrap(err, "get connection pool")
	}
	return &DB{db: pool}, nil
}
