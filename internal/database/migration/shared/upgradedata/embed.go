package shared

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
)

//go:generate go run ./cmd/generator
// Ensure stitched-migration-graph.json is generated

//go:embed stitched-migration-graph.json
var upgradeDataPayloadContents string

// stitchedMigationsBySchemaName is a map from schema name to migration upgrade metadata.
// The data backing the map is updated by `go generating` this package.
var stitchedMigationsBySchemaName = map[string]shared.StitchedMigration{}

func init() {
	if err := json.Unmarshal([]byte(upgradeDataPayloadContents), &stitchedMigationsBySchemaName); err != nil {
		panic(fmt.Sprintf("failed to load upgrade data (check the contents of internal/database/migration/shared/upgradedata/stitched-migration-graph.json): %s", err))
	}
}
