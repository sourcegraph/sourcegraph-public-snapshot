package db

import (
	"context"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
)

type eventLogs struct{}

type EventLogInfo struct {
	Name            string
	URL             string
	UserID          int32
	AnonymousUserID string
	Argument        string
}

func (*eventLogs) Insert(ctx context.Context, info *EventLogInfo) error {
	if info.Name == "" {
		return errors.New("empty event name")
	} else if info.UserID <= 0 || info.AnonymousUserID == "" {
		return errors.New("one of UserID or AnonymousUserID must have valid value")
	}

	_, err := dbconn.Global.ExecContext(
		ctx,
		"INSERT INTO event_logs(name, url, user_id, anonymous_user_id,argument) VALUES($1, $2, $3, $4, $5)",
		info.Name,
		info.URL,
		info.UserID,
		info.AnonymousUserID,
		info.Argument,
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}
