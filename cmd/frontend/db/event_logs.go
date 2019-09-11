package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type eventLogs struct{}

// UserEvent contains information needed for logging a user event.
type UserEvent struct {
	Name            string
	URL             string
	UserID          int32
	AnonymousUserID string
	Argument        string
}

func (*eventLogs) Insert(ctx context.Context, e *UserEvent) error {
	if e.Name == "" {
		return errors.New("empty event name")
	} else if e.UserID <= 0 && e.AnonymousUserID == "" {
		return errors.New("one of UserID or AnonymousUserID must have valid value")
	}

	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id, argument) VALUES($1, $2, $3, $4, $5)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Argument,
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}
