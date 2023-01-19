package batches

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestRoleAssignmentMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := basestore.NewWithHandle(db.Handle())

	migrator := NewRoleAssignmentMigrator(store, 5)
	progress, err := migrator.Progress(ctx, false)
	assert.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	if err = store.Exec(ctx, sqlf.Sprintf(`
		INSERT INTO users (username, display_name, created_at, site_admin)
		VALUES
			(%s, %s, NOW(), %s),
			(%s, %s, NOW(), %s)
	`,
		"testuser-0",
		"testuser",
		true,
		"testuser-1",
		"testuser1",
		false,
	)); err != nil {
		t.Fatal(err)
	}

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

	// Three records should be inserted into the user_roles table:
	// 1. For testuser-0 with DEFAULT role
	// 2. For testuser-0 WITH SITE_ADMINISTRATOR role
	// 3. For testuser-1 WITH DEFAULT role
	q := `SELECT role_id, user_id FROM user_roles ORDER BY user_id, role_id`
	rows, err := db.QueryContext(ctx, q)
	assert.NoError(t, err)
	defer rows.Close()
	var have []*types.UserRole
	for rows.Next() {
		var ur = types.UserRole{}
		if err := rows.Scan(&ur.RoleID, &ur.UserID); err != nil {
			t.Fatal(err, "error scanning user role")
		}
		have = append(have, &ur)
	}

	want := []*types.UserRole{
		{UserID: 1, RoleID: 1},
		{UserID: 1, RoleID: 2},
		{UserID: 2, RoleID: 1},
	}

	assert.Len(t, have, 3)
	if diff := cmp.Diff(have, want); diff != "" {
		t.Error(diff)
	}
}
