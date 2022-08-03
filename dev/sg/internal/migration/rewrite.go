package migration

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
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

	fs, err := stitch.ReadMigrations(database.Name, repoRoot, rev)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(migrationsDir); err != nil {
		return err
	}

	root, err := http.FS(fs).Open("/")
	if err != nil {
		return err
	}
	defer func() { _ = root.Close() }()

	migrations, err := root.Readdir(0)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		if err := os.MkdirAll(filepath.Join(migrationsDir, migration.Name()), os.ModePerm); err != nil {
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

			if err := os.WriteFile(
				filepath.Join(migrationsDir, migration.Name(), basename),
				[]byte(definition.CanonicalizeQuery(string(contents))),
				os.ModePerm,
			); err != nil {
				return err
			}
		}
	}

	return nil
}
