package batches

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log/logtest"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRolePermissionAssignmentMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := basestore.NewWithHandle(db.Handle())

	migrator := NewRolePermissionAssignmentMigrator(store)
	progress, err := migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	seedPermissions(ctx, t, store)

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 0.0; have != want {
		t.Fatalf("got invalid progress with one unmigrated entry, want=%f have=%f", want, have)
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress after up migration, want=%f have=%f", want, have)
	}

	// Six records should be inserted into the role_permissions table.
	// Two default roles multiplied by 3 seeded permissions in the `seedPermissions`
	// method

	query := `SELECT role_id, permission_id FROM role_permissions ORDER BY role_id, permission_id`
	rows, err := db.QueryContext(ctx, query)
	assert.NoError(t, err)
	defer rows.Close()

	var have []*types.RolePermission
	for rows.Next() {
		var rp = types.RolePermission{}
		if err := rows.Scan(&rp.RoleID, &rp.PermissionID); err != nil {
			t.Fatal(err, "error scanning role permission")
		}
		have = append(have, &rp)
	}

	want := []*types.RolePermission{
		{RoleID: 1, PermissionID: 1},
		{RoleID: 1, PermissionID: 2},
		{RoleID: 1, PermissionID: 3},
		{RoleID: 2, PermissionID: 1},
		{RoleID: 2, PermissionID: 2},
		{RoleID: 2, PermissionID: 3},
	}

	assert.Len(t, have, 6)
	if diff := cmp.Diff(have, want); diff != "" {
		t.Error(diff)
	}
}

func seedPermissions(ctx context.Context, t *testing.T, store *basestore.Store) {
	if err := store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO permissions (namespace, action)
		VALUES
			(%s, %s),
			(%s, %s),
			(%s, %s)
	`,
		"TEST-NAMSPACE-1",
		"READ",
		"TEST-NAMESPACE-1",
		"WRITE",
		"TEST-NAMESPACE-2",
		"READ",
	)); err != nil {
		t.Fatal(err)
	}
}
