package localstore

import (
	"context"
	"fmt"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
)

type userActivity struct{}

type ErrUserActivityNotFound struct {
	args []interface{}
}

func (err ErrUserActivityNotFound) Error() string {
	return fmt.Sprintf("server user events not found: %v", err.args)

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
		return nil, err
	}
	return s, nil
}

// Create if the server user does not yet exist in the table
func (s *userActivity) CreateIfNotExists(ctx context.Context, userID int32) (*sourcegraph.UserActivity, error) {
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

	user, err := s.GetByUserID(ctx, userID)
	if err != nil {
		if _, ok := err.(ErrUserActivityNotFound); !ok {
			return nil, err
		}
		return s.Create(ctx, userID)
	}
	return user, nil
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
