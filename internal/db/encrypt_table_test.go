package db

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbconn"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtesting"
)

func TestTableRotateEncryption(t *testing.T) {

	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	db := dbconn.Global
	ctx := context.Background()
	err := insertTestData()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name    string
		table   string
		cols    []SecretColumn
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			"bootstrap",
			"external_service_repos",
			[]SecretColumn{{
				Name:     "clone_url",
				Nullable: false,
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := TableRotateEncryption(ctx, db, tt.table, tt.cols...); (err != nil) != tt.wantErr {
				t.Errorf("TableRotateEncryption() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func insertTestData() error {
	const testUser = `INSERT INTO users(id,username) VALUES (1,'test')`
	_, err := dbconn.Global.Exec(testUser)
	const testData = `INSERT INTO saved_searches(id,user_id,query,notify_owner,notify_slack,description) VALUES (1,1,'secretTokenString',false,false,false) `

	_, err = dbconn.Global.Exec(testData)
	return err
}
