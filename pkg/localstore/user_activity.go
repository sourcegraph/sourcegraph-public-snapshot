package localstore

import (
	"context"
	"fmt"
	"time"

	"github.com/lib/pq"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type userActivity struct{}

type ErrUserActivityNotFound struct {
	args []interface{}
}

func (err ErrUserActivityNotFound) Error() string {
	return fmt.Sprintf("server user events not found: %v", err.args)
}

type ErrUserExistsInUserActivity struct{}

func (err ErrUserExistsInUserActivity) Error() string {
	return "user already exists in user activity table"
}

func (*userActivity) Create(ctx context.Context, userID int32) (*sourcegraph.UserActivity, error) {
	s := &sourcegraph.UserActivity{
		UserID: userID,
	}
	err := globalDB.QueryRowContext(
		ctx,
		"INSERT INTO user_activity(user_id) VALUES ($1) RETURNING id",
		userID).Scan(&s.ID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Constraint == "user_activity_user_id_key" {
				return nil, ErrUserExistsInUserActivity{}
			}
		}

		return nil, err
	}
	return s, nil
}

func (s *userActivity) CreateIfNotExists(ctx context.Context, userID int32) (*sourcegraph.UserActivity, error) {
	activity, err := s.Create(ctx, userID)
	if err != nil {
		if _, ok := err.(ErrUserExistsInUserActivity); !ok {
			return nil, err
		}
		activity, err = s.GetByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
	}
	return activity, nil
}

func (*userActivity) getBySQL(ctx context.Context, query string, args ...interface{}) ([]*sourcegraph.UserActivity, error) {
	rows, err := globalDB.QueryContext(ctx, "SELECT id, user_id, page_views, search_queries FROM user_activity "+query, args...)
	if err != nil {
		return nil, err
	}
	events := []*sourcegraph.UserActivity{}
	defer rows.Close()
	for rows.Next() {
		e := sourcegraph.UserActivity{}
		err := rows.Scan(&e.ID, &e.UserID, &e.PageViews, &e.SearchQueries)
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

func (s *userActivity) getOneBySQL(ctx context.Context, query string, args ...interface{}) (*sourcegraph.UserActivity, error) {
	rows, err := s.getBySQL(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	if len(rows) != 1 {
		return nil, ErrUserActivityNotFound{args}
	}

	return rows[0], nil
}

func (s *userActivity) GetByUserID(ctx context.Context, userID int32) (*sourcegraph.UserActivity, error) {
	return s.getOneBySQL(ctx, "WHERE user_id=$1", userID)
}

func (s *userActivity) LogPageView(ctx context.Context, userID int32) error {
	_, err := s.CreateIfNotExists(ctx, userID)
	if err != nil {
		return err
	}
	updatedAt := time.Now()
	_, err = globalDB.ExecContext(ctx, "UPDATE user_activity SET page_views = user_activity.page_views + 1, updated_at = $1 WHERE user_activity.user_id=$2", updatedAt, userID)
	return err
}

func (s *userActivity) LogSearchQuery(ctx context.Context, userID int32) error {
	_, err := s.CreateIfNotExists(ctx, userID)
	if err != nil {
		return err
	}
	updatedAt := time.Now()
	_, err = globalDB.ExecContext(ctx, "UPDATE user_activity SET search_queries = user_activity.search_queries + 1, updated_at = $1 WHERE user_activity.user_id=$2", updatedAt, userID)
	return err
}
