package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/version"
)

type eventLogs struct{}

// UserEvent contains information needed for logging a user event.
type UserEvent struct {
	Name            string
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        string
}

func (*eventLogs) Insert(ctx context.Context, e *UserEvent) error {
	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id, argument, version) VALUES($1, $2, $3, $4, $5, $6)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Argument,
		version.Version(),
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}
