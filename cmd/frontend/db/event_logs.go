package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/version"
)

type eventLogs struct{}

// Event contains information needed for logging an event.
type Event struct {
	Name            string
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        string
	Source          string
}

func (*eventLogs) Insert(ctx context.Context, e *Event) error {
	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id, source, argument, version) VALUES($1, $2, $3, $4, $5, $6, $7)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Source,
		e.Argument,
		version.Version(),
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}
