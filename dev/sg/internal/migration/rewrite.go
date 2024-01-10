package migration

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/stitch"
)

func Rewrite(database db.Database, rev string) error {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return err
	}
	migrationsDir := filepath.Join(repoRoot, "migrations", database.Name)

	ma := stitch.NewLazyMigrationsReader()
	fs, err := stitch.ReadMigrations(ma, database.Name, rev)
	if err != nil {
		return err
	}

	migrationsDirTemp := migrationsDir + ".working"
	defer func() {
		_ = os.RemoveAll(migrationsDirTemp)
	}()

	rootDir, err := http.FS(fs).Open("/")
	if err != nil {
		return err
	}
	defer func() { _ = rootDir.Close() }()

	migrations, err := rootDir.Readdir(0)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := os.MkdirAll(filepath.Join(migrationsDirTemp, migration.Name()), os.ModePerm); err != nil {
			return err
		}

		for _, basename := range []string{"up.sql", "down.sql", "metadata.yaml"} {
			f, err := fs.Open(filepath.Join(migration.Name(), basename))
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()

			contents, err := io.ReadAll(f)
			if err != nil {
				return err
			}

			filename := filepath.Join(migrationsDirTemp, migration.Name(), basename)
			std.Out.Writef("Writing %s", filename)

			if err := os.WriteFile(
				filename,
				[]byte(definition.CanonicalizeQuery(string(contents))),
				os.ModePerm,
			); err != nil {
				return err
			}
		}
	}

	if err := os.RemoveAll(migrationsDir); err != nil {
		return err
	}

	std.Out.Writef("Renaming %s -> %s", migrationsDirTemp, migrationsDir)

	if err := os.Rename(migrationsDirTemp, migrationsDir); err != nil {
		return err
	}

	return nil
}
