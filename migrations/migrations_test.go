package migrations_test

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	codeintel "github.com/sourcegraph/sourcegraph/migrations/codeintel"
	frontend "github.com/sourcegraph/sourcegraph/migrations/frontend"
)

// DO NOT HAND EDIT
// These constants are updated automatically by the ./dev/squash_migrations.sh script.

const frontendFirstMigration = 1528395702
const codeintelFirstMigration = 1000000001

var databases = map[string]struct {
	FirstMigration int
	AssetNames     []string
}{
	"frontend":  {frontendFirstMigration, frontend.AssetNames()},
	"codeintel": {codeintelFirstMigration, codeintel.AssetNames()},
}

func TestIDConstraints(t *testing.T) {
	for databaseName, db := range databases {
		ups, err := filepath.Glob(fmt.Sprintf("%s/*.up.sql", databaseName))
		if err != nil {
			t.Fatal(err)
		}

		byID := map[int][]string{}
		for _, name := range ups {
			id, err := strconv.Atoi(name[len(databaseName)+1 : strings.IndexByte(name, '_')])
			if err != nil {
				t.Fatalf("failed to parse name %q: %v", name, err)
			}
			byID[id] = append(byID[id], name)
		}

		for id, names := range byID {
			// Check if we are using sequential migrations from a certain point.
			if _, hasPrev := byID[id-1]; id > db.FirstMigration && !hasPrev {
				t.Errorf("migration with ID %d exists, but previous one (%d) does not", id, id-1)
			}
			if len(names) > 1 {
				t.Errorf("multiple migrations with ID %d: %s", id, strings.Join(names, " "))
			}
		}
	}
}

func TestNeedsGenerate(t *testing.T) {
	for databaseName, db := range databases {
		want, err := filepath.Glob(fmt.Sprintf("%s/*.sql", databaseName))
		if err != nil {
			t.Fatal(err)
		}
		for i, name := range want {
			want[i] = name[len(databaseName)+1:]
		}

		sort.Strings(want)
		sort.Strings(db.AssetNames)
		if !reflect.DeepEqual(db.AssetNames, want) {
			t.Fatal("bindata out of date. Please run:\n  go generate github.com/sourcegraph/sourcegraph/migrations/...")
		}
	}
}
