package db

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
)

type userActivity struct{}

func (*userActivity) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*types.UserActivity, error) {
	rows, err := dbconn.Global.QueryContext(ctx, "SELECT id, page_views, search_queries FROM users "+query, args...)
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

// DEPRECATED: use package useractivity instead
func (s *userActivity) GetAll(ctx context.Context) ([]*types.UserActivity, error) {
	return s.getBySQL(ctx, "")
}
