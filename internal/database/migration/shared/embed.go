package shared

import (
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	root       = "internal/database/migration/shared/data"
	stitchfile = filepath.Join(root, "stitched-migration-graph.json")
	constfile  = filepath.Join(root, "cmd/generator/consts.go")
)

//go:embed data/stitched-migration-graph.json
var stitchedPayloadContents string

// StitchedMigationsBySchemaName is a map from schema name to migration upgrade metadata.
// The data backing the map is updated by `go generating` this package.
var StitchedMigationsBySchemaName = map[string]StitchedMigration{}

func init() {
	if err := json.Unmarshal([]byte(stitchedPayloadContents), &StitchedMigationsBySchemaName); err != nil {
		panic(fmt.Sprintf("failed to load upgrade data (check the contents of %s): %s", stitchfile, err))
	}
}

//go:embed data/frozen/*
var frozenDataDir embed.FS

// GetFrozenDefinitions returns the schema definitions frozen at a given revision. This
// function returns an error if the given schema has not been generated into data/frozen.
func GetFrozenDefinitions(schemaName, rev string) (*definition.Definitions, error) {
	f, err := frozenDataDir.Open(fmt.Sprintf("data/frozen/%s.json", rev))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.Newf("failed to load schema at revision %q (check the versions listed in %s)", rev, constfile)
		}

		return nil, err
	}
	defer f.Close()

	content, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var definitionBySchema map[string]struct {
		Definitions *definition.Definitions
	}
	if err := json.Unmarshal(content, &definitionBySchema); err != nil {
		return nil, err
	}

	return definitionBySchema[schemaName].Definitions, nil
}
