package migration

import (
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

// LeavesForCommit prints the leaves defined at the given commit (for every schema).
func LeavesForCommit(databases []db.Database, commit string) error {
	leavesByDatabase := make(map[string][]definition.Definition, len(databases))
	for _, database := range databases {
		definitions, err := readDefinitions(database)
		if err != nil {
			return err
		}

		leaves, err := selectLeavesForCommit(database, definitions, commit)
		if err != nil {
			return err
		}

		leavesByDatabase[database.Name] = leaves
	}

	for name, leaves := range leavesByDatabase {
		block := std.Out.Block(output.Styledf(output.StyleBold, "Leaf migrations for %q defined at commit %q", name, commit))
		for _, leaf := range leaves {
			block.Writef("%d: (%s)", leaf.ID, leaf.Name)
		}
		block.Close()
	}

	return nil
}

// selectLeavesForCommit selects the leaf definitions defined at the given commit for the
// gvien database.
func selectLeavesForCommit(database db.Database, ds *definition.Definitions, commit string) ([]definition.Definition, error) {
	migrationsDir := filepath.Join("migrations", database.Name)

	gitCmdOutput, err := run.GitCmd("ls-tree", "-r", "--name-only", commit, migrationsDir)
	if err != nil {
		return nil, err
	}

	ds, err = ds.Filter(parseVersions(strings.Split(gitCmdOutput, "\n"), migrationsDir))
	if err != nil {
		return nil, err
	}

	return ds.Leaves(), nil
}
