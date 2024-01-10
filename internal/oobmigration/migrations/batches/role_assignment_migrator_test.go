package batches

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestUserRoleAssignmentMigrator(t *testing.T) {
	ctx := context.Background()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))
	store := basestore.NewWithHandle(db.Handle())

	migrator := NewUserRoleAssignmentMigrator(store, 5)
	progress, err := migrator.Progress(ctx, false)
	require.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress with no DB entries, want=%f have=%f", want, have)
	}

	user1 := createTestUser(t, db, "testuser-1", true)
	user2 := createTestUser(t, db, "testuser-2", false)
	user3 := createTestUser(t, db, "testuser-3", true)

	users := []*types.User{user1, user2, user3}

	{
		// We calculate the progress when none of the created users have roles assigned to them.
		progress, err = migrator.Progress(ctx, false)
		require.NoError(t, err)

		// No user is assigned a role, so the progress should be 0.0.
		if have, want := progress, 0.0; have != want {
			t.Fatalf("got invalid progress with unmigrated entries, want=%f have=%f", want, have)
		}
	}

	{
		// We assign the USER role to `testuser-0` to simulate a bug in which not all permissions were assigned to a user during OOB.
		// This most likely occurred because a restart happened while the OOB migration was in progress.
		db.UserRoles().AssignSystemRole(ctx, database.AssignSystemRoleOpts{
			Role:   types.UserSystemRole,
			UserID: user1.ID,
		})

		// We calculate the progress when none of the created users have roles assigned to them.
		progress, err = migrator.Progress(ctx, false)
		require.NoError(t, err)

		// User1 is a site admin that has the USER role assigned to them, they need to have a SITE_ADMINISTRATOR role assigned to them also.
		// While User2 requires the USER role assigned to them since they aren't a site admin.
		// User3 requires both USER and SITE_ADMINISTRATOR role assigned to them.

		// This means only one role out of 5 roles that should be assigned is assigned. That's 1/5 = 0.2
		if have, want := progress, 0.2; have != want {
			t.Fatalf("got invalid progress with unmigrated entries, want=%f have=%f", want, have)
		}
	}

	if err := migrator.Up(ctx); err != nil {
		t.Fatal(err)
	}

	progress, err = migrator.Progress(ctx, false)
	require.NoError(t, err)

	if have, want := progress, 1.0; have != want {
		t.Fatalf("got invalid progress after up migration, want=%f have=%f", want, have)
	}

	userRole, err := db.Roles().Get(ctx, database.GetRoleOpts{
		Name: string(types.UserSystemRole),
	})
	require.NoError(t, err)

	siteAdminRole, err := db.Roles().Get(ctx, database.GetRoleOpts{
		Name: string(types.SiteAdministratorSystemRole),
	})
	require.NoError(t, err)

	for _, user := range users {
		assertRolesForUser(ctx, t, db, user, userRole, siteAdminRole)
	}
}

func createTestUser(t *testing.T, db database.DB, username string, siteAdmin bool) *types.User {
	t.Helper()

	user := &types.User{
		Username: username,
	}

	q := sqlf.Sprintf("INSERT INTO users (username, site_admin) VALUES (%s, %t) RETURNING id, site_admin", user.Username, siteAdmin)
	err := db.QueryRowContext(context.Background(), q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&user.ID, &user.SiteAdmin)
	if err != nil {
		t.Fatal(err)
	}

	if user.SiteAdmin != siteAdmin {
		t.Fatalf("user.SiteAdmin=%t, but expected is %t", user.SiteAdmin, siteAdmin)
	}

	_, err = db.ExecContext(context.Background(), "INSERT INTO names(name, user_id) VALUES($1, $2)", user.Username, user.ID)
	if err != nil {
		t.Fatalf("failed to create name: %s", err)
	}

	return user
}

func assertRolesForUser(ctx context.Context, t *testing.T, db database.DB, user *types.User, userRole *types.Role, siteAdminRole *types.Role) {
	// Get roles for user1
	have, err := db.UserRoles().GetByUserID(ctx, database.GetUserRoleOpts{UserID: user.ID})
	if err != nil {
		t.Fatal(err)
	}

	want := []*types.UserRole{
		{UserID: user.ID, RoleID: userRole.ID},
	}

	if user.SiteAdmin {
		// if the user is a site admin, the site administrator role should be assigned to them.
		want = append(want, &types.UserRole{UserID: user.ID, RoleID: siteAdminRole.ID})
	}

	if diff := cmp.Diff(have, want, cmpopts.IgnoreFields(types.UserRole{}, "CreatedAt")); diff != "" {
		t.Error(diff)
	}
}
