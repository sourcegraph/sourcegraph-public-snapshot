package permissions

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestStore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()
	db := database.NewDB(dbtest.NewDB(t))

	ctx := context.Background()
	jobID, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project1", 2, []types.UserPermission{
		{BindID: "user1", Permission: "read"},
		{BindID: "user2", Permission: "admin"},
	}, false)
	require.NoError(t, err)
	require.NotZero(t, jobID)

	store := createBitbucketProjectPermissionsStore(db)
	count, err := store.QueuedCount(ctx, true, nil)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func intPtr(v int) *int              { return &v }
func stringPtr(v string) *string     { return &v }
func timePtr(v time.Time) *time.Time { return &v }

func mustParseTime(v string) time.Time {
	t, err := time.Parse("2006-01-02", v)
	if err != nil {
		panic(err)
	}
	return t
}

func TestGetBitbucketClient(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Parallel()
	ctx := context.Background()

	var c schema.BitbucketServerConnection
	c.Token = "secret"
	c.Url = "http://some-url"
	c.Username = "username"

	cfg, err := json.Marshal(&c)
	require.NoError(t, err)

	svc := types.ExternalService{
		Config: string(cfg),
	}

	var handler bitbucketProjectPermissionsHandler
	client, err := handler.getBitbucketClient(ctx, &svc)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestSetPermissionsForUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := context.Background()

	db := edb.NewEnterpriseDB(database.NewDB(dbtest.NewDB(t)))

	// create 3 users
	users := db.Users()
	igor, err := users.Create(ctx,
		database.NewUser{
			Email:           "igor@example.com",
			Username:        "igor",
			EmailIsVerified: true,
		},
	)
	require.NoError(t, err)
	pushpa, err := users.Create(ctx,
		database.NewUser{
			Email:           "pushpa@example.com",
			Username:        "pushpa",
			EmailIsVerified: true,
		},
	)
	require.NoError(t, err)
	_, err = users.Create(ctx,
		database.NewUser{
			Email:           "omar@example.com",
			Username:        "omar",
			EmailIsVerified: true,
		},
	)
	require.NoError(t, err)

	// create 3 repos
	repos := db.Repos()
	err = repos.Create(ctx, &types.Repo{
		ID:   1,
		Name: "sourcegraph/sourcegraph",
	})
	require.NoError(t, err)
	err = repos.Create(ctx, &types.Repo{
		ID:   2,
		Name: "sourcegraph/handbook",
	})
	require.NoError(t, err)
	err = repos.Create(ctx, &types.Repo{
		ID:   3,
		Name: "sourcegraph/src-cli",
	})
	require.NoError(t, err)

	check := func() {
		// check that the permissions were set
		perms := db.Perms()

		p := authz.RepoPermissions{RepoID: 1, Perm: authz.Read}
		err = perms.LoadRepoPermissions(ctx, &p)
		require.NoError(t, err)
		require.Equal(t, map[int32]struct{}{
			pushpa.ID: {},
			igor.ID:   {},
		}, p.UserIDs)

		up := authz.UserPermissions{UserID: pushpa.ID, Perm: authz.Read, Type: authz.PermRepos}
		err = perms.LoadUserPermissions(ctx, &up)
		require.NoError(t, err)
		require.Equal(t, map[int32]struct{}{
			1: {},
			2: {},
		}, up.IDs)
	}

	// set permissions for 3 users (2 existing, 1 pending) and 2 repos
	err = setPermissionsForUsers(
		ctx,
		db,
		[]types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "sayako", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
	)
	require.NoError(t, err)
	check()

	// run the same set of permissions again, shouldn't change anything
	err = setPermissionsForUsers(
		ctx,
		db,
		[]types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "sayako", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
	)
	require.NoError(t, err)
	check()

	// test with wrong bindids
	err = setPermissionsForUsers(
		ctx,
		db,
		[]types.UserPermission{
			{BindID: "pushpa", Permission: "read"},
			{BindID: "igor", Permission: "read"},
			{BindID: "sayako", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
	)
	// should fail if the bind ids are wrong
	require.Error(t, err)

	// ensure this unsets the unrestricted flag
	_, err = db.ExecContext(ctx, "UPDATE repo_permissions SET unrestricted = true WHERE repo_id = 1")
	require.NoError(t, err)

	// run the same set of permissions again
	err = setPermissionsForUsers(
		ctx,
		db,
		[]types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "sayako", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
	)
	require.NoError(t, err)
	check()

	// check the unrestricted flag
	var unrestricted bool
	err = db.QueryRowContext(ctx, "SELECT unrestricted FROM repo_permissions WHERE repo_id = 1").Scan(&unrestricted)
	require.NoError(t, err)
	require.False(t, unrestricted)
}
