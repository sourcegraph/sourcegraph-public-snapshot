package cliutil

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// runUpgrade initializes a schema and out-of-band migration runner and performs the given upgrade plan.
func runUpgrade(ctx context.Context, runnerFactory RunnerFactoryWithSchemas, plan upgradePlan) error {
	var runnerSchemas []*schemas.Schema
	for _, schemaName := range schemas.SchemaNames {
		runnerSchemas = append(runnerSchemas, &schemas.Schema{
			Name:                schemaName,
			MigrationsTableName: schemas.MigrationsTableName(schemaName),
			Definitions:         plan.stitchedDefinitionsBySchemaName[schemaName],
		})
	}

	r, err := runnerFactory(ctx, schemas.SchemaNames, runnerSchemas)
	if err != nil {
		return err
	}

	for _, step := range plan.steps {
		operations := make([]runner.MigrationOperation, 0, len(step.schemaMigrationLeafIDsBySchemaName))
		for schemaName, leafMigrationIDs := range step.schemaMigrationLeafIDsBySchemaName {
			operations = append(operations, runner.MigrationOperation{
				SchemaName:     schemaName,
				Type:           runner.MigrationOperationTypeTargetedUp,
				TargetVersions: leafMigrationIDs,
			})
		}

		if len(step.outOfBandMigrationIDs) > 0 {
			// TODO - implement in https://github.com/sourcegraph/sourcegraph/issues/39578
			return errors.Newf("unimplemented - out of band migrations %v were deprecated in this upload", step.outOfBandMigrationIDs)
		}

		if err := r.Run(ctx, runner.Options{
			Operations:           operations,
			PrivilegedMode:       runner.ApplyPrivilegedMigrations,
			PrivilegedHash:       "",
			IgnoreSingleDirtyLog: false,
		}); err != nil {
			return err
		}
	}

	return nil
}
