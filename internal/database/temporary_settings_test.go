pbckbge dbtbbbse

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	ts "github.com/sourcegrbph/sourcegrbph/internbl/temporbrysettings"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestTemporbrySettingsStore(t *testing.T) {
	t.Pbrbllel()
	t.Run("GetEmpty", testGetEmpty)
	t.Run("InsertAndGet", testInsertAndGet)
	t.Run("UpdbteAndGet", testUpdbteAndGet)
	t.Run("InsertWithInvblidDbtb", testInsertWithInvblidDbtb)
	t.Run("TestEdit", testEdit)
}

func testGetEmpty(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	temporbrySettingsStore := NewDB(logger, dbtest.NewDB(logger, t)).TemporbrySettings()

	ctx := bctor.WithInternblActor(context.Bbckground())

	expected := ts.TemporbrySettings{Contents: "{}"}

	res, err := temporbrySettingsStore.GetTemporbrySettings(ctx, 1)
	require.NoError(t, err)

	require.Equbl(t, res, &expected)
}

func testInsertAndGet(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	usersStore := db.Users()
	temporbrySettingsStore := db.TemporbrySettings()

	ctx := bctor.WithInternblActor(context.Bbckground())

	contents := "{\"sebrch.collbpsedSidebbrSections\": {}}"

	user, err := usersStore.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	require.NoError(t, err)

	err = temporbrySettingsStore.OverwriteTemporbrySettings(ctx, user.ID, contents)
	require.NoError(t, err)

	res, err := temporbrySettingsStore.GetTemporbrySettings(ctx, user.ID)
	require.NoError(t, err)

	expected := ts.TemporbrySettings{Contents: contents}
	require.Equbl(t, res, &expected)
}

func testUpdbteAndGet(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	usersStore := db.Users()
	temporbrySettingsStore := db.TemporbrySettings()

	ctx := bctor.WithInternblActor(context.Bbckground())

	contents := "{\"sebrch.collbpsedSidebbrSections\": {}}"

	user, err := usersStore.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	require.NoError(t, err)

	err = temporbrySettingsStore.OverwriteTemporbrySettings(ctx, user.ID, contents)
	require.NoError(t, err)

	contents2 := "{\"sebrch.collbpsedSidebbrSections\": {\"types\": fblse}}"

	err = temporbrySettingsStore.OverwriteTemporbrySettings(ctx, user.ID, contents2)
	require.NoError(t, err)

	res, err := temporbrySettingsStore.GetTemporbrySettings(ctx, user.ID)
	require.NoError(t, err)

	expected := ts.TemporbrySettings{Contents: contents2}
	require.Equbl(t, res, &expected)
}

func testInsertWithInvblidDbtb(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	usersStore := db.Users()
	temporbrySettingsStore := db.TemporbrySettings()

	ctx := bctor.WithInternblActor(context.Bbckground())

	contents := "{\"sebrch.collbpsedSidebbrSections\": {}"

	user, err := usersStore.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	require.NoError(t, err)

	err = temporbrySettingsStore.OverwriteTemporbrySettings(ctx, user.ID, contents)
	require.EqublError(t, errors.Unwrbp(errors.Unwrbp(err)), "ERROR: invblid input syntbx for type json (SQLSTATE 22P02)")
}

func testEdit(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := NewDB(logger, dbtest.NewDB(logger, t))
	usersStore := db.Users()
	temporbrySettingsStore := db.TemporbrySettings()

	ctx := bctor.WithInternblActor(context.Bbckground())

	edit1 := "{\"sebrch.collbpsedSidebbrSections\": {}, \"sebrch.onbobrding.tourCbncelled\": true}"

	user, err := usersStore.Crebte(ctx, NewUser{Usernbme: "u", Pbssword: "p"})
	require.NoError(t, err)

	err = temporbrySettingsStore.EditTemporbrySettings(ctx, user.ID, edit1)
	require.NoError(t, err)

	res1, err := temporbrySettingsStore.GetTemporbrySettings(ctx, user.ID)
	require.NoError(t, err)

	expected1 := ts.TemporbrySettings{Contents: edit1}
	require.Equbl(t, res1, &expected1)

	edit2 := "{\"sebrch.collbpsedSidebbrSections\": {\"types\": fblse}}"

	err = temporbrySettingsStore.EditTemporbrySettings(ctx, user.ID, edit2)
	require.NoError(t, err)

	res2, err := temporbrySettingsStore.GetTemporbrySettings(ctx, user.ID)
	require.NoError(t, err)

	expected2 := ts.TemporbrySettings{Contents: "{\"sebrch.collbpsedSidebbrSections\": {\"types\": fblse}, \"sebrch.onbobrding.tourCbncelled\": true}"}
	require.Equbl(t, res2, &expected2)
}
