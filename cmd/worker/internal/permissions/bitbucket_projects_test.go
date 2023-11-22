package permissions

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbmocks"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/bitbucketserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestStore(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(t))

	ctx := context.Background()
	jobID, err := db.BitbucketProjectPermissions().Enqueue(ctx, "project1", 2, []types.UserPermission{
		{BindID: "user1", Permission: "read"},
		{BindID: "user2", Permission: "admin"},
	}, false)
	require.NoError(t, err)
	require.NotZero(t, jobID)

	store := createBitbucketProjectPermissionsStore(observation.TestContextTB(t), db, &config{})
	count, err := store.QueuedCount(ctx, true)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

func TestGetBitbucketClient(t *testing.T) {
	t.Parallel()

	var c schema.BitbucketServerConnection
	c.Token = "secret"
	c.Url = "http://some-url"
	c.Username = "username"

	cfg, err := json.Marshal(&c)
	require.NoError(t, err)

	svc := types.ExternalService{
		Config: extsvc.NewUnencryptedConfig(string(cfg)),
	}

	ctx := context.Background()
	var handler bitbucketProjectPermissionsHandler
	client, err := handler.getBitbucketClient(ctx, logtest.Scoped(t), &svc)
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestHandle_UnsupportedCodeHost(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	externalServices := dbmocks.NewMockExternalServiceStore()
	externalServices.GetByIDFunc.SetDefaultReturn(
		&types.ExternalService{
			ID:          1,
			Kind:        extsvc.KindGitHub,
			DisplayName: "github",
			Config:      extsvc.NewEmptyConfig(),
		},
		nil,
	)

	db := dbmocks.NewMockDB()
	db.ExternalServicesFunc.SetDefaultReturn(externalServices)

	handler := &bitbucketProjectPermissionsHandler{db: db}
	err := handler.Handle(ctx, logtest.Scoped(t), &types.BitbucketProjectPermissionJob{ExternalServiceID: 1})

	require.True(t, errcode.IsNonRetryable(err))
}

func TestSetPermissionsForUsers(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))

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

		p, err := perms.LoadRepoPermissions(ctx, 1)
		require.NoError(t, err)
		gotIDs := make([]int32, len(p))
		for i, perm := range p {
			gotIDs[i] = perm.UserID
		}
		slices.Sort(gotIDs)

		require.Equal(t, []int32{igor.ID, pushpa.ID}, gotIDs)

		up, err := perms.LoadUserPermissions(ctx, pushpa.ID)
		require.NoError(t, err)
		gotIDs = make([]int32, len(up))
		for i, perm := range up {
			gotIDs[i] = perm.RepoID
		}
		slices.Sort(gotIDs)

		require.Equal(t, []int32{1, 2}, gotIDs)
	}

	checkPendingPerms := func(bindIDs []string) {
		perms := db.Perms()

		for _, bindID := range bindIDs {
			userPerms := &authz.UserPendingPermissions{
				ServiceType: authz.SourcegraphServiceType,
				ServiceID:   authz.SourcegraphServiceID,
				BindID:      bindID,
				Perm:        authz.Read,
				Type:        authz.PermRepos,
			}

			err := perms.LoadUserPendingPermissions(ctx, userPerms)
			require.NoError(t, err)
			require.Equal(t, []int32{1, 2}, userPerms.IDs.Sorted(collections.NaturalCompare[int32]))
		}
	}

	h := bitbucketProjectPermissionsHandler{db: db}
	// set permissions for 3 users (2 existing, 1 pending) and 2 repos
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "username1@foo.bar", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
		"foo",
	)
	require.NoError(t, err)
	check()
	checkPendingPerms([]string{"username1@foo.bar"})

	// run the same set of permissions again, shouldn't change anything
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "username1@foo.bar", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
		"foo",
	)
	require.NoError(t, err)
	check()
	checkPendingPerms([]string{"username1@foo.bar"})

	// test with only non-existent users
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "username1@foo.bar", Permission: "read"},
			{BindID: "username2@foo.bar", Permission: "read"},
			{BindID: "username3@foo.bar", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
		"foo",
	)
	// should fail if the bind ids are wrong
	require.NoError(t, err)
	checkPendingPerms([]string{"username1@foo.bar", "username2@foo.bar", "username3@foo.bar"})

	// ensure this unsets the unrestricted flag
	_, err = db.ExecContext(ctx, "UPDATE repo_permissions SET unrestricted = true WHERE repo_id = 1")
	require.NoError(t, err)

	// run the same set of permissions again
	err = h.setPermissionsForUsers(
		ctx,
		logtest.Scoped(t),
		[]types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "username1@foo.bar", Permission: "read"},
		},
		[]api.RepoID{
			1,
			2,
		},
		"foo",
	)
	require.NoError(t, err)
	check()

	// check the unrestricted flag
	var unrestricted bool
	err = db.QueryRowContext(ctx, "SELECT unrestricted FROM repo_permissions WHERE repo_id = 1").Scan(&unrestricted)
	require.NoError(t, err)
	require.False(t, unrestricted)
}

func TestHandleRestricted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)

	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	// create an external service
	err := db.ExternalServices().Create(ctx, confGet, &types.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.sgdev.org", "username": "username", "token": "qwerty", "projectKeys": ["SGDEMO"]}`),
	})
	require.NoError(t, err)

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

	// create 6 repos
	_, err = db.ExecContext(ctx, `--sql
	INSERT INTO repo (id, external_id, external_service_type, external_service_id, name, fork, private)
	VALUES
		(1, 10060, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/go', false, true),
		(2, 10056, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/jenkins', false, true),
		(3, 10061, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/mux', false, true),
		(4, 10058, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sentry', false, true),
		(5, 10059, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sinatra', false, true),
		(6, 10072, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sourcegraph', false, true);

	INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
	VALUES
		(1, 1, 'bitbucket.sgdev.org/SGDEMO/go'),
		(1, 2, 'bitbucket.sgdev.org/SGDEMO/jenkins'),
		(1, 3, 'bitbucket.sgdev.org/SGDEMO/mux'),
		(1, 4, 'bitbucket.sgdev.org/SGDEMO/sentry'),
		(1, 5, 'bitbucket.sgdev.org/SGDEMO/sinatra'),
		(1, 6, 'bitbucket.sgdev.org/SGDEMO/sourcegraph');
`)
	require.NoError(t, err)

	h := bitbucketProjectPermissionsHandler{
		db:     db,
		client: bitbucketserver.NewTestClient(t, "client", false),
	}

	// set permissions for 3 users (2 existing, 1 pending) and 2 repos
	err = h.Handle(ctx, logtest.Scoped(t), &types.BitbucketProjectPermissionJob{
		ExternalServiceID: 1,
		ProjectKey:        "SGDEMO",
		Permissions: []types.UserPermission{
			{BindID: "pushpa@example.com", Permission: "read"},
			{BindID: "igor@example.com", Permission: "read"},
			{BindID: "sayako", Permission: "read"},
		},
	})
	require.NoError(t, err)

	// check that the permissions were set
	perms := db.Perms()

	for _, repoID := range []int32{1, 2, 3, 4, 5, 6} {
		p, err := perms.LoadRepoPermissions(ctx, repoID)
		require.NoError(t, err)
		gotIDs := make([]int32, len(p))
		for i, perm := range p {
			gotIDs[i] = perm.UserID
		}
		slices.Sort(gotIDs)

		require.Equal(t, []int32{igor.ID, pushpa.ID}, gotIDs)
	}

	up, err := perms.LoadUserPermissions(ctx, pushpa.ID)
	require.NoError(t, err)
	gotIDs := make([]int32, len(up))
	for i, perm := range up {
		gotIDs[i] = perm.RepoID
	}
	slices.Sort(gotIDs)

	require.Equal(t, []int32{1, 2, 3, 4, 5, 6}, gotIDs)
}

func TestHandleUnrestricted(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	t.Parallel()

	logger := logtest.Scoped(t)
	ctx := context.Background()

	db := database.NewDB(logger, dbtest.NewDB(t))

	confGet := func() *conf.Unified {
		return &conf.Unified{}
	}
	// create an external service
	err := db.ExternalServices().Create(ctx, confGet, &types.ExternalService{
		Kind:        extsvc.KindBitbucketServer,
		DisplayName: "Bitbucket #1",
		Config:      extsvc.NewUnencryptedConfig(`{"url": "https://bitbucket.sgdev.org", "username": "username", "token": "qwerty", "projectKeys": ["SGDEMO"]}`),
	})
	require.NoError(t, err)

	// create 3 users
	users := db.Users()
	_, err = users.Create(ctx,
		database.NewUser{
			Email:           "igor@example.com",
			Username:        "igor",
			EmailIsVerified: true,
		},
	)
	require.NoError(t, err)
	_, err = users.Create(ctx,
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

	// create 6 repos
	_, err = db.ExecContext(ctx, `--sql
	INSERT INTO repo (id, external_id, external_service_type, external_service_id, name, fork, private)
	VALUES
		(1, 10060, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/go', false, true),
		(2, 10056, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/jenkins', false, true),
		(3, 10061, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/mux', false, true),
		(4, 10058, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sentry', false, true),
		(5, 10059, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sinatra', false, true),
		(6, 10072, 'bitbucketServer', 'https://bitbucket.sgdev.org/', 'bitbucket.sgdev.org/SGDEMO/sourcegraph', false, true);

	INSERT INTO external_service_repos (external_service_id, repo_id, clone_url)
	VALUES
		(1, 1, 'bitbucket.sgdev.org/SGDEMO/go'),
		(1, 2, 'bitbucket.sgdev.org/SGDEMO/jenkins'),
		(1, 3, 'bitbucket.sgdev.org/SGDEMO/mux'),
		(1, 4, 'bitbucket.sgdev.org/SGDEMO/sentry'),
		(1, 5, 'bitbucket.sgdev.org/SGDEMO/sinatra'),
		(1, 6, 'bitbucket.sgdev.org/SGDEMO/sourcegraph');

	INSERT INTO repo_permissions (repo_id, permission, updated_at)
	VALUES
		(1, 'read', now()),
		(2, 'read', now()),
		(3, 'read', now()),
		(4, 'read', now()),
		(5, 'read', now()),
		(6, 'read', now());
`)
	require.NoError(t, err)

	h := bitbucketProjectPermissionsHandler{
		db:     db,
		client: bitbucketserver.NewTestClient(t, "client", false),
	}

	// set permissions for 3 users (2 existing, 1 pending) and 2 repos
	err = h.Handle(ctx, logtest.Scoped(t), &types.BitbucketProjectPermissionJob{
		ExternalServiceID: 1,
		ProjectKey:        "SGDEMO",
		Unrestricted:      true,
	})
	require.NoError(t, err)

	// check that the permissions were set
	perms := db.Perms()

	for _, repoID := range []int32{1, 2, 3, 4, 5, 6} {
		p, err := perms.LoadRepoPermissions(ctx, repoID)
		require.NoError(t, err)
		// if there's only 1 item and userID is 0, it means that the repo is unrestricted
		require.Equal(t, 1, len(p))
		require.Equal(t, int32(0), p[0].UserID)
	}
}
