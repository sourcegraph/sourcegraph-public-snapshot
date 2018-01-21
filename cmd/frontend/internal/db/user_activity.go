package db

import (
	"context"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
)

type userActivity struct{}

func (*userActivity) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.UserActivity, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, page_views, search_queries FROM users "+query, args...)
	if err != nil {
		return nil, err
	}
	events := []*types.UserActivity{}
	defer rows.Close()
	for rows.Next() {
		e := types.UserActivity{}
		err := rows.Scan(&e.UserID, &e.PageViews, &e.SearchQueries)
		if err != nil {
			return nil, err
		}
		events = append(events, &e)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return events, nil
}

func (s *userActivity) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*types.UserActivity, error) {
	rows, err := s.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, ErrUserNotFound{args}
	}

	return rows[0], nil
}

func (s *userActivity) GetByUserID(ctx context.Context, userID int32) (*types.UserActivity, error) {
	return s.getOneBySQL(ctx, "WHERE id=$1", userID)
}

func (s *userActivity) LogPageView(ctx context.Context, userID int32) error {
	_, err := globalDB.ExecContext(ctx, "UPDATE users SET page_views=users.page_views + 1 WHERE users.id=$1", userID)
	return err
}

func (s *userActivity) LogSearchQuery(ctx context.Context, userID int32) error {
	_, err := globalDB.ExecContext(ctx, "UPDATE users SET search_queries=users.search_queries + 1 WHERE users.id=$1", userID)
	return err
}
