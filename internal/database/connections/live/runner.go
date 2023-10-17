package connections

import (
	"context"
	"database/sql"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func RunnerFromDSNs(out *output.Output, logger log.Logger, dsns map[string]string, appName string, newStore StoreFactory) (*runner.Runner, error) {
	return RunnerFromDSNsWithSchemas(out, logger, dsns, appName, newStore, schemas.Schemas)
}

func RunnerFromDSNsWithSchemas(out *output.Output, logger log.Logger, dsns map[string]string, appName string, newStore StoreFactory, availableSchemas []*schemas.Schema) (*runner.Runner, error) {
	var verbose = env.LogLevel == "dbug"
	frontendSchema, ok := schemaByName(availableSchemas, "frontend")
	if !ok {
		return nil, errors.Newf("no available schema matches %q", "frontend")
	}
	codeintelSchema, ok := schemaByName(availableSchemas, "codeintel")
	if !ok {
		return nil, errors.Newf("no available schema matches %q", "codeintel")
	}

	makeFactory := func(
		name string,
		schema *schemas.Schema,
		factory func(observationCtx *observation.Context, dsn, appName string) (*sql.DB, error),
	) runner.StoreFactory {
		return func(ctx context.Context) (runner.Store, error) {
			var pending output.Pending
			if verbose {
				pending = out.Pending(output.Styledf(output.StylePending, "Attempting connection to %s: %s", schema.Name, dsns[name]))
			} else {
				pending = out.Pending(output.Styledf(output.StylePending, "Attempting connection to %s", schema.Name))
			}
			db, err := factory(observation.NewContext(logger), dsns[name], appName)
			if err != nil {
				pending.Destroy()
				return nil, err
			}
			if verbose {
				pending.Complete(output.Emojif(output.EmojiSuccess, "Connection to %s: %s succeeded", schema.Name, dsns[name]))
			} else {
				pending.Complete(output.Emojif(output.EmojiSuccess, "Connection to %s succeeded", schema.Name))
			}

			return initStore(ctx, newStore, db, schema)
		}
	}
	storeFactoryMap := map[string]runner.StoreFactory{
		"frontend":  makeFactory("frontend", frontendSchema, RawNewFrontendDB),
		"codeintel": makeFactory("codeintel", codeintelSchema, RawNewCodeIntelDB),
	}

	codeinsightsSchema, ok := schemaByName(availableSchemas, "codeinsights")
	if ok {
		storeFactoryMap["codeinsights"] = makeFactory("codeinsights", codeinsightsSchema, RawNewCodeInsightsDB)
	}
	return runner.NewRunnerWithSchemas(logger, storeFactoryMap, availableSchemas), nil
}

func schemaByName(schemas []*schemas.Schema, name string) (*schemas.Schema, bool) {
	for _, schema := range schemas {
		if schema.Name == name {
			return schema, true
		}
	}

	return nil, false
}

func runnerFromDB(logger log.Logger, newStore StoreFactory, db *sql.DB, schemas ...*schemas.Schema) *runner.Runner {
	storeFactoryMap := make(map[string]runner.StoreFactory, len(schemas))
	for _, schema := range schemas {
		schema := schema

		storeFactoryMap[schema.Name] = func(ctx context.Context) (runner.Store, error) {
			return initStore(ctx, newStore, db, schema)
		}
	}

	return runner.NewRunner(logger, storeFactoryMap)
}
