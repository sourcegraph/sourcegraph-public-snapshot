package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/shared"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/stitch"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

func main() {
	if err := mainErr(); err != nil {
		panic(fmt.Sprintf("error: %s", err))
	}
}

func mainErr() error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// This script is invoked via a go:generate directive in internal/database/migration/shared (embed.go)
	repoRoot := filepath.Join(wd, "..", "..", "..", "..")

	//
	// Write stitched migrations

	versions, err := oobmigration.UpgradeRange(MinVersion, MaxVersion)
	if err != nil {
		return err
	}
	versionTags := make([]string, 0, len(versions))
	for _, version := range versions {
		versionTags = append(versionTags, version.GitTag())
	}
	if err := stitchAndWrite(repoRoot, filepath.Join(wd, "data", "stitched-migration-graph.json"), versionTags); err != nil {
		return err
	}

	//
	// Write frozen migrations

	for _, rev := range FrozenRevisions {
		if err := stitchAndWrite(repoRoot, filepath.Join(wd, "data", "frozen", fmt.Sprintf("%s.json", rev)), []string{rev}); err != nil {
			return err
		}
	}

	return nil
}

func stitchAndWrite(repoRoot, filepath string, versionTags []string) error {
	stitchedMigrationBySchemaName := map[string]shared.StitchedMigration{}
	for _, schemaName := range schemas.SchemaNames {
		stitched, err := stitch.StitchDefinitions(schemaName, repoRoot, versionTags)
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
