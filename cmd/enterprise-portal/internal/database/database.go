package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"gorm.io/gorm/schema"

	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/managedservicesplatform/runtime"
)

var databaseTracer = otel.Tracer("enterprise-portal/internal/database")

// DB is the database handle for the storage layer.
type DB struct {
	db *pgxpool.Pool
}

func (db *DB) Subscriptions() *SubscriptionsStore {
	return newSubscriptionsStore(db.db)
}

// ⚠️ WARNING: This list is meant to be read-only.
var allTables = []schema.Tabler{
	&Subscription{},
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
	return &DB{db: pool}, nil
}

// transaction executes the given function within a transaction. If the function
// returns an error, the transaction will be rolled back.
func transaction(ctx context.Context, db *pgxpool.Pool, fn func(tx pgx.Tx) error) (err error) {
	tx, err := db.Begin(ctx)
	if err != nil {
		return errors.Wrap(err, "begin")
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		// Only return the rollback error if there is no other error.
		if err == nil {
			err = errors.Wrap(rollbackErr, "rollback")
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		return errors.Wrap(err, "commit")
	}
	return nil
}
