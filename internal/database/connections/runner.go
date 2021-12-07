package connections

import (
	"context"
	"database/sql"

	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func NewDefaultRunner(dsns map[string]string, appName string, observationContext *observation.Context) *runner.Runner {
	operations := store.NewOperations(observationContext)
	makeFactory := func(
		name string,
		schema *schemas.Schema,
		factory func(dsn, appName string, migrate bool) (*sql.DB, error),
	) runner.StoreFactory {
		return func(ctx context.Context) (runner.Store, error) {
			db, err := factory(dsns[name], appName, false)
			if err != nil {
				return nil, err
			}

			store := store.NewWithDB(db, schema.MigrationsTableName, operations)

			if err := store.EnsureSchemaTable(ctx); err != nil {
				if closeErr := db.Close(); closeErr != nil {
					err = multierror.Append(err, closeErr)
				}

				return nil, err
			}

			return store, nil
		}
	}

	storeFactoryMap := map[string]runner.StoreFactory{
		"frontend":     makeFactory("frontend", schemas.Frontend, NewFrontendDB),
		"codeintel":    makeFactory("codeintel", schemas.CodeIntel, NewCodeIntelDB),
		"codeinsights": makeFactory("codeinsights", schemas.CodeInsights, NewCodeInsightsDB),
	}

	return runner.NewRunner(storeFactoryMap)
}
