package migrations

import (
	"fmt"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

// DO NOT HAND EDIT
// THis constant is updated automatically by the ./dev/squash_migrations.sh script.
const FirstMigration = 1528395702

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

func TestNeedsGenerate(t *testing.T) {
	want, err := filepath.Glob(fmt.Sprintf("*.sql"))
	if err != nil {
		t.Fatal(err)
	}

	assetNames := AssetNames()
	sort.Strings(want)
	sort.Strings(assetNames)
	if !reflect.DeepEqual(assetNames, want) {
		t.Fatal("bindata out of datea Please run:\n  go generate github.com/sourcegraph/sourcegraph/migrations/...")
	}
}
