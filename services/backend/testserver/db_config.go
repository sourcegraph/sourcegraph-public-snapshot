// +build exectest

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
	AppDBH    gorp.SqlExecutor
	appDBDone func()

	GraphDBH    gorp.SqlExecutor
	graphDBDone func()
}

func (s *dbConfig) configDB() error {
	s.AppDBH, s.appDBDone = testdb.NewHandle("app", &localstore.AppSchema)
	if _, ok := s.AppDBH.(*dbutil2.Handle); !ok {
		return fmt.Errorf("test app requires a real app db *dbutil.Handle not %T (must run with -pgsqltest.init=full)", s.AppDBH)
	}
	s.GraphDBH, s.graphDBDone = testdb.NewHandle("graph", &localstore.GraphSchema)
	if _, ok := s.GraphDBH.(*dbutil2.Handle); !ok {
		return fmt.Errorf("test app requires a real graph db *dbutil.Handle not %T (must run with -pgsqltest.init=full)", s.GraphDBH)
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
	v = append(v, "SG_GRAPH_PGDATABASE="+parseDBName(s.GraphDBH.(*dbutil2.Handle).DataSource))
	v = append(v, "PGSSLMODE=disable")
	if u := os.Getenv("PGUSER"); u != "" {
		v = append(v, "PGUSER="+u)
	}
	return v
}

func (s *dbConfig) close() {
	s.appDBDone()
	s.graphDBDone()
}
