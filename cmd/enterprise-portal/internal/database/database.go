package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"

	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/codyaccess"
	"github.com/sourcegraph/sourcegraph/cmd/enterprise-portal/internal/database/subscriptions"
)

var databaseTracer = otel.Tracer("enterprise-portal/internal/database")

// DB is the database handle for the storage layer.
type DB struct {
	DB *pgxpool.Pool
}

func (db *DB) Subscriptions() *subscriptions.Store {
	return subscriptions.NewStore(db.DB)
}

func (db *DB) CodyAccess() *codyaccess.Store {
	return codyaccess.NewStore(db.DB)
}

func databaseName(msp bool) string {
	if msp {
		return "enterprise_portal"
	}

	// Use whatever the current database is for local development.
	return os.Getenv("PGDATABASE")
}

// NewHandle returns a new database handle with the given configuration. It may
// attempt to auto-migrate the database schema if the application version has
// changed.
func NewHandle(ctx context.Context, logger log.Logger, contract runtime.Contract, redisClient *redis.Client, currentVersion string) (*DB, error) {
	err := maybeMigrate(ctx, logger, contract, redisClient, currentVersion)
	if err != nil {
		return nil, errors.Wrap(err, "maybe migrate")
	}

	pool, err := contract.PostgreSQL.GetConnectionPool(ctx, databaseName(contract.MSP))
	if err != nil {
		return nil, errors.Wrap(err, "get connection pool")
	}
	return &DB{DB: pool}, nil
}

// Close closes all connections in the pool and rejects future Acquire calls.
// Blocks until all connections are returned to pool and closed.
func (db *DB) Close() {
	db.DB.Close()
}
