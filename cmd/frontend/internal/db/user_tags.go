package db

import (
	"context"
	"fmt"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

type userTags struct{}

type ErrUserTagNotFound struct {
	args []interface{}
}

func (err ErrUserTagNotFound) Error() string {
	return fmt.Sprintf("tag not found: %v", err.args)
}

func (*userTags) Create(ctx context.Context, userID int32, name string) (*types.UserTag, error) {
	t := &types.UserTag{
		UserID: userID,
		Name:   name,
	}
	err := globalDB.QueryRowContext(
		ctx,
		"INSERT INTO user_tags(user_id, name) VALUES($1, $2) RETURNING id",
		userID, name).Scan(&t.ID)
	if err != nil {
		return nil, err
	}
	return t, nil
}

// Create a tag for the user if the user does not already have the tag
func (t *userTags) CreateIfNotExists(ctx context.Context, userID int32, name string) (*types.UserTag, error) {
	tag, err := t.GetByUserIDAndTagName(ctx, userID, name)
	if err != nil {
		if _, ok := err.(ErrUserTagNotFound); !ok {
			return nil, err
		}
		// Create if the user does not have the tag in the table
		return t.Create(ctx, userID, name)
	}
	return tag, nil
}

func (*userTags) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.UserTag, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, user_id, name FROM user_tags "+query, args...)
	if err != nil {
		return nil, err
	}
	tags := []*types.UserTag{}
	defer rows.Close()
	for rows.Next() {
		t := types.UserTag{}
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

func (t *userTags) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*types.UserTag, error) {
	rows, err := t.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, ErrUserTagNotFound{args}
	}
	return rows[0], nil
}

func (t *userTags) GetByUserID(ctx context.Context, userID int32) ([]*types.UserTag, error) {
	return t.getBySQL(ctx, "WHERE user_id=$1 AND deleted_at IS NULL", userID)
}

func (t *userTags) GetByUserIDAndTagName(ctx context.Context, userID int32, name string) (*types.UserTag, error) {
	return t.getOneBySQL(ctx, "WHERE user_id=$1 AND name=$2 AND deleted_at IS NULL LIMIT 1", userID, name)
}
