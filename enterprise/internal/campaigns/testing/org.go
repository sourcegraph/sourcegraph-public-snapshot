package testing

import (
	"database/sql"
	"testing"
)

func InsertTestOrg(t *testing.T, db *sql.DB) (orgID int32) {
	t.Helper()

	err := db.QueryRow("INSERT INTO orgs (name) VALUES ('bbs-org') RETURNING id").Scan(&orgID)
	if err != nil {
		t.Fatal(err)
	}

	return orgID
}
