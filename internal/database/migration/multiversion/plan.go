package multiversion

import (
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/version/upgradestore"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type MigrationPlan struct {
	// the source and target instance versions
	from, to oobmigration.Version

	// the stitched schema migration definitions over the entire version range by schema name
	stitchedDefinitionsBySchemaName map[string]*definition.Definitions

	// the sequence of migration steps over the stitched schema migration definitions; we can't
	// simply apply all schema migrations as out-of-band migration can only run within a certain
	// slice of the schema's definition where that out-of-band migration was defined
	steps []MigrationStep
}

// SerializeUpgradePlan converts a MigrationPlan into a relevant UpgradePlan for display in
// the "hobbled" UI displayed during a multi-version upgrade.
func SerializeUpgradePlan(plan MigrationPlan) upgradestore.UpgradePlan {
	if len(plan.steps) == 0 {
		return upgradestore.UpgradePlan{}
	}

	oobMigrationIDs := []int{}
	for _, step := range plan.steps {
		oobMigrationIDs = append(oobMigrationIDs, step.outOfBandMigrationIDs...)
	}

	n := len(plan.steps)
	lastStep := plan.steps[n-1]
	leafIDsBySchemaName := lastStep.schemaMigrationLeafIDsBySchemaName

	migrations := map[string][]int{}
	migrationNames := map[string]map[int]string{}
	for schema, leafIDs := range leafIDsBySchemaName {
		migrationNames[schema] = map[int]string{}

		if definitions, err := plan.stitchedDefinitionsBySchemaName[schema].Up(nil, leafIDs); err == nil {
			for _, definition := range definitions {
				migrations[schema] = append(migrations[schema], definition.ID)
				migrationNames[schema][definition.ID] = definition.Name
			}
		}
	}

	return upgradestore.UpgradePlan{
		OutOfBandMigrationIDs: oobMigrationIDs,
		Migrations:            migrations,
		MigrationNames:        migrationNames,
	}
}

type MigrationStep struct {
	// the target version to migrate to
	instanceVersion oobmigration.Version

	// the leaf migrations of this version by schema name
	schemaMigrationLeafIDsBySchemaName map[string][]int

	// the set of out-of-band migrations that must complete before schema migrations begin
	// for the following minor instance version
	outOfBandMigrationIDs []int
}

func PlanMigration(from, to oobmigration.Version, versionRange []oobmigration.Version, interrupts []oobmigration.MigrationInterrupt) (MigrationPlan, error) {
	versionTags := make([]string, 0, len(versionRange))
	for _, version := range versionRange {
		versionTags = append(versionTags, version.GitTag())
	}

	// Retrieve relevant stitched migrations for this version range
	stitchedMigrationBySchemaName, err := filterStitchedMigrationsForTags(versionTags)
	if err != nil {
		return MigrationPlan{}, err
	}

	// Extract/rotate stitched migration definitions so we can query them by schem name
	stitchedDefinitionsBySchemaName := make(map[string]*definition.Definitions, len(stitchedMigrationBySchemaName))
	for schemaName, stitchedMigration := range stitchedMigrationBySchemaName {
		stitchedDefinitionsBySchemaName[schemaName] = stitchedMigration.Definitions
	}

	// Extract/rotate leaf identifiers so we can query them by version/git-tag first
	leafIDsBySchemaNameByTag := make(map[string]map[string][]int, len(versionRange))
	for schemaName, stitchedMigration := range stitchedMigrationBySchemaName {
		for tag, bounds := range stitchedMigration.BoundsByRev {
			if _, ok := leafIDsBySchemaNameByTag[tag]; !ok {
				leafIDsBySchemaNameByTag[tag] = map[string][]int{}
			}

			leafIDsBySchemaNameByTag[tag][schemaName] = bounds.LeafIDs
		}
	}

	//
	// Interleave out-of-band migration interrupts and schema migrations

	steps := make([]MigrationStep, 0, len(interrupts)+1)
	for _, interrupt := range interrupts {
		steps = append(steps, MigrationStep{
			instanceVersion:                    interrupt.Version,
			schemaMigrationLeafIDsBySchemaName: leafIDsBySchemaNameByTag[interrupt.Version.GitTag()],
			outOfBandMigrationIDs:              interrupt.MigrationIDs,
		})
	}
	steps = append(steps, MigrationStep{
		instanceVersion:                    to,
		schemaMigrationLeafIDsBySchemaName: leafIDsBySchemaNameByTag[to.GitTag()],
		outOfBandMigrationIDs:              nil, // all required out of band migrations have already completed
	})

	return MigrationPlan{
		from:                            from,
		to:                              to,
		stitchedDefinitionsBySchemaName: stitchedDefinitionsBySchemaName,
		steps:                           steps,
	}, nil
}

// filterStitchedMigrationsForTags returns a copy of the pre-compiled stitchedMap with references
// to tags outside of the given set removed. This allows a migrator instance that knows the migration
// path from X -> Y to also know the path from any partial migration X <= W -> Z <= Y.
func filterStitchedMigrationsForTags(tags []string) (map[string]shared.StitchedMigration, error) {
	filteredStitchedMigrationBySchemaName := make(map[string]shared.StitchedMigration, len(schemas.SchemaNames))
	for _, schemaName := range schemas.SchemaNames {
		boundsByRev := make(map[string]shared.MigrationBounds, len(tags))
		for _, tag := range tags {
			bounds, ok := shared.StitchedMigationsBySchemaName[schemaName].BoundsByRev[tag]
			if !ok {
				return nil, errors.Newf("unknown tag %q", tag)
			}

			boundsByRev[tag] = bounds
		}

		filteredStitchedMigrationBySchemaName[schemaName] = shared.StitchedMigration{
			Definitions: shared.StitchedMigationsBySchemaName[schemaName].Definitions,
			BoundsByRev: boundsByRev,
		}
	}

	return filteredStitchedMigrationBySchemaName, nil
}
