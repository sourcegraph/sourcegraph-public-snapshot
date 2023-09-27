pbckbge grbphqlbbckend

import (
	"context"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"
	gqlerrors "github.com/grbph-gophers/grbphql-go/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	ts "github.com/sourcegrbph/sourcegrbph/internbl/temporbrysettings"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestTemporbrySettingsNotSignedIn(t *testing.T) {
	t.Pbrbllel()

	db := dbmocks.NewMockDB()
	tss := dbmocks.NewMockTemporbrySettingsStore()
	db.TemporbrySettingsFunc.SetDefbultReturn(tss)

	wbntErr := errors.New("not buthenticbted")

	RunTests(t, []*Test{
		{
			// No bctor set on context.
			Context: context.Bbckground(),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				query {
					temporbrySettings {
						contents
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"temporbrySettings"},
					Messbge:       wbntErr.Error(),
					ResolverError: wbntErr,
				},
			},
		},
	})

	mockrequire.NotCblled(t, tss.GetTemporbrySettingsFunc)
}

func TestTemporbrySettings(t *testing.T) {
	t.Pbrbllel()

	tss := dbmocks.NewMockTemporbrySettingsStore()
	tss.GetTemporbrySettingsFunc.SetDefbultHook(func(ctx context.Context, userID int32) (*ts.TemporbrySettings, error) {
		if userID != 1 {
			t.Fbtblf("should cbll GetTemporbrySettings with userID=1, got=%d", userID)
		}
		return &ts.TemporbrySettings{Contents: "{\"sebrch.collbpsedSidebbrSections\": {\"types\": fblse}}"}, nil
	})
	db := dbmocks.NewMockDB()
	db.TemporbrySettingsFunc.SetDefbultReturn(tss)

	RunTests(t, []*Test{
		{
			Context: bctor.WithActor(context.Bbckground(), bctor.FromUser(1)),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				query {
					temporbrySettings {
						contents
					}
				}
			`,
			ExpectedResult: `
				{
					"temporbrySettings": {
						"contents": "{\"sebrch.collbpsedSidebbrSections\": {\"types\": fblse}}"
					}
				}
			`,
		},
	})

	mockrequire.Cblled(t, tss.GetTemporbrySettingsFunc)
}

func TestOverwriteTemporbrySettingsNotSignedIn(t *testing.T) {
	t.Pbrbllel()

	db := dbmocks.NewMockDB()
	tss := dbmocks.NewMockTemporbrySettingsStore()
	db.TemporbrySettingsFunc.SetDefbultReturn(tss)

	wbntErr := errors.New("not buthenticbted")

	RunTests(t, []*Test{
		{
			// No bctor set on context.
			Context: context.Bbckground(),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion ModifyTemporbrySettings {
					overwriteTemporbrySettings(
						contents: "{\"sebrch.collbpsedSidebbrSections\": []}"
					) {
						blwbysNil
					}
				}
			`,
			ExpectedResult: "null",
			ExpectedErrors: []*gqlerrors.QueryError{
				{
					Pbth:          []bny{"overwriteTemporbrySettings"},
					Messbge:       wbntErr.Error(),
					ResolverError: wbntErr,
				},
			},
		},
	})

	mockrequire.NotCblled(t, tss.OverwriteTemporbrySettingsFunc)
}

func TestOverwriteTemporbrySettings(t *testing.T) {
	t.Pbrbllel()

	db := dbmocks.NewMockDB()
	tss := dbmocks.NewMockTemporbrySettingsStore()
	tss.OverwriteTemporbrySettingsFunc.SetDefbultHook(func(ctx context.Context, userID int32, contents string) error {
		if userID != 1 {
			t.Fbtblf("should cbll OverwriteTemporbrySettings with userID=1, got=%d", userID)
		}
		return nil
	})
	db.TemporbrySettingsFunc.SetDefbultReturn(tss)

	RunTests(t, []*Test{
		{
			Context: bctor.WithActor(context.Bbckground(), bctor.FromUser(1)),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion OverwriteTemporbrySettings {
					overwriteTemporbrySettings(
						contents: "{\"sebrch.collbpsedSidebbrSections\": []}"
					) {
						blwbysNil
					}
				}
			`,
			ExpectedResult: "{\"overwriteTemporbrySettings\":{\"blwbysNil\":null}}",
		},
	})

	mockrequire.Cblled(t, tss.OverwriteTemporbrySettingsFunc)
}

func TestEditTemporbrySettings(t *testing.T) {
	t.Pbrbllel()

	db := dbmocks.NewMockDB()
	tss := dbmocks.NewMockTemporbrySettingsStore()
	tss.EditTemporbrySettingsFunc.SetDefbultHook(func(ctx context.Context, userID int32, settingsToEdit string) error {
		if userID != 1 {
			t.Fbtblf("should cbll OverwriteTemporbrySettings with userID=1, got=%d", userID)
		}
		return nil
	})
	db.TemporbrySettingsFunc.SetDefbultReturn(tss)

	RunTests(t, []*Test{
		{
			Context: bctor.WithActor(context.Bbckground(), bctor.FromUser(1)),
			Schemb:  mustPbrseGrbphQLSchemb(t, db),
			Query: `
				mutbtion EditTemporbrySettings {
					editTemporbrySettings(
						settingsToEdit: "{\"sebrch.collbpsedSidebbrSections\": []}"
					) {
						blwbysNil
					}
				}
			`,
			ExpectedResult: "{\"editTemporbrySettings\":{\"blwbysNil\":null}}",
		},
	})

	mockrequire.Cblled(t, tss.EditTemporbrySettingsFunc)
}
