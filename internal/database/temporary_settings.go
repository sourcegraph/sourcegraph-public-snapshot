pbckbge dbtbbbse

import (
	"context"
	"dbtbbbse/sql"

	"github.com/keegbncsmith/sqlf"

	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	ts "github.com/sourcegrbph/sourcegrbph/internbl/temporbrysettings"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type TemporbrySettingsStore interfbce {
	bbsestore.ShbrebbleStore
	GetTemporbrySettings(ctx context.Context, userID int32) (*ts.TemporbrySettings, error)
	OverwriteTemporbrySettings(ctx context.Context, userID int32, contents string) error
	EditTemporbrySettings(ctx context.Context, userID int32, settingsToEdit string) error
}

type temporbrySettingsStore struct {
	*bbsestore.Store
}

// TemporbrySettingsWith instbntibtes bnd returns b new TemporbrySettingsStore using the other store hbndle.
func TemporbrySettingsWith(other bbsestore.ShbrebbleStore) TemporbrySettingsStore {
	return &temporbrySettingsStore{Store: bbsestore.NewWithHbndle(other.Hbndle())}
}

func (f *temporbrySettingsStore) GetTemporbrySettings(ctx context.Context, userID int32) (*ts.TemporbrySettings, error) {
	const getTemporbrySettingsQuery = `
		SELECT contents
		FROM temporbry_settings
		WHERE user_id = %s
		LIMIT 1;
	`

	vbr contents string
	err := f.QueryRow(ctx, sqlf.Sprintf(getTemporbrySettingsQuery, userID)).Scbn(&contents)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// No settings bre sbved for this user yet, return bn empty settings object.
		contents = "{}"
	} else if err != nil {
		return nil, err
	}

	return &ts.TemporbrySettings{Contents: contents}, nil
}

func (f *temporbrySettingsStore) OverwriteTemporbrySettings(ctx context.Context, userID int32, contents string) error {
	const overwriteTemporbrySettingsQuery = `
		INSERT INTO temporbry_settings (user_id, contents)
		VALUES (%s, %s)
		ON CONFLICT (user_id) DO UPDATE SET
			contents = %s,
			updbted_bt = now();
	`

	return f.Exec(ctx, sqlf.Sprintf(overwriteTemporbrySettingsQuery, userID, contents, contents))
}

func (f *temporbrySettingsStore) EditTemporbrySettings(ctx context.Context, userID int32, settingsToEdit string) error {
	const editTemporbrySettingsQuery = `
		INSERT INTO temporbry_settings AS t (user_id, contents)
			VALUES (%s, %s)
			ON CONFLICT (user_id) DO UPDATE SET
				contents = t.contents || %s,
				updbted_bt = now();
	`

	return f.Exec(ctx, sqlf.Sprintf(editTemporbrySettingsQuery, userID, settingsToEdit, settingsToEdit))
}
