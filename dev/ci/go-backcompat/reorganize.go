package main

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/keegancsmith/sqlf"
	"gopkg.in/yaml.v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func main() {
	if len(os.Args) != 2 {
		panic("expected temp directory as argument")
	}
	tempDirectory := os.Args[1]

	contents := map[string]string{}
	for _, schema := range schemas.Schemas {
		for _, definition := range schema.Definitions.All() {
			metadata, err := renderMetadata(definition)
			if err != nil {
				panic(err.Error())
			}

			migrationDirectory := filepath.Join(tempDirectory, schema.Name, strconv.Itoa(definition.ID))

			contents[filepath.Join(migrationDirectory, "metadata.yaml")] = string(metadata)
			contents[filepath.Join(migrationDirectory, "up.sql")] = definition.UpQuery.Query(sqlf.PostgresBindVar)
			contents[filepath.Join(migrationDirectory, "down.sql")] = definition.DownQuery.Query(sqlf.PostgresBindVar)
		}
	}

	for path, contents := range contents {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			panic(err.Error())
		}

		if err := os.WriteFile(path, []byte(contents), os.FileMode(0644)); err != nil {
			panic(err.Error())
		}
	}
}

func renderMetadata(definition definition.Definition) ([]byte, error) {
	return yaml.Marshal(struct {
		Name                    string `yaml:"name"`
		Parents                 []int  `yaml:"parents"`
		CreateIndexConcurrently bool   `yaml:"createIndexConcurrently"`
		Privileged              bool   `yaml:"privileged"`
		NonIdempotent           bool   `yaml:"nonIdempotent"`
	}{
		Name:                    definition.Name,
		Parents:                 definition.Parents,
		CreateIndexConcurrently: definition.IsCreateIndexConcurrently,
		Privileged:              definition.Privileged,
		NonIdempotent:           definition.NonIdempotent,
	})
}
