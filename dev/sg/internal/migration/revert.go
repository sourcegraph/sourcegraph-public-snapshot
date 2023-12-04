package migration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// Revert creates a new migration that reverts the set of migrations from a target commit.
func Revert(databases []db.Database, commit string) error {
	versionsByDatabase := make(map[string][]int, len(databases))
	for _, database := range databases {
		versions, err := selectMigrationsDefinedInCommit(database, commit)
		if err != nil {
			return err
		}

		versionsByDatabase[database.Name] = versions
	}

	redacted := false
	for dbName, versions := range versionsByDatabase {
		if len(versions) == 0 {
			continue
		}
		redacted = true

		var (
			database, _ = db.DatabaseByName(dbName)
			upPaths     = make([]string, 0, len(versions))
			downQueries = make([]string, 0, len(versions))
		)

		defs, err := readDefinitions(database)
		if err != nil {
			return err
		}

		for _, version := range versions {
			def, ok := defs.GetByID(version)
			if !ok {
				return errors.Newf("could not find migration %d in database %q", version, dbName)
			}

			files, err := makeMigrationFilenames(database, version, def.Name)
			if err != nil {
				return err
			}

			downQuery, err := os.ReadFile(files.DownFile)
			if err != nil {
				return err
			}
			upPaths = append(upPaths, files.UpFile)
			downQueries = append(downQueries, string(downQuery))

			contents := map[string]string{
				files.UpFile: "-- REDACTED\n",
			}
			if err := writeMigrationFiles(contents); err != nil {
				return err
			}
		}

		block := std.Out.Block(output.Styled(output.StyleBold, "Migration files redacted"))
		for _, path := range upPaths {
			block.Writef("Up query file: %s", path)
		}
		block.Close()

		if err := AddWithTemplate(database, fmt.Sprintf("revert %s", commit), strings.Join(downQueries, "\n\n"), "-- No-op\n"); err != nil {
			return err
		}
	}
	if !redacted {
		return errors.Newf("No migrations defined on commit %q", commit)
	}

	return nil
}

// selectMigrationsDefinedInCommit returns the identifiers of migrations defined in the given
// commit for the given schema.a
func selectMigrationsDefinedInCommit(database db.Database, commit string) ([]int, error) {
	migrationsDir := filepath.Join("migrations", database.Name)

	gitCmdOutput, err := run.GitCmd("diff", "--name-only", commit+".."+commit+"~1", migrationsDir)
	if err != nil {
		return nil, err
	}

	versions := parseVersions(strings.Split(gitCmdOutput, "\n"), migrationsDir)
	return versions, nil
}
