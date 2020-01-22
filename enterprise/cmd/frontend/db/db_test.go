package db

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "enterprisedb"
}

func equal(t testing.TB, name string, want, have interface{}) {
	t.Helper()
	if !reflect.DeepEqual(want, have) {
		t.Fatalf("%q: %s", name, cmp.Diff(want, have))
	}
}
