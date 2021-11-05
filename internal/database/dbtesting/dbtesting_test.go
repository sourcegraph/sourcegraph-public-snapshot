package dbtesting

import "testing"

func TestDBName(t *testing.T) {
	want := "sourcegraph-test-internal-database-dbtesting"
	got, err := dbName()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Fatalf("unexpected dbname\ngot:  %q\nwant: %q", got, want)
	}
}
