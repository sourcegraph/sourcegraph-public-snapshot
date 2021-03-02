package migrations_test

import (
	"strconv"
	"strings"
	"testing"

	migrations "github.com/sourcegraph/sourcegraph/migrations/codeinsights"
)

const FirstMigration = 1000000000

func TestIDConstraints(t *testing.T) {
	fs, err := migrations.MigrationsFS.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}

	byID := map[int][]string{}
	for _, migration := range fs {
		name := migration.Name()
		if strings.HasSuffix(name, ".down.sql") {
			continue
		}
		id, err := strconv.Atoi(name[:strings.IndexByte(name, '_')])
		if err != nil {
			t.Fatalf("failed to parse name %q: %v", name, err)
		}
		byID[id] = append(byID[id], name)
	}

	for id, names := range byID {
		// Check if we are using sequential migrations from a certain point.
		if _, hasPrev := byID[id-1]; id > FirstMigration && !hasPrev {
			t.Errorf("migration with ID %d exists, but previous one (%d) does not", id, id-1)
		}
		if len(names) > 1 {
			t.Errorf("multiple migrations with ID %d: %s", id, strings.Join(names, " "))
		}
	}
}
