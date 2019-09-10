package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type eventLogs struct{}

type EventLogInfo struct {
	Name            string
	Argument        string
	URL             string
	UserID          int32
	AnonymousUserID string
}

func (*eventLogs) Insert(ctx context.Context, info *EventLogInfo) error {
	if info.Name == "" {
		return errors.New("empty event name")
	} else if info.UserID <= 0 || info.AnonymousUserID == "" {
		return errors.New("one of UserID or AnonymousUserID must have valid value")
	}

	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, argument, url, user_id, anonymous_user_id) VALUES($1, $2, $3, $4, $5)",
		info.Name,
		info.Argument,
		info.URL,
		info.UserID,
		info.AnonymousUserID,
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}
