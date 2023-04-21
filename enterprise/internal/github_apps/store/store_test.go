package store

import (
	"context"
	"testing"

	"github.com/keegancsmith/sqlf"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/github_apps/types"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
)

func TestCreateGitHubApp(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)
	db := database.NewDB(logger, dbtest.NewDB(logger, t))
	store := &gitHubAppsStore{Store: basestore.NewWithHandle(db.Handle())}
	ctx := context.Background()

	app := &types.GitHubApp{
		AppID:        1,
		Name:         "Test App",
		Slug:         "test-app",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	err := store.Create(ctx, app)
	require.NoError(t, err)

	var createdApp types.GitHubApp
	query := sqlf.Sprintf(`SELECT app_id, name, slug, base_url, client_id, client_secret, private_key, encryption_key_id, logo FROM github_apps WHERE app_id=%s`, app.AppID)
	err = store.QueryRow(ctx, query).Scan(
		&createdApp.AppID,
		&createdApp.Name,
		&createdApp.Slug,
		&createdApp.BaseURL,
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

	app := &types.GitHubApp{
		ID:           1,
		Name:         "Test App",
		Slug:         "test-app",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
	}

	err := store.Create(ctx, app)
	require.NoError(t, err)

	err = store.Delete(ctx, app.ID)
	require.NoError(t, err)

	query := sqlf.Sprintf(`SELECT * FROM github_apps WHERE id=%s`, app.ID)
	row, err := store.Query(ctx, query)
	require.NoError(t, err)
	// expect false since the query should not return any results
	require.False(t, row.Next())

	// deleting non-existent should not return error
	err = store.Delete(ctx, app.ID)
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

	app := &types.GitHubApp{
		AppID:        123,
		Name:         "Test App",
		Slug:         "test-app",
		BaseURL:      "https://example.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
	}

	err := store.Create(ctx, app)
	require.NoError(t, err)

	app, err = store.GetByID(ctx, 1)
	require.NoError(t, err)

	updated := &types.GitHubApp{
		AppID:        234,
		Name:         "Updated Name",
		Slug:         "updated-slug",
		BaseURL:      "https://updated-example.com",
		ClientID:     "def456",
		ClientSecret: "updated-secret",
		PrivateKey:   "updated-private-key",
	}

	fetched, err := store.Update(ctx, 1, updated)
	require.NoError(t, err)

	require.Greater(t, fetched.UpdatedAt, app.UpdatedAt)

	require.Equal(t, updated.AppID, fetched.AppID)
	require.Equal(t, updated.Name, fetched.Name)
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

	app1 := &types.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Slug:         "test-app-1",
		BaseURL:      "https://github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	app2 := &types.GitHubApp{
		AppID:        5678,
		Name:         "Test App 2",
		Slug:         "test-app-2",
		BaseURL:      "https://enterprise.github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	err := store.Create(ctx, app1)
	require.NoError(t, err)
	err = store.Create(ctx, app2)
	require.NoError(t, err)

	fetched, err := store.GetByID(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, app1.AppID, fetched.AppID)
	require.Equal(t, app1.Name, fetched.Name)
	require.Equal(t, app1.Slug, fetched.Slug)
	require.Equal(t, app1.BaseURL, fetched.BaseURL)
	require.Equal(t, app1.ClientID, fetched.ClientID)
	require.Equal(t, app1.ClientSecret, fetched.ClientSecret)
	require.Equal(t, app1.PrivateKey, fetched.PrivateKey)
	require.Equal(t, app1.Logo, fetched.Logo)
	require.NotZero(t, fetched.CreatedAt)
	require.NotZero(t, fetched.UpdatedAt)

	fetched, err = store.GetByID(ctx, 2)
	require.NoError(t, err)
	require.Equal(t, app2.AppID, fetched.AppID)

	// does not exist
	_, err = store.GetByID(ctx, 3)
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

	app1 := &types.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Slug:         "test-app-1",
		BaseURL:      "https://github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	app2 := &types.GitHubApp{
		AppID:        1234,
		Name:         "Test App 2",
		Slug:         "test-app-2",
		BaseURL:      "https://enterprise.github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	err := store.Create(ctx, app1)
	require.NoError(t, err)
	err = store.Create(ctx, app2)
	require.NoError(t, err)

	fetched, err := store.GetByAppID(ctx, 1234, "https://github.com")
	require.NoError(t, err)
	require.Equal(t, app1.AppID, fetched.AppID)
	require.Equal(t, app1.Name, fetched.Name)
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

	app1 := &types.GitHubApp{
		AppID:        1234,
		Name:         "Test App 1",
		Slug:         "test-app",
		BaseURL:      "https://github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	app2 := &types.GitHubApp{
		AppID:        5678,
		Name:         "Test App",
		Slug:         "test-app",
		BaseURL:      "https://enterprise.github.com",
		ClientID:     "abc123",
		ClientSecret: "secret",
		PrivateKey:   "private-key",
		Logo:         "logo.png",
	}

	err := store.Create(ctx, app1)
	require.NoError(t, err)
	err = store.Create(ctx, app2)
	require.NoError(t, err)

	fetched, err := store.GetBySlug(ctx, "test-app", "https://github.com")
	require.NoError(t, err)
	require.Equal(t, app1.AppID, fetched.AppID)
	require.Equal(t, app1.Name, fetched.Name)
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
