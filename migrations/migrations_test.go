package migrations_test

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/migrations"
)

const FirstMigration = 1528395604

func TestIDConstraints(t *testing.T) {
	ups, err := filepath.Glob("*.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	byID := map[int][]string{}
	for _, name := range ups {
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

// Makes sure that every migration contains exactly one `COMMIT;` so that
// `InjectVersionUpdate` in internal/db/dbutil/dbutil.go is guaranteed to succeed.
func TestTransactions(t *testing.T) {
	ups, err := filepath.Glob("*.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, name := range ups {
		contents, err := ioutil.ReadFile(name)
		if err != nil {
			t.Fatalf("failed to read migration file %q: %v", name, err)
		}
		commitCount := strings.Count(string(contents), "COMMIT;")
		if commitCount != 1 {
			t.Fatalf("expected migration %q to contain exactly one COMMIT; but it contains %d", name, commitCount)
		}
	}
}

func TestNeedsGenerate(t *testing.T) {
	want, err := filepath.Glob("*.sql")
	if err != nil {
		t.Fatal(err)
	}
	got := migrations.AssetNames()
	sort.Strings(want)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		t.Fatal("bindata out of date. Please run:\n  go generate github.com/sourcegraph/sourcegraph/migrations")
	}
}
