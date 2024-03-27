package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/stitch"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func main() {
	liblog := log.Init(log.Resource{Name: "stitched-migration-generator"})
	defer liblog.Sync()

	if err := mainErr(); err != nil {
		panic(fmt.Sprintf("error: %s", err))
	}
}

var frozenMigrationsFlag = flag.Bool("write-frozen", true, "write frozen revision migration files")
var outputPath = flag.String("output", "data/stitched-migration-graph.json", "where to put the stitched migration graph JSON")
var frozenOutputPath = flag.String("frozen-output", "data/frozen", "where to put the stitched migration graph JSON")
var archivePath = flag.String("archive", "", "where to find migration dump")

func mainErr() error {
	flag.Parse()

	if *archivePath == "" {
		return errors.New("missing -archive flag")
	}

	// Write stitched migrations
	versions, err := oobmigration.UpgradeRange(MinVersion, MaxVersion)
	if err != nil {
		return err
	}
	versionTags := make([]string, 0, len(versions))
	for _, version := range versions {
		versionTags = append(versionTags, version.GitTag())
	}
	fmt.Printf("Generating stitched migration files for range [%s, %s]\n", MinVersion, MaxVersion)
	if err := stitchAndWrite(*archivePath, *outputPath, versionTags); err != nil {
		return err
	}

	if *frozenMigrationsFlag {
		fmt.Println("Generating frozen migrations")
		// Write frozen migrations. There is an optional flag that will short circuit this step. This is useful for
		// clients that are only interested in the stitch graph, such as the release tool.
		for _, rev := range FrozenRevisions {
			if err := stitchAndWrite(*archivePath, filepath.Join(*frozenOutputPath, fmt.Sprintf("%s.json", rev)), []string{rev}); err != nil {
				return err
			}
		}
	}

	return nil
}

func stitchAndWrite(archivesPath string, filepath string, versionTags []string) error {
	stitchedMigrationBySchemaName := map[string]shared.StitchedMigration{}
	ma, err := stitch.NewLocalMigrationsReader(archivesPath, maxVersionString)
	if err != nil {
		return err
	}
	for _, schemaName := range schemas.SchemaNames {
		stitched, err := stitch.StitchDefinitions(ma, schemaName, versionTags)
		if err != nil {
			return err
		}

		stitchedMigrationBySchemaName[schemaName] = stitched
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "\t")

	if err := encoder.Encode(stitchedMigrationBySchemaName); err != nil {
		return err
	}

	fmt.Printf("Wrote to %s\n", filepath)
	return nil
}
