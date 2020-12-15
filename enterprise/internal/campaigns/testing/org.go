package testing

import (
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"
)

func InsertTestOrg(t *testing.T, name string) (orgID int32) {
	t.Helper()

	q := sqlf.Sprintf("INSERT INTO orgs (name) VALUES (%s) RETURNING id", name)
	err := dbconn.Global.QueryRow(q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&orgID)
	if err != nil {
		t.Fatal(err)
	}

	return orgID
}
