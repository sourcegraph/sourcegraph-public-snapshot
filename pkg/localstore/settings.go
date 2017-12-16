package localstore

import (
	"context"
	"database/sql"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"

	"github.com/hashicorp/go-multierror"
	"github.com/keegancsmith/sqlf"
)

type settings struct{}

func (o *settings) CreateIfUpToDate(ctx context.Context, subject sourcegraph.ConfigurationSubject, lastKnownSettingsID *int32, authorAuth0ID, contents string) (latestSetting *sourcegraph.Settings, err error) {
	if Mocks.Settings.CreateIfUpToDate != nil {
		return Mocks.Settings.CreateIfUpToDate(ctx, subject, lastKnownSettingsID, authorAuth0ID, contents)
	}

	s := sourcegraph.Settings{
		Subject:       subject,
		AuthorAuth0ID: authorAuth0ID,
		Contents:      contents,
	}

	tx, err := globalDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			rollErr := tx.Rollback()
			if rollErr != nil {
				err = multierror.Append(err, rollErr)
			}
			return
		}
		err = tx.Commit()
	}()

	latestSetting, err = o.getLatest(ctx, tx, subject)
	if err != nil {
		return nil, err
	}

	creatorIsUpToDate := latestSetting != nil && lastKnownSettingsID != nil && latestSetting.ID == *lastKnownSettingsID
	if latestSetting == nil || creatorIsUpToDate {
		err := tx.QueryRow(
			"INSERT INTO settings(org_id, user_id, author_auth_id, contents) VALUES($1, $2, $3, $4) RETURNING id, created_at",
			s.Subject.Org, s.Subject.User, s.AuthorAuth0ID, s.Contents).Scan(&s.ID, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		latestSetting = &s
	}

	return latestSetting, nil
}

func (o *settings) GetLatest(ctx context.Context, subject sourcegraph.ConfigurationSubject) (*sourcegraph.Settings, error) {
	if Mocks.Settings.GetLatest != nil {
		return Mocks.Settings.GetLatest(ctx, subject)
	}

	return o.getLatest(ctx, globalDB, subject)
}

func (o *settings) getLatest(ctx context.Context, queryTarget queryable, subject sourcegraph.ConfigurationSubject) (*sourcegraph.Settings, error) {
	var cond *sqlf.Query
	switch {
	case subject.Org != nil:
		cond = sqlf.Sprintf("org_id=%d", *subject.Org)
	case subject.User != nil:
		cond = sqlf.Sprintf("user_id=%d", *subject.User)
	default:
		panic("no settings subject")
	}

	q := sqlf.Sprintf(`
		SELECT id, org_id, user_id, author_auth_id, contents, created_at FROM settings
		WHERE %s
		ORDER BY id DESC LIMIT 1`, cond)
	rows, err := queryTarget.QueryContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...)
	if err != nil {
		return nil, err
	}
	settings, err := o.parseQueryRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(settings) != 1 {
		// No configuration has been set for this subject yet.
		return nil, nil
	}
	return settings[0], nil
}

// queryable allows us to reuse the same logic for certain operations both
// inside and outside an explicit transaction.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func (o *settings) parseQueryRows(ctx context.Context, rows *sql.Rows) ([]*sourcegraph.Settings, error) {
	settings := []*sourcegraph.Settings{}
	defer rows.Close()
	for rows.Next() {
		s := sourcegraph.Settings{}
		err := rows.Scan(&s.ID, &s.Subject.Org, &s.Subject.User, &s.AuthorAuth0ID, &s.Contents, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		settings = append(settings, &s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return settings, nil
}
