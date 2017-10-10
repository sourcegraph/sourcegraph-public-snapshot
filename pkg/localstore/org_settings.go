package localstore

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgSettings struct{}

func (*orgSettings) Create(ctx context.Context, orgID int32, authorAuth0ID, contents string) (*sourcegraph.OrgSettings, error) {
	s := sourcegraph.OrgSettings{
		OrgID:         orgID,
		AuthorAuth0ID: authorAuth0ID,
		Contents:      contents,
	}

	err := globalDB.QueryRow(
		"INSERT INTO org_settings(org_id, author_auth0_id, contents) VALUES($1, $2, $3) RETURNING id, created_at",
		s.OrgID, s.AuthorAuth0ID, s.Contents).Scan(&s.ID, &s.CreatedAt)

	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (o *orgSettings) GetLatestByOrgID(ctx context.Context, orgID int32) (*sourcegraph.OrgSettings, error) {
	settings, err := o.getBySQL(ctx, "WHERE org_id = $1 ORDER BY id DESC LIMIT 1", orgID)
	if err != nil {
		return nil, err
	}
	if len(settings) != 1 {
		// No configuration has been set for this org yet.
		return nil, nil
	}
	return settings[0], nil
}

func (*orgSettings) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.OrgSettings, error) {
	rows, err := globalDB.Query("SELECT id, org_id, author_auth0_id, contents, created_at FROM org_settings "+query, args...)
	if err != nil {
		return nil, err
	}

	settings := []*sourcegraph.OrgSettings{}
	defer rows.Close()
	for rows.Next() {
		s := sourcegraph.OrgSettings{}
		err := rows.Scan(&s.ID, &s.OrgID, &s.AuthorAuth0ID, &s.Contents, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		settings = append(settings, &s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return settings, nil
}
