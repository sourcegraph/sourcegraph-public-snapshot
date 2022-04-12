package connections

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RunnerFromDSNs(dsns map[string]string, appName string, newStore StoreFactory) (*runner.Runner, error) {
	return RunnerFromDSNsWithSchemas(dsns, appName, newStore, schemas.Schemas)
}

func RunnerFromDSNsWithSchemas(dsns map[string]string, appName string, newStore StoreFactory, availableSchemas []*schemas.Schema) (*runner.Runner, error) {
	frontendSchema, ok := schemaByName(availableSchemas, "frontend")
	if !ok {
		return nil, errors.Newf("no available schema matches %q", "frontend")
	}
	codeintelSchema, ok := schemaByName(availableSchemas, "codeintel")
	if !ok {
		return nil, errors.Newf("no available schema matches %q", "codeintel")
	}
	codeinsightsSchema, ok := schemaByName(availableSchemas, "codeinsights")
	if !ok {
		return nil, errors.Newf("no available schema matches %q", "codeinsights")
	}

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
		"frontend":     makeFactory("frontend", frontendSchema, RawNewFrontendDB),
		"codeintel":    makeFactory("codeintel", codeintelSchema, RawNewCodeIntelDB),
		"codeinsights": makeFactory("codeinsights", codeinsightsSchema, RawNewCodeInsightsDB),
	}

	return runner.NewRunnerWithSchemas(storeFactoryMap, availableSchemas), nil
}

func schemaByName(schemas []*schemas.Schema, name string) (*schemas.Schema, bool) {
	for _, schema := range schemas {
		if schema.Name == name {
			return schema, true
		}
	}

	return nil, false
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
