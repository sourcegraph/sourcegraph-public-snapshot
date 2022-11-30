package dbtest

import (
	"os"
	"strings"
	"testing"
)

func TestGetDSN(t *testing.T) {
	// clear out PG envvars for this test
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "PG") {
			t.Setenv(strings.Split(e, "=")[0], "")
		}
	}

	cases := []struct {
		Name string
		Env  map[string]string
		DSN  string
	}{{
		Name: "default",
		DSN:  "postgres://sourcegraph:sourcegraph@127.0.0.1:5432/sourcegraph?sslmode=disable&timezone=UTC",
	}, {
		// test we mux into the default
		Name: "PGDATABASE",
		Env: map[string]string{
			"PGDATABASE": "TESTDB",
		},
		DSN: "postgres://sourcegraph:sourcegraph@127.0.0.1:5432/TESTDB?sslmode=disable&timezone=UTC",
	}, {
		// if we have pgdatasource set, do not use the default
		Name: "PGDATASOURCE",
		Env: map[string]string{
			"PGDATASOURCE": "postgres://ignore.other/env/vars",
			"PGDATABASE":   "foo",
			"PGUSER":       "bar",
		},
		DSN: "postgres://ignore.other/env/vars",
	}}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			for k, v := range tc.Env {
				t.Setenv(k, v)
			}
			u, err := GetDSN()
			if err != nil {
				t.Fatal(err)
			}
			got := u.String()
			if got != tc.DSN {
				t.Fatalf("unexpected:\ngot:  %s\nwant: %s", got, tc.DSN)
			}
		})
	}
}
