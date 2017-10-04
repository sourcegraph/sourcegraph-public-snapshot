package localstore

import (
	"context"

	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type userTags struct{}

func (*userTags) Create(ctx context.Context, userID int32, name string) (*sourcegraph.UserTag, error) {
	t := sourcegraph.UserTag{
		UserID: userID,
		Name:   name,
	}
	err := globalDB.QueryRow(
		"INSERT INTO user_tags(user_id, name) VALUES($1, $2) RETURNING id",
		t.UserID, t.Name).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (*userTags) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.UserTag, error) {
	rows, err := globalDB.Query("SELECT user_id, name FROM user_tags "+query, args...)
	if err != nil {
		return nil, err
	}

	tags := []*sourcegraph.UserTag{}
	defer rows.Close()
	for rows.Next() {
		t := sourcegraph.UserTag{}
		err := rows.Scan(&t.ID, &t.UserID, &t.Name)
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

func (t *userTags) GetByUserID(ctx context.Context, userID int32) ([]*sourcegraph.UserTag, error) {
	return t.getBySQL(ctx, "WHERE user_id=$1 AND deleted_at IS NULL", userID)
}
