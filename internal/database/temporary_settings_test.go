package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtest"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
)

func TestTemporarySettingsStore(t *testing.T) {
	t.Parallel()
	t.Run("GetEmpty", testGetEmpty)
	t.Run("InsertAndGet", testInsertAndGet)
	t.Run("UpdateAndGet", testUpdateAndGet)
	t.Run("InsertWithInvalidData", testInsertWithInvalidData)
}

func testGetEmpty(t *testing.T) {
	t.Parallel()
	temporarySettingsStore := TemporarySettings(dbtest.NewDB(t, ""))

	ctx := actor.WithInternalActor(context.Background())

	expected := ts.TemporarySettings{Contents: "{}"}

	res, err := temporarySettingsStore.GetTemporarySettings(ctx, 1)
	require.NoError(t, err)

	require.Equal(t, res, &expected)
}

func testInsertAndGet(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	usersStore := Users(db)
	temporarySettingsStore := TemporarySettings(db)

	ctx := actor.WithInternalActor(context.Background())

	contents := "{\"search.collapsedSidebarSections\": {}}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.UpsertTemporarySettings(ctx, user.ID, contents)
	require.NoError(t, err)

	res, err := temporarySettingsStore.GetTemporarySettings(ctx, user.ID)
	require.NoError(t, err)

	expected := ts.TemporarySettings{Contents: contents}
	require.Equal(t, res, &expected)
}

func testUpdateAndGet(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	usersStore := Users(db)
	temporarySettingsStore := TemporarySettings(db)

	ctx := actor.WithInternalActor(context.Background())

	contents := "{\"search.collapsedSidebarSections\": {}}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.UpsertTemporarySettings(ctx, user.ID, contents)
	require.NoError(t, err)

	contents2 := "{\"search.collapsedSidebarSections\": {\"types\": false}}"

	err = temporarySettingsStore.UpsertTemporarySettings(ctx, user.ID, contents2)
	require.NoError(t, err)

	res, err := temporarySettingsStore.GetTemporarySettings(ctx, user.ID)
	require.NoError(t, err)

	expected := ts.TemporarySettings{Contents: contents2}
	require.Equal(t, res, &expected)
}

func testInsertWithInvalidData(t *testing.T) {
	t.Parallel()
	db := dbtest.NewDB(t, "")
	usersStore := Users(db)
	temporarySettingsStore := TemporarySettings(db)

	ctx := actor.WithInternalActor(context.Background())

	contents := "{\"search.collapsedSidebarSections\": {}"

	user, err := usersStore.Create(ctx, NewUser{Username: "u", Password: "p"})
	require.NoError(t, err)

	err = temporarySettingsStore.UpsertTemporarySettings(ctx, user.ID, contents)
	require.EqualError(t, err, "ERROR: invalid input syntax for type json (SQLSTATE 22P02)")
}
