package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/version"
)

// SecurityEvent contains information needed for logging a security-relevant event.
type SecurityEvent struct {
	Name            string
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        json.RawMessage
	Source          string
	Timestamp       time.Time
}

// A SecurityEventLogStore provides persistence for security events.
type SecurityEventLogStore struct {
	*basestore.Store
}

// SecurityEventLogs instantiates and returns a new SecurityEventLogStore with prepared statements.
func SecurityEventLogs(db dbutil.DB) *SecurityEventLogStore {
	return &SecurityEventLogStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// Insert adds a new security event to the store.
func (s *SecurityEventLogStore) Insert(ctx context.Context, e *SecurityEvent) error {
	argument := e.Argument
	if argument == nil {
		argument = json.RawMessage([]byte(`{}`))
	}

	_, err := s.Handle().DB().ExecContext(
		ctx,
		"INSERT INTO security_event_logs(name, url, user_id, anonymous_user_id, source, argument, version, timestamp) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
		e.Name,
		e.URL,
		e.UserID,
		e.AnonymousUserID,
		e.Source,
		argument,
		version.Version(),
		e.Timestamp.UTC(),
	)
	if err != nil {
		return errors.Wrap(err, "INSERT")
	}
	return nil
}
