package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/log/logtest"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestTemporarySettingsStore(t *testing.T) {
	t.Parallel()
	t.Run("GetEmpty", testGetEmpty)
	t.Run("InsertAndGet", testInsertAndGet)
	t.Run("UpdateAndGet", testUpdateAndGet)
	t.Run("InsertWithInvalidData", testInsertWithInvalidData)
	t.Run("TestEdit", testEdit)
}

func testGetEmpty(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	temporarySettingsStore := NewDB(logger, dbtest.NewDB(t)).TemporarySettings()

	ctx := actor.WithInternalActor(context.Background())

	expected := ts.TemporarySettings{Contents: "{}"}

	res, err := temporarySettingsStore.GetTemporarySettings(ctx, 1)
	require.NoError(t, err)

	require.Equal(t, res, &expected)
}

func testInsertAndGet(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	usersStore := db.Users()
	temporarySettingsStore := db.TemporarySettings()

	ctx := actor.WithInternalActor(context.Background())

	contents := "{\"search.collapsedSidebarSections\": {}}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.OverwriteTemporarySettings(ctx, user.ID, contents)
	require.NoError(t, err)

	res, err := temporarySettingsStore.GetTemporarySettings(ctx, user.ID)
	require.NoError(t, err)

	expected := ts.TemporarySettings{Contents: contents}
	require.Equal(t, res, &expected)
}

func testUpdateAndGet(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	usersStore := db.Users()
	temporarySettingsStore := db.TemporarySettings()

	ctx := actor.WithInternalActor(context.Background())

	contents := "{\"search.collapsedSidebarSections\": {}}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.OverwriteTemporarySettings(ctx, user.ID, contents)
	require.NoError(t, err)

	contents2 := "{\"search.collapsedSidebarSections\": {\"types\": false}}"

	err = temporarySettingsStore.OverwriteTemporarySettings(ctx, user.ID, contents2)
	require.NoError(t, err)

	res, err := temporarySettingsStore.GetTemporarySettings(ctx, user.ID)
	require.NoError(t, err)

	expected := ts.TemporarySettings{Contents: contents2}
	require.Equal(t, res, &expected)
}

func testInsertWithInvalidData(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	usersStore := db.Users()
	temporarySettingsStore := db.TemporarySettings()

	ctx := actor.WithInternalActor(context.Background())

	contents := "{\"search.collapsedSidebarSections\": {}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.OverwriteTemporarySettings(ctx, user.ID, contents)
	require.EqualError(t, errors.Unwrap(errors.Unwrap(err)), "ERROR: invalid input syntax for type json (SQLSTATE 22P02)")
}

func testEdit(t *testing.T) {
	t.Parallel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(t))
	usersStore := db.Users()
	temporarySettingsStore := db.TemporarySettings()

	ctx := actor.WithInternalActor(context.Background())

	edit1 := "{\"search.collapsedSidebarSections\": {}, \"search.onboarding.tourCancelled\": true}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.EditTemporarySettings(ctx, user.ID, edit1)
	require.NoError(t, err)

	res1, err := temporarySettingsStore.GetTemporarySettings(ctx, user.ID)
	require.NoError(t, err)

	expected1 := ts.TemporarySettings{Contents: edit1}
	require.Equal(t, res1, &expected1)

	edit2 := "{\"search.collapsedSidebarSections\": {\"types\": false}}"

	err = temporarySettingsStore.EditTemporarySettings(ctx, user.ID, edit2)
	require.NoError(t, err)

	res2, err := temporarySettingsStore.GetTemporarySettings(ctx, user.ID)
	require.NoError(t, err)

	expected2 := ts.TemporarySettings{Contents: "{\"search.collapsedSidebarSections\": {\"types\": false}, \"search.onboarding.tourCancelled\": true}"}
	require.Equal(t, res2, &expected2)
}
