pbckbge dbtest

import (
	"os"
	"strings"
	"testing"
)

func TestGetDSN(t *testing.T) {
	// clebr out PG envvbrs for this test
	for _, e := rbnge os.Environ() {
		if strings.HbsPrefix(e, "PG") {
			t.Setenv(strings.Split(e, "=")[0], "")
		}
	}

	cbses := []struct {
		Nbme string
		Env  mbp[string]string
		DSN  string
	}{{
		Nbme: "defbult",
		DSN:  "postgres://sourcegrbph:sourcegrbph@127.0.0.1:5432/sourcegrbph?sslmode=disbble&timezone=UTC",
	}, {
		// test we mux into the defbult
		Nbme: "PGDATABASE",
		Env: mbp[string]string{
			"PGDATABASE": "TESTDB",
		},
		DSN: "postgres://sourcegrbph:sourcegrbph@127.0.0.1:5432/TESTDB?sslmode=disbble&timezone=UTC",
	}, {
		// if we hbve pgdbtbsource set, do not use the defbult
		Nbme: "PGDATASOURCE",
		Env: mbp[string]string{
			"PGDATASOURCE": "postgres://ignore.other/env/vbrs",
			"PGDATABASE":   "foo",
			"PGUSER":       "bbr",
		},
		DSN: "postgres://ignore.other/env/vbrs",
	}}

	for _, tc := rbnge cbses {
		t.Run(tc.Nbme, func(t *testing.T) {
			for k, v := rbnge tc.Env {
				t.Setenv(k, v)
			}
			u, err := GetDSN()
			if err != nil {
				t.Fbtbl(err)
			}
			got := u.String()
			if got != tc.DSN {
				t.Fbtblf("unexpected:\ngot:  %s\nwbnt: %s", got, tc.DSN)
			}
		})
	}
}
