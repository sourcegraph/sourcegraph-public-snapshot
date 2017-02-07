package testserver

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/gorp.v1"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/dbutil2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/testdb"
	"sourcegraph.com/sourcegraph/sourcegraph/services/backend/internal/localstore"
)

// dbConfig is embedded in TestServer.
type dbConfig struct {
	AppDBH gorp.SqlExecutor
}

func (s *dbConfig) configDB() error {
	s.AppDBH = testdb.NewHandle("app", &localstore.AppSchema)
	if _, ok := s.AppDBH.(*dbutil2.Handle); !ok {
		return fmt.Errorf("test app requires a real app db *dbutil.Handle not %T (must run with -pgsqltest.init=full)", s.AppDBH)
	}
	return nil
}

func (s *dbConfig) dbEnvConfig() []string {
	parseDBName := func(s string) string {
		fs := strings.Fields(s)
		for _, f := range fs {
			if strings.HasPrefix(f, "dbname=") {
				return strings.TrimPrefix(f, "dbname=")
			}
		}
		panic("no dbname= found in data source: '" + s + "'")
	}
	v := []string{"PGDATABASE=" + parseDBName(s.AppDBH.(*dbutil2.Handle).DataSource)}
	v = append(v, "PGSSLMODE=disable")
	if u := os.Getenv("PGUSER"); u != "" {
		v = append(v, "PGUSER="+u)
	}
	if u := os.Getenv("PGPASSWORD"); u != "" {
		v = append(v, "PGPASSWORD="+u)
	}
	if u := os.Getenv("PGPORT"); u != "" {
		v = append(v, "PGPORT="+u)
	}
	return v
}
