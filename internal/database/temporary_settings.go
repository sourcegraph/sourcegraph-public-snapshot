package database

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
)

type TemporarySettingsStore struct {
	*basestore.Store
}

func TemporarySettings(db dbutil.DB) *TemporarySettingsStore {
	return &TemporarySettingsStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func (f *TemporarySettingsStore) GetTemporarySettings(ctx context.Context, userID int32) (*ts.TemporarySettings, error) {
	if Mocks.TemporarySettings.GetTemporarySettings != nil {
		return Mocks.TemporarySettings.GetTemporarySettings(ctx, userID)
	}

	const getTemporarySettingsQuery = `
		SELECT contents
		FROM temporary_settings
		WHERE user_id = %s
		LIMIT 1;
	`

	var contents string
	err := f.QueryRow(ctx, sqlf.Sprintf(getTemporarySettingsQuery, userID)).Scan(&contents)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		// No settings are saved for this user yet, return an empty settings object.
		contents = "{}"
	} else if err != nil {
		return nil, err
	}

	return &ts.TemporarySettings{Contents: contents}, nil
}

func (f *TemporarySettingsStore) OverwriteTemporarySettings(ctx context.Context, userID int32, contents string) error {
	if Mocks.TemporarySettings.OverwriteTemporarySettings != nil {
		return Mocks.TemporarySettings.OverwriteTemporarySettings(ctx, userID, contents)
	}

	const overwriteTemporarySettingsQuery = `
		INSERT INTO temporary_settings (user_id, contents)
		VALUES (%s, %s)
		ON CONFLICT (user_id) DO UPDATE SET
			contents = %s,
			updated_at = now();
	`

	return f.Exec(ctx, sqlf.Sprintf(overwriteTemporarySettingsQuery, userID, contents, contents))
}

func (f *TemporarySettingsStore) EditTemporarySettings(ctx context.Context, userID int32, settingsToEdit string) error {
	if Mocks.TemporarySettings.EditTemporarySettings != nil {
		return Mocks.TemporarySettings.EditTemporarySettings(ctx, userID, settingsToEdit)
	}

	const editTemporarySettingsQuery = `
		INSERT INTO temporary_settings AS t (user_id, contents)
			VALUES (%s, %s)
			ON CONFLICT (user_id) DO UPDATE SET
				contents = t.contents || %s,
				updated_at = now();
	`

	return f.Exec(ctx, sqlf.Sprintf(editTemporarySettingsQuery, userID, settingsToEdit, settingsToEdit))
}
