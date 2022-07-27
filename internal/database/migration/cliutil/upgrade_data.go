package cliutil

import (
	_ "embed"
	"encoding/json"
	"fmt"

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
