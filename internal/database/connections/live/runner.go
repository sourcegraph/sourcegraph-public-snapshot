package connections

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func RunnerFromDSNs(dsns map[string]string, appName string, newStore StoreFactory) *runner.Runner {
	makeFactory := func(
		name string,
		schema *schemas.Schema,
		factory func(dsn, appName string, observationContext *observation.Context) (*sql.DB, error),
	) runner.StoreFactory {
		return func(ctx context.Context) (runner.Store, error) {
			db, err := factory(dsns[name], appName, &observation.TestContext)
			if err != nil {
				return nil, err
			}

			return initStore(ctx, newStore, db, schema)
		}
	}

	storeFactoryMap := map[string]runner.StoreFactory{
		"frontend":     makeFactory("frontend", schemas.Frontend, RawNewFrontendDB),
		"codeintel":    makeFactory("codeintel", schemas.CodeIntel, RawNewCodeIntelDB),
		"codeinsights": makeFactory("codeinsights", schemas.CodeInsights, RawNewCodeInsightsDB),
	}

	return runner.NewRunner(storeFactoryMap)
}

func runnerFromDB(newStore StoreFactory, db *sql.DB, schemas ...*schemas.Schema) *runner.Runner {
	storeFactoryMap := make(map[string]runner.StoreFactory, len(schemas))
	for _, schema := range schemas {
		schema := schema

		storeFactoryMap[schema.Name] = func(ctx context.Context) (runner.Store, error) {
			return initStore(ctx, newStore, db, schema)
		}
	}

	return runner.NewRunner(storeFactoryMap)
}
