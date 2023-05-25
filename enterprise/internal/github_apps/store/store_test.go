package store

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	ghtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestCreateGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := &gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app := &ghtypes.GitHubApp{
		AppID:        1,
		Name:         "Test App",
		Domain:       "repos",
		Slug:         "test-app",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
		AppURL:       "https://github.com/apps/testapp",
	}

	id, err := store.Create(ctx, app)
	require.NoError(t, err)

	var createdApp ghtypes.GitHubApp
	query := sqlf.Sprintf(`SELECT app_id, name, domain, slug, base_url, app_url, client_id, client_secret, private_key, encryption_key_id, logo FROM github_apps WHERE id=%s`, id)
	err = store.QueryRow(ctx, query).Scan(
		&createdApp.AppID,
		&createdApp.Name,
		&createdApp.Domain,
		&createdApp.Slug,
		&createdApp.BaseURL,
		&createdApp.AppURL,
		&createdApp.ClientID,
		&createdApp.ClientSecret,
		&createdApp.PrivateKey,
		&createdApp.EncryptionKey,
		&createdApp.Logo,
	)
	require.NoError(t, err)
	require.Equal(t, app, &createdApp)
}

func TestDeleteGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app := &ghtypes.GitHubApp{
		Name:         "Test App",
		Domain:       "repos",
		Slug:         "test-app",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
	}

	id, err := store.Create(ctx, app)
	require.NoError(t, err)

	err = store.Delete(ctx, id)
	require.NoError(t, err)

	query := sqlf.Sprintf(`SELECT * FROM github_apps WHERE id=%s`, id)
	row, err := store.Query(ctx, query)
	require.NoError(t, err)
	// expect false since the query should not return any results
	require.False(t, row.Next())

	// deleting non-existent should not return error
	err = store.Delete(ctx, id)
	require.NoError(t, err)
}

func TestUpdateGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app := &ghtypes.GitHubApp{
		AppID:        123,
		Name:         "Test App",
		Domain:       "repos",
		Slug:         "test-app",
		BaseURL:      "https://example.com",
		AppURL:       "https://example.com/apps/testapp",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
	}

	id, err := store.Create(ctx, app)
	require.NoError(t, err)

	app, err = store.GetByID(ctx, id)
	require.NoError(t, err)

	updated := &ghtypes.GitHubApp{
		AppID:        234,
		Name:         "Updated Name",
		Domain:       "repos",
		Slug:         "updated-slug",
		BaseURL:      "https://updated-example.com",
		AppURL:       "https://updated-example.com/apps/updated-app",
		ClientID:     "def456",
		ClientSecret: "updated-secret",
		PrivateKey:   "updated-private-key",
	}

	fetched, err := store.Update(ctx, 1, updated)
	require.NoError(t, err)

	require.Greater(t, fetched.UpdatedAt, app.UpdatedAt)

	require.Equal(t, updated.AppID, fetched.AppID)
	require.Equal(t, updated.Name, fetched.Name)
	require.Equal(t, updated.Domain, fetched.Domain)
	require.Equal(t, updated.Slug, fetched.Slug)
	require.Equal(t, updated.BaseURL, fetched.BaseURL)
	require.Equal(t, updated.ClientID, fetched.ClientID)
	require.Equal(t, updated.ClientSecret, fetched.ClientSecret)
	require.Equal(t, updated.PrivateKey, fetched.PrivateKey)
	require.Equal(t, updated.Logo, fetched.Logo)

	// updating non-existent should result in error
	_, err = store.Update(ctx, 42, updated)
	require.Error(t, err)
}

func TestGetByID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := &gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app1 := &ghtypes.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Domain:       "repos",
		Slug:         "test-app-1",
		BaseURL:      "https://github.com",
		AppURL:       "https://github.com/apps/testapp",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	app2 := &ghtypes.GitHubApp{
		AppID:        5678,
		Name:         "Test App 2",
		Domain:       "repos",
		Slug:         "test-app-2",
		BaseURL:      "https://enterprise.github.com",
		AppURL:       "https://enterprise.github.com/apps/testapp",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	id1, err := store.Create(ctx, app1)
	require.NoError(t, err)
	id2, err := store.Create(ctx, app2)
	require.NoError(t, err)

	fetched, err := store.GetByID(ctx, id1)
	require.NoError(t, err)
	require.Equal(t, app1.AppID, fetched.AppID)
	require.Equal(t, app1.Name, fetched.Name)
	require.Equal(t, app1.Domain, fetched.Domain)
	require.Equal(t, app1.Slug, fetched.Slug)
	require.Equal(t, app1.BaseURL, fetched.BaseURL)
	require.Equal(t, app1.ClientID, fetched.ClientID)
	require.Equal(t, app1.ClientSecret, fetched.ClientSecret)
	require.Equal(t, app1.PrivateKey, fetched.PrivateKey)
	require.Equal(t, app1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CreatedAt)
	require.NotZero(t, fetched.UpdatedAt)

	fetched, err = store.GetByID(ctx, id2)
	require.NoError(t, err)
	require.Equal(t, app2.AppID, fetched.AppID)

	// does not exist
	_, err = store.GetByID(ctx, 42)
	require.Error(t, err)
}

func TestGetByAppID(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := &gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app1 := &ghtypes.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Domain:       "repos",
		Slug:         "test-app-1",
		BaseURL:      "https://github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	app2 := &ghtypes.GitHubApp{
		AppID:        1234,
		Name:         "Test App 2",
		Domain:       "repos",
		Slug:         "test-app-2",
		BaseURL:      "https://enterprise.github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	_, err := store.Create(ctx, app1)
	require.NoError(t, err)
	_, err = store.Create(ctx, app2)
	require.NoError(t, err)

	fetched, err := store.GetByAppID(ctx, 1234, "https://github.com")
	require.NoError(t, err)
	require.Equal(t, app1.AppID, fetched.AppID)
	require.Equal(t, app1.Name, fetched.Name)
	require.Equal(t, app1.Domain, fetched.Domain)
	require.Equal(t, app1.Slug, fetched.Slug)
	require.Equal(t, app1.BaseURL, fetched.BaseURL)
	require.Equal(t, app1.ClientID, fetched.ClientID)
	require.Equal(t, app1.ClientSecret, fetched.ClientSecret)
	require.Equal(t, app1.PrivateKey, fetched.PrivateKey)
	require.Equal(t, app1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CreatedAt)
	require.NotZero(t, fetched.UpdatedAt)

	fetched, err = store.GetByAppID(ctx, 1234, "https://enterprise.github.com")
	require.NoError(t, err)
	require.Equal(t, app2.AppID, fetched.AppID)
	require.Equal(t, app2.Slug, fetched.Slug)

	// does not exist
	_, err = store.GetByAppID(ctx, 3456, "https://github.com")
	require.Error(t, err)
}

func TestGetBySlug(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := &gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app1 := &ghtypes.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Domain:       "repos",
		Slug:         "test-app",
		BaseURL:      "https://github.com",
		AppURL:       "https://github.com/apps/testapp1",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	app2 := &ghtypes.GitHubApp{
		AppID:        5678,
		Name:         "Test App",
		Domain:       "repos",
		Slug:         "test-app",
		BaseURL:      "https://enterprise.github.com",
		AppURL:       "https://enterprise.github.com/apps/testapp",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	_, err := store.Create(ctx, app1)
	require.NoError(t, err)
	_, err = store.Create(ctx, app2)
	require.NoError(t, err)

	fetched, err := store.GetBySlug(ctx, "test-app", "https://github.com")
	require.NoError(t, err)
	require.Equal(t, app1.AppID, fetched.AppID)
	require.Equal(t, app1.Name, fetched.Name)
	require.Equal(t, app1.Domain, fetched.Domain)
	require.Equal(t, app1.Slug, fetched.Slug)
	require.Equal(t, app1.BaseURL, fetched.BaseURL)
	require.Equal(t, app1.ClientID, fetched.ClientID)
	require.Equal(t, app1.ClientSecret, fetched.ClientSecret)
	require.Equal(t, app1.PrivateKey, fetched.PrivateKey)
	require.Equal(t, app1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CreatedAt)
	require.NotZero(t, fetched.UpdatedAt)

	fetched, err = store.GetBySlug(ctx, "test-app", "https://enterprise.github.com")
	require.NoError(t, err)
	require.Equal(t, app2.AppID, fetched.AppID)

	// does not exist
	_, err = store.GetBySlug(ctx, "foo", "bar")
	require.Error(t, err)
}

func TestListGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := &gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	repoApp := &ghtypes.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Domain:       types.ReposDomain,
		Slug:         "test-app-1",
		BaseURL:      "https://github.com",
		AppURL:       "https://github.com/apps/testapp",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	batchesApp := &ghtypes.GitHubApp{
		AppID:        5678,
		Name:         "Test App 2",
		Domain:       types.BatchesDomain,
		Slug:         "test-app-2",
		BaseURL:      "https://enterprise.github.com",
		AppURL:       "https://enterprise.github.com/apps/testapp",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	_, err := store.Create(ctx, repoApp)
	require.NoError(t, err)
	_, err = store.Create(ctx, batchesApp)
	require.NoError(t, err)

	t.Run("all github apps", func(t *testing.T) {
		fetched, err := store.List(ctx, nil)
		require.NoError(t, err)
		require.Len(t, fetched, 2)

		apps := []*ghtypes.GitHubApp{app1, app2}
		for index, curr := range fetched {
			app := apps[index]
			require.Equal(t, app.AppID, curr.AppID)
			require.Equal(t, app.Name, curr.Name)
			require.Equal(t, app.Domain, curr.Domain)
			require.Equal(t, app.Slug, curr.Slug)
			require.Equal(t, app.BaseURL, curr.BaseURL)
			require.Equal(t, app.ClientID, curr.ClientID)
			require.Equal(t, app.ClientSecret, curr.ClientSecret)
			require.Equal(t, app.PrivateKey, curr.PrivateKey)
			require.Equal(t, app.Logo, curr.Logo)
			require.NotZero(t, curr.CreatedAt)
			require.NotZero(t, curr.UpdatedAt)
		}
	})

	t.Run("domain-filtered github apps", func(t *testing.T) {
		domain := types.ReposDomain
		fetched, err := store.List(ctx, &domain)
		require.NoError(t, err)
		require.Len(t, fetched, 1)

		curr := fetched[0]
		require.Equal(t, curr.AppID, repoApp.AppID)
		require.Equal(t, curr.Name, repoApp.Name)
		require.Equal(t, curr.Domain, repoApp.Domain)
		require.Equal(t, curr.Slug, repoApp.Slug)
		require.Equal(t, curr.BaseURL, repoApp.BaseURL)
		require.Equal(t, curr.ClientID, repoApp.ClientID)
		require.Equal(t, curr.ClientSecret, repoApp.ClientSecret)
		require.Equal(t, curr.PrivateKey, repoApp.PrivateKey)
		require.Equal(t, curr.Logo, repoApp.Logo)
		require.NotZero(t, curr.CreatedAt)
		require.NotZero(t, curr.UpdatedAt)
	})
}
