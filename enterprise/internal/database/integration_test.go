package database

import (
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

// Toggles particularly slow tests. To enable, use `go test` with this flag, for example:
//
//   go test -timeout 360s -v -run ^TestIntegration_PermsStore$ github.com/sourcegraph/sourcegraph/enterprise/internal/database -slow-tests
//
var slowTests = flag.Bool("slow-tests", false, "Enable very slow tests")

// postgresParameterLimitTest names tests that are focused on ensuring the default
// behaviour of various queries do not run into the Postgres parameter limit at scale
// (error `extended protocol limited to 65535 parameters`).
//
// They are typically flagged behind `-slow-tests` - when changing queries make sure to
// enable these tests and add more where relevant (see `slowTests`).
const postgresParameterLimitTest = "ensure we do not exceed postgres parameter limit"

func TestIntegration_PermsStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db := dbtest.NewDB(t, *dsn)

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"LoadUserPermissions", testPermsStore_LoadUserPermissions(db)},
		{"LoadRepoPermissions", testPermsStore_LoadRepoPermissions(db)},
		{"SetUserPermissions", testPermsStore_SetUserPermissions(db)},
		{"SetRepoPermissions", testPermsStore_SetRepoPermissions(db)},
		{"TouchRepoPermissions", testPermsStore_TouchRepoPermissions(db)},
		{"LoadUserPendingPermissions", testPermsStore_LoadUserPendingPermissions(db)},
		{"SetRepoPendingPermissions", testPermsStore_SetRepoPendingPermissions(db)},
		{"ListPendingUsers", testPermsStore_ListPendingUsers(db)},
		{"GrantPendingPermissions", testPermsStore_GrantPendingPermissions(db)},
		{"SetPendingPermissionsAfterGrant", testPermsStore_SetPendingPermissionsAfterGrant(db)},
		{"DeleteAllUserPermissions", testPermsStore_DeleteAllUserPermissions(db)},
		{"DeleteAllUserPendingPermissions", testPermsStore_DeleteAllUserPendingPermissions(db)},
		{"DatabaseDeadlocks", testPermsStore_DatabaseDeadlocks(db)},

		{"ListExternalAccounts", testPermsStore_ListExternalAccounts(db)},
		{"GetUserIDsByExternalAccounts", testPermsStore_GetUserIDsByExternalAccounts(db)},

		{"UserIDsWithOutdatedPerms", testPermsStore_UserIDsWithOutdatedPerms(db)},
		{"UserIDsWithNoPerms", testPermsStore_UserIDsWithNoPerms(db)},
		{"RepoIDsWithNoPerms", testPermsStore_RepoIDsWithNoPerms(db)},
		{"UserIDsWithOldestPerms", testPermsStore_UserIDsWithOldestPerms(db)},
		{"ReposIDsWithOldestPerms", testPermsStore_ReposIDsWithOldestPerms(db)},
		{"Metrics", testPermsStore_Metrics(db)},
	} {
		t.Run(tc.name, tc.test)
	}
}
