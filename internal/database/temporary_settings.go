package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	ts "github.com/sourcegraph/sourcegraph/internal/temporarysettings"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type TemporarySettingsStore interface {
	basestore.ShareableStore
	GetTemporarySettings(ctx context.Context, userID int32) (*ts.TemporarySettings, error)
	OverwriteTemporarySettings(ctx context.Context, userID int32, contents string) error
	EditTemporarySettings(ctx context.Context, userID int32, settingsToEdit string) error
}

type temporarySettingsStore struct {
	*basestore.Store
}

// TemporarySettingsWith instantiates and returns a new TemporarySettingsStore using the other store handle.
func TemporarySettingsWith(other basestore.ShareableStore) TemporarySettingsStore {
	return &temporarySettingsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (f *temporarySettingsStore) GetTemporarySettings(ctx context.Context, userID int32) (*ts.TemporarySettings, error) {
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

func (f *temporarySettingsStore) OverwriteTemporarySettings(ctx context.Context, userID int32, contents string) error {
	const overwriteTemporarySettingsQuery = `
		INSERT INTO temporary_settings (user_id, contents)
		VALUES (%s, %s)
		ON CONFLICT (user_id) DO UPDATE SET
			contents = %s,
			updated_at = now();
	`

	return f.Exec(ctx, sqlf.Sprintf(overwriteTemporarySettingsQuery, userID, contents, contents))
}

func (f *temporarySettingsStore) EditTemporarySettings(ctx context.Context, userID int32, settingsToEdit string) error {
	const editTemporarySettingsQuery = `
		INSERT INTO temporary_settings AS t (user_id, contents)
			VALUES (%s, %s)
			ON CONFLICT (user_id) DO UPDATE SET
				contents = t.contents || %s,
				updated_at = now();
	`

	return f.Exec(ctx, sqlf.Sprintf(editTemporarySettingsQuery, userID, settingsToEdit, settingsToEdit))
}
