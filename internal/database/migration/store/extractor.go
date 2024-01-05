package store

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type BasestoreExtractor struct {
	Runner *runner.Runner
}

func (r BasestoreExtractor) Store(ctx context.Context, schemaName string) (*basestore.Store, error) {
	shareableStore, err := ExtractDB(ctx, r.Runner, schemaName)
	if err != nil {
		return nil, err
	}

	return basestore.NewWithHandle(basestore.NewHandleWithDB(log.NoOp(), shareableStore, sql.TxOptions{})), nil
}

func ExtractDatabase(ctx context.Context, r *runner.Runner) (database.DB, error) {
	db, err := ExtractDB(ctx, r, "frontend")
	if err != nil {
		return nil, err
	}

	return database.NewDB(log.Scoped("migrator"), db), nil
}

func ExtractDB(ctx context.Context, r *runner.Runner, schemaName string) (*sql.DB, error) {
	store, err := r.Store(ctx, schemaName)
	if err != nil {
		return nil, err
	}

	// NOTE: The migration runner package cannot import basestore without
	// creating a cyclic import in db connection packages. Hence, we cannot
	// embed basestore.ShareableStore here and must "backdoor" extract the
	// database connection.
	shareableStore, ok := basestore.Raw(store)
	if !ok {
		return nil, errors.New("store does not support direct database handle access")
	}

	return shareableStore, nil
}
