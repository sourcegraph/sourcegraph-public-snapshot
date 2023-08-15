package database

import (
	"context"
	"database/sql"
	"strings"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type SettingsStore interface {
	CreateIfUpToDate(ctx context.Context, subject api.SettingsSubject, lastID *int32, authorUserID *int32, contents string) (*api.Settings, error)
	Done(error) error
	GetLatestSchemaSettings(context.Context, api.SettingsSubject) (*schema.Settings, error)
	GetLatest(context.Context, api.SettingsSubject) (*api.Settings, error)
	ListAll(ctx context.Context, impreciseSubstring string) ([]*api.Settings, error)
	Transact(context.Context) (SettingsStore, error)
	With(basestore.ShareableStore) SettingsStore
	basestore.ShareableStore
}

type settingsStore struct {
	*basestore.Store
}

// SettingsWith instantiates and returns a new SettingsStore using the other store handle.
func SettingsWith(other basestore.ShareableStore) SettingsStore {
	return &settingsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *settingsStore) With(other basestore.ShareableStore) SettingsStore {
	return &settingsStore{Store: s.Store.With(other)}
}

func (s *settingsStore) Transact(ctx context.Context) (SettingsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &settingsStore{Store: txBase}, err
}

func (o *settingsStore) CreateIfUpToDate(ctx context.Context, subject api.SettingsSubject, lastID *int32, authorUserID *int32, contents string) (latestSetting *api.Settings, err error) {
	// Validate settings for syntax and by the JSON Schema.
	if problems := conf.ValidateSettings(contents); len(problems) > 0 {
		return nil, errors.Errorf("invalid settings: %s", strings.Join(problems, ","))
	}

	s := api.Settings{
		Subject:      subject,
		AuthorUserID: authorUserID,
		Contents:     contents,
	}

	tx, err := o.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		err = tx.Done(err)
	}()

	latestSetting, err = tx.GetLatest(ctx, subject)
	if err != nil {
		return nil, err
	}

	creatorIsUpToDate := latestSetting != nil && lastID != nil && latestSetting.ID == *lastID
	if latestSetting == nil || creatorIsUpToDate {
		err := tx.Handle().QueryRowContext(
			ctx,
			"INSERT INTO settings(org_id, user_id, author_user_id, contents) VALUES($1, $2, $3, $4) RETURNING id, created_at",
			s.Subject.Org, s.Subject.User, s.AuthorUserID, s.Contents).Scan(&s.ID, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		latestSetting = &s
	}

	return latestSetting, nil
}

func (o *settingsStore) GetLatest(ctx context.Context, subject api.SettingsSubject) (*api.Settings, error) {
	var cond *sqlf.Query
	switch {
	case subject.Org != nil:
		cond = sqlf.Sprintf("org_id=%d", *subject.Org)
	case subject.User != nil:
		cond = sqlf.Sprintf("user_id=%d AND EXISTS (SELECT NULL FROM users WHERE id=%d AND deleted_at IS NULL)", *subject.User, *subject.User)
	default:
		// No org and no user represents global site settings.
		cond = sqlf.Sprintf("user_id IS NULL AND org_id IS NULL")
	}

	q := sqlf.Sprintf(`
		SELECT s.id, s.org_id, s.user_id, CASE WHEN users.deleted_at IS NULL THEN s.author_user_id ELSE NULL END, s.contents, s.created_at FROM settings s
		LEFT JOIN users ON users.id=s.author_user_id
		WHERE %s
		ORDER BY id DESC LIMIT 1`, cond)
	rows, err := o.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	settings, err := parseQueryRows(rows)
	if err != nil {
		return nil, err
	}
	if len(settings) != 1 {
		// No configuration has been set for this subject yet.
		return nil, nil
	}
	return settings[0], nil
}

func (o *settingsStore) GetLatestSchemaSettings(ctx context.Context, subject api.SettingsSubject) (*schema.Settings, error) {
	apiSettings, err := o.GetLatest(ctx, subject)
	if err != nil {
		return nil, err
	}

	if apiSettings == nil {
		// Settings have never been saved for this subject; equivalent to `{}`.
		return &schema.Settings{}, nil
	}

	var v schema.Settings
	if err := jsonc.Unmarshal(apiSettings.Contents, &v); err != nil {
		return nil, err
	}

	return &v, nil

}

// ListAll lists ALL settings (across all users, orgs, etc).
//
// If impreciseSubstring is given, only settings whose raw JSONC string contains the substring are
// returned. This is only intended for use by migration code that needs to update settings, where
// limiting to matching settings can significantly narrow the amount of work necessary. For example,
// a migration to rename a settings property `foo` to `bar` would narrow the search space by only
// processing settings that contain the string `foo` (which will yield false positives, but that's
// acceptable because the migration shouldn't modify those results anyway).
//
// 🚨 SECURITY: This method does NOT verify the user is an admin. The caller is
// responsible for ensuring this or that the response never makes it to a user.
func (o *settingsStore) ListAll(ctx context.Context, impreciseSubstring string) (_ []*api.Settings, err error) {
	tr, ctx := trace.New(ctx, "database.Settings.ListAll")
	defer tr.EndWithErr(&err)

	q := sqlf.Sprintf(`
		WITH q AS (
			SELECT DISTINCT
				ON (org_id, user_id, author_user_id)
				id, org_id, user_id, author_user_id, contents, created_at
				FROM settings
				ORDER BY org_id, user_id, author_user_id, id DESC
		)
		SELECT q.id, q.org_id, q.user_id, CASE WHEN users.deleted_at IS NULL THEN q.author_user_id ELSE NULL END, q.contents, q.created_at
		FROM q
		LEFT JOIN users ON users.id=q.author_user_id
		WHERE contents LIKE %s
		ORDER BY q.org_id, q.user_id, q.author_user_id, q.id DESC
	`, "%"+impreciseSubstring+"%")
	rows, err := o.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	return parseQueryRows(rows)
}

func parseQueryRows(rows *sql.Rows) ([]*api.Settings, error) {
	settings := []*api.Settings{}
	defer rows.Close()
	for rows.Next() {
		s := api.Settings{}
		err := rows.Scan(&s.ID, &s.Subject.Org, &s.Subject.User, &s.AuthorUserID, &s.Contents, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		if s.Subject.Org == nil && s.Subject.User == nil {
			s.Subject.Site = true
		}
		settings = append(settings, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return settings, nil
}
