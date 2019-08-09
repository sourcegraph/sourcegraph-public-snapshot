package cli

import (
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf/conftypes"
)

func TestPostgresDSN(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		dsn  string
	}{{
		name: "default",
		env:  map[string]string{},
		dsn:  "postgres://testuser@127.0.0.1:5432",
	}, {
		name: "deploy-sourcegraph",
		env: map[string]string{
			"PGDATABASE": "sg",
			"PGHOST":     "pgsql",
			"PGPORT":     "5432",
			"PGSSLMODE":  "disable",
			"PGUSER":     "sg",
		},
		dsn: "postgres://sg@pgsql:5432/sg?sslmode=disable",
	}, {
		name: "deploy-sourcegraph password",
		env: map[string]string{
			"PGDATABASE": "sg",
			"PGHOST":     "pgsql",
			"PGPASSWORD": "REDACTED",
			"PGPORT":     "5432",
			"PGSSLMODE":  "disable",
			"PGUSER":     "sg",
		},
		dsn: "postgres://sg:REDACTED@pgsql:5432/sg?sslmode=disable",
	}, {
		name: "sourcegraph server",
		env: map[string]string{
			"PGHOST":     "127.0.0.1",
			"PGUSER":     "postgres",
			"PGDATABASE": "sourcegraph",
			"PGSSLMODE":  "disable",
		},
		dsn: "postgres://postgres@127.0.0.1/sourcegraph?sslmode=disable",
	}, {
		name: "datasource",
		env: map[string]string{
			"PGDATASOURCE": "postgres://foo@bar/bam",
		},
		dsn: "postgres://foo@bar/bam",
	}, {
		name: "datasource ignore",
		env: map[string]string{
			"PGHOST":       "pgsql",
			"PGDATASOURCE": "postgres://foo@bar/bam",
		},
		dsn: "postgres://foo@bar/bam",
	}}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			have := postgresDSN("testuser", func(e string) string {
				return tc.env[e]
			})
			if have != tc.dsn {
				t.Errorf("unexpected computed DSN\nhave: %s\nwant: %s", have, tc.dsn)
			}
		})
	}
}

func TestServiceConnections(t *testing.T) {
	// We only test that we get something non-empty back.
	sc := serviceConnections()
	if reflect.DeepEqual(sc, conftypes.ServiceConnections{}) {
		t.Fatal("expected non-empty service connections")
	}
}

// random will create a file of size bytes (rounded up to next 1024 size)
func random_324(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
