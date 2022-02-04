package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
)

func main() {
	if len(os.Args) != 2 {
		panic("expected temp directory as argument")
	}
	tempDirectory := os.Args[1]

	minMigrationVersions := map[string]int{
		"frontend":     1528395834,
		"codeintel":    1000000015,
		"codeinsights": 1000000000,
	}

	contents := map[string]string{}
	for _, schema := range schemas.Schemas {
		versionBase := minMigrationVersions[schema.Name]

		for versionOffset, definition := range schema.Definitions.All() {
			version := versionBase + versionOffset

			metadataLines := []string{}
			metadataLines = append(metadataLines, fmt.Sprintf("name: '%s'", definition.Name))
			if versionOffset > 0 {
				metadataLines = append(metadataLines, fmt.Sprintf("parent: %d", version-1))
			}
			if definition.IsCreateIndexConcurrently {
				metadataLines = append(metadataLines, "createIndexConcurrently: true")
			}
			metadata := strings.Join(metadataLines, "\n") + "\n"

			preambleLines := []string{}
			for _, line := range metadataLines {
				preambleLines = append(preambleLines, "-- "+line)
			}
			preambleLines = append(preambleLines, "-- +++")
			preamble := "-- +++\n" + strings.Join(preambleLines, "\n") + "\n"

			// Old format
			contents[filepath.Join(tempDirectory, "flat", schema.Name, fmt.Sprintf("%d_%s.up.sql", version, definition.Name))] = preamble + definition.UpQuery.Query(sqlf.PostgresBindVar)
			contents[filepath.Join(tempDirectory, "flat", schema.Name, fmt.Sprintf("%d_%s.down.sql", version, definition.Name))] = definition.DownQuery.Query(sqlf.PostgresBindVar)

			// New format
			contents[filepath.Join(tempDirectory, "dirs", schema.Name, strconv.Itoa(version), "metadata.yaml")] = metadata
			contents[filepath.Join(tempDirectory, "dirs", schema.Name, strconv.Itoa(version), "up.sql")] = definition.UpQuery.Query(sqlf.PostgresBindVar)
			contents[filepath.Join(tempDirectory, "dirs", schema.Name, strconv.Itoa(version), "down.sql")] = definition.DownQuery.Query(sqlf.PostgresBindVar)
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
