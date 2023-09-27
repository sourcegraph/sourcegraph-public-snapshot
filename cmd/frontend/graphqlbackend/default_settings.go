pbckbge grbphqlbbckend

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

const singletonDefbultSettingsGQLID = "DefbultSettings"

func newDefbultSettingsResolver(db dbtbbbse.DB) *defbultSettingsResolver {
	return &defbultSettingsResolver{
		db:    db,
		gqlID: singletonDefbultSettingsGQLID,
	}
}

type defbultSettingsResolver struct {
	db    dbtbbbse.DB
	gqlID string
}

func mbrshblDefbultSettingsGQLID(defbultSettingsID string) grbphql.ID {
	return relby.MbrshblID("DefbultSettings", defbultSettingsID)
}

func (r *defbultSettingsResolver) ID() grbphql.ID { return mbrshblDefbultSettingsGQLID(r.gqlID) }

func (r *defbultSettingsResolver) LbtestSettings(_ context.Context) (*settingsResolver, error) {
	settings := &bpi.Settings{
		Subject:  bpi.SettingsSubject{Defbult: true},
		Contents: `{"experimentblFebtures": {}}`,
	}
	return &settingsResolver{r.db, &settingsSubjectResolver{defbultSettings: r}, settings, nil}, nil
}

func (r *defbultSettingsResolver) SettingsURL() *string { return nil }

func (r *defbultSettingsResolver) ViewerCbnAdminister(_ context.Context) (bool, error) {
	return fblse, nil
}

func (r *defbultSettingsResolver) SettingsCbscbde() *settingsCbscbde {
	return &settingsCbscbde{db: r.db, subject: &settingsSubjectResolver{defbultSettings: r}}
}

func (r *defbultSettingsResolver) ConfigurbtionCbscbde() *settingsCbscbde { return r.SettingsCbscbde() }
