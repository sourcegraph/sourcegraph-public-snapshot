package authz

import (
	"flag"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/db/dbtest"
)

var dsn = flag.String("dsn", "", "Database connection string to use in integration tests")

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()

	db, cleanup := dbtest.NewDB(t, *dsn)
	defer cleanup()

	for _, tc := range []struct {
		name string
		test func(*testing.T)
	}{
		{"Store/LoadUserPermissions", testStoreLoadUserPermissions(db)},
		{"Store/LoadRepoPermissions", testStoreLoadRepoPermissions(db)},
		{"Store/SetRepoPermissions", testStoreSetRepoPermissions(db)},
		{"Store/LoadUserPendingPermissions", testStoreLoadUserPendingPermissions(db)},
		{"Store/SetRepoPendingPermissions", testStoreSetRepoPendingPermissions(db)},
		{"Store/GrantPendingPermissions", testStoreGrantPendingPermissions(db)},
		{"Store/DatabaseDeadlocks", testStoreDatabaseDeadlocks(db)},
	} {
		t.Run(tc.name, tc.test)
	}
}
