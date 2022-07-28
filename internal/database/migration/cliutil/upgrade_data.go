package cliutil

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/stitch"
)

//go:generate go run ./upgrade_data
// Ensure upgrade_data/payload.json is generated

//go:embed upgrade_data/payload.json
var upgradeDataPayloadContents string

// stitchedMigationsBySchemaName is a map from schema name to migration upgrade metadata.
// The data backing the map is updated by `go generating` this package.
var stitchedMigationsBySchemaName = map[string]stitch.StitchedMigration{}

func init() {
	if err := json.Unmarshal([]byte(upgradeDataPayloadContents), &stitchedMigationsBySchemaName); err != nil {
		panic(fmt.Sprintf("failed to load upgrade data (check the contents of internal/database/migration/cliutil/upgrade_data/payload.json): %s", err))
	}
}

// filterStitchedMigrationsForTags returns a copy of the pre-compiled stitchedMap with references
// to tags outside of the given set removed. This allows a migrator instance that knows the upgrade
// path from X -> Y to also know the path from any partial upgrade X <= W -> Z <= Y.
func filterStitchedMigrationsForTags(tags []string) (map[string]stitch.StitchedMigration, error) {
	stitchedMigrationBySchemaName := make(map[string]stitch.StitchedMigration, len(schemas.SchemaNames))
	for _, schemaName := range schemas.SchemaNames {
		leafIDsByRev := make(map[string][]int, len(tags))
		for _, tag := range tags {
			leafIDsByRev[tag] = stitchedMigationsBySchemaName[schemaName].LeafIDsByRev[tag]
		}

		stitchedMigrationBySchemaName[schemaName] = stitch.StitchedMigration{
			Definitions:  stitchedMigationsBySchemaName[schemaName].Definitions,
			LeafIDsByRev: leafIDsByRev,
		}
	}

	return stitchedMigrationBySchemaName, nil
}
