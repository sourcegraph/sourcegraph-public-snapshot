package localstore

import (
	"context"
	"fmt"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type orgTags struct{}

type ErrOrgTagNotFound struct {
	args []interface{}
}

func (err ErrOrgTagNotFound) Error() string {
	return fmt.Sprintf("tag not found: %v", err.args)
}

func (*orgTags) Create(ctx context.Context, orgID int32, name string) (*sourcegraph.OrgTag, error) {
	t := &sourcegraph.OrgTag{
		OrgID: orgID,
		Name:  name,
	}
	err := globalDB.QueryRow(
		"INSERT INTO org_tags(org_id, name) VALUES($1, $2) RETURNING id",
		t.OrgID, t.Name).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *orgTags) CreateIfNotExists(ctx context.Context, orgID int32, name string) (*sourcegraph.OrgTag, error) {
	tag, err := t.GetByOrgIDAndTagName(ctx, orgID, name)
	if err != nil {
		if _, ok := err.(ErrOrgTagNotFound); !ok {
			return nil, err
		}
		// Create if the org does not have the tag in the table
		return t.Create(ctx, orgID, name)
	}
	return tag, nil
}

func (*orgTags) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.OrgTag, error) {
	rows, err := globalDB.Query("SELECT id, org_id, name FROM org_tags "+query, args...)
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

func (t *orgTags) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.OrgTag, error) {
	rows, err := t.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, ErrOrgTagNotFound{args}
	}
	return rows[0], nil
}

func (t *orgTags) GetByOrgID(ctx context.Context, orgID int32) ([]*sourcegraph.OrgTag, error) {
	return t.getBySQL(ctx, "WHERE org_id=$1 AND deleted_at IS NULL", orgID)
}

func (t *orgTags) GetByOrgIDAndTagName(ctx context.Context, orgID int32, name string) (*sourcegraph.OrgTag, error) {
	return t.getOneBySQL(ctx, "WHERE org_id=$1 AND name=$2 AND deleted_at IS NULL LIMIT 1", orgID, name)
}
