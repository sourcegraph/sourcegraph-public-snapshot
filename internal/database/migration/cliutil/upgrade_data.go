package cliutil

import (
	_ "embed"
	"encoding/json"

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
		panic(errors.Wrap(err.Error(), "parsing bundled upgrade data payload, check go generate produced valid json"))
	}
}
