package localstore

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgTags struct{}

func (*orgTags) Create(ctx context.Context, orgID int32, name string) (*sourcegraph.OrgTag, error) {
	t := &sourcegraph.OrgTag{
		OrgID: orgID,
		Name:  name,
	}
	err := globalDB.QueryRow(
		"INSERT INTO user_tags(org_id, name) VALUES($1, $2) RETURNING id",
		t.OrgID, t.Name).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (*orgTags) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.OrgTag, error) {
	rows, err := globalDB.Query("SELECT org_id, name FROM org_tags "+query, args...)
	if err != nil {
		return nil, err
	}

	tags := []*sourcegraph.OrgTag{}
	defer rows.Close()
	for rows.Next() {
		t := sourcegraph.OrgTag{}
		err := rows.Scan(&t.ID, &t.OrgID, &t.Name)
		if err != nil {
			return nil, err
		}
		tags = append(tags, &t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tags, nil
}

func (t *orgTags) GetByOrgID(ctx context.Context, orgID int32) ([]*sourcegraph.OrgTag, error) {
	return t.getBySQL(ctx, "WHERE org_id=$1 AND deleted_at IS NULL", orgID)
}
