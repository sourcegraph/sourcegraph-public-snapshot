package localstore

import (
	"context"
	"database/sql"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"

	"github.com/hashicorp/go-multierror"
)

type settings struct{}

func (o *settings) CreateIfUpToDate(ctx context.Context, orgID int32, lastKnownSettingsID *int32, authorAuth0ID, contents string) (latestSetting *sourcegraph.Settings, err error) {
	s := sourcegraph.Settings{
		OrgID:         orgID,
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

	latestSetting, err = o.getLatestByOrgID(ctx, tx, orgID)
	if err != nil {
		return nil, err
	}

	creatorIsUpToDate := latestSetting != nil && lastKnownSettingsID != nil && latestSetting.ID == *lastKnownSettingsID
	if latestSetting == nil || creatorIsUpToDate {
		err := tx.QueryRow(
			"INSERT INTO settings(org_id, author_auth_id, contents) VALUES($1, $2, $3) RETURNING id, created_at",
			s.OrgID, s.AuthorAuth0ID, s.Contents).Scan(&s.ID, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		latestSetting = &s
	}

	return latestSetting, nil
}

func (o *settings) GetLatestByOrgID(ctx context.Context, orgID int32) (*sourcegraph.Settings, error) {
	return o.getLatestByOrgID(ctx, globalDB, orgID)
}

func (o *settings) getLatestByOrgID(ctx context.Context, queryTarget queryable, orgID int32) (*sourcegraph.Settings, error) {
	rows, err := queryTarget.QueryContext(ctx, getLatestByOrgIDSql, orgID)
	if err != nil {
		return nil, err
	}
	settings, err := o.parseQueryRows(ctx, rows)
	if err != nil {
		return nil, err
	}
	if len(settings) != 1 {
		// No configuration has been set for this org yet.
		return nil, nil
	}
	return settings[0], nil
}

// queryable allows us to reuse the same logic for certain operations both
// inside and outside an explicit transaction.
type queryable interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

const getLatestByOrgIDSql = "SELECT id, org_id, author_auth_id, contents, created_at FROM settings WHERE org_id = $1 ORDER BY id DESC LIMIT 1"

func (o *settings) parseQueryRows(ctx context.Context, rows *sql.Rows) ([]*sourcegraph.Settings, error) {
	settings := []*sourcegraph.Settings{}
	defer rows.Close()
	for rows.Next() {
		s := sourcegraph.Settings{}
		err := rows.Scan(&s.ID, &s.OrgID, &s.AuthorAuth0ID, &s.Contents, &s.CreatedAt)
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
