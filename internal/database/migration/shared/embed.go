package shared

import (
	_ "embed"
	"encoding/json"
	"fmt"
)

//go:generate go run ./upgradedata/cmd/generator
// Ensure upgradedata/stitched-migration-graph.json is generated

//go:embed upgradedata/stitched-migration-graph.json
var upgradeDataPayloadContents string

// StitchedMigationsBySchemaName is a map from schema name to migration upgrade metadata.
// The data backing the map is updated by `go generating` this package.
var StitchedMigationsBySchemaName = map[string]StitchedMigration{}

func init() {
	if err := json.Unmarshal([]byte(upgradeDataPayloadContents), &StitchedMigationsBySchemaName); err != nil {
		panic(fmt.Sprintf("failed to load upgrade data (check the contents of internal/database/migration/shared/upgradedata/stitched-migration-graph.json): %s", err))
	}
}
