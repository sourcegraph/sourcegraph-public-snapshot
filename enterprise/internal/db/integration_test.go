package db

import (
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

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
		{"PermsStore/LoadUserPermissions", testPermsStore_LoadUserPermissions(db)},
		{"PermsStore/LoadRepoPermissions", testPermsStore_LoadRepoPermissions(db)},
		{"PermsStore/SetUserPermissions", testPermsStore_SetUserPermissions(db)},
		{"PermsStore/SetRepoPermissions", testPermsStore_SetRepoPermissions(db)},
		{"PermsStore/LoadUserPendingPermissions", testPermsStore_LoadUserPendingPermissions(db)},
		{"PermsStore/SetRepoPendingPermissions", testPermsStore_SetRepoPendingPermissions(db)},
		{"PermsStore/ListPendingUsers", testPermsStore_ListPendingUsers(db)},
		{"PermsStore/GrantPendingPermissions", testPermsStore_GrantPendingPermissions(db)},
		{"PermsStore/DeleteAllUserPermissions", testPermsStore_DeleteAllUserPermissions(db)},
		{"PermsStore/DeleteAllUserPendingPermissions", testPermsStore_DeleteAllUserPendingPermissions(db)},
		{"PermsStore/DatabaseDeadlocks", testPermsStore_DatabaseDeadlocks(db)},

		{"PermsStore/ListExternalAccounts", testPermsStore_ListExternalAccounts(db)},
		{"PermsStore/GetUserIDsByExternalAccounts", testPermsStore_GetUserIDsByExternalAccounts(db)},

		{"PermsStore/UserIDsWithNoPerms", testPermsStore_UserIDsWithNoPerms(db)},
		{"PermsStore/RepoIDsWithNoPerms", testPermsStore_RepoIDsWithNoPerms(db)},
		{"PermsStore/UserIDsWithOldestPerms", testPermsStore_UserIDsWithOldestPerms(db)},
		{"PermsStore/ReposIDsWithOldestPerms", testPermsStore_ReposIDsWithOldestPerms(db)},
		{"PermsStore/Metrics", testPermsStore_Metrics(db)},
	} {
		t.Run(tc.name, tc.test)
	}
}
