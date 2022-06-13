package database

import (
	"context"
	"encoding/json"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type SecurityEventName string

const (
	SecurityEventNameSignOutAttempted SecurityEventName = "SignOutAttempted"
	SecurityEventNameSignOutFailed    SecurityEventName = "SignOutFailed"
	SecurityEventNameSignOutSucceeded SecurityEventName = "SignOutSucceeded"

	SecurityEventNameSignInAttempted SecurityEventName = "SignInAttempted"
	SecurityEventNameSignInFailed    SecurityEventName = "SignInFailed"
	SecurityEventNameSignInSucceeded SecurityEventName = "SignInSucceeded"

	SecurityEventNameAccountCreated SecurityEventName = "AccountCreated"
	SecurityEventNameAccountDeleted SecurityEventName = "AccountDeleted"
	SecurityEventNameAccountNuked   SecurityEventName = "AccountNuked"

	SecurityEventNamPasswordResetRequested SecurityEventName = "PasswordResetRequested"
	SecurityEventNamPasswordRandomized     SecurityEventName = "PasswordRandomized"
	SecurityEventNamePasswordChanged       SecurityEventName = "PasswordChanged"

	SecurityEventNameEmailVerified SecurityEventName = "EmailVerified"

	SecurityEventNameRoleChangeDenied  SecurityEventName = "RoleChangeDenied"
	SecurityEventNameRoleChangeGranted SecurityEventName = "RoleChangeGranted"

	SecurityEventNameAccessGranted SecurityEventName = "AccessGranted"
)

// SecurityEvent contains information needed for logging a security-relevant event.
type SecurityEvent struct {
	Name            SecurityEventName
	URL             string
	UserID          uint32
	AnonymousUserID string
	Argument        json.RawMessage
	Source          string
	Timestamp       time.Time
}

// SecurityEventLogsStore provides persistence for security events.
type SecurityEventLogsStore interface {
	basestore.ShareableStore

	// Insert adds a new security event to the store.
	Insert(ctx context.Context, e *SecurityEvent) error
	// LogEvent logs the given security events.
	//
	// It logs errors directly instead of returning to callers.
	LogEvent(ctx context.Context, e *SecurityEvent)
}

type securityEventLogsStore struct {
	*basestore.Store
}

// SecurityEventLogsWith instantiates and returns a new SecurityEventLogsStore
// using the other store handle.
func SecurityEventLogsWith(other basestore.ShareableStore) SecurityEventLogsStore {
	return &securityEventLogsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *securityEventLogsStore) Insert(ctx context.Context, e *SecurityEvent) error {
	argument := e.Argument
	if argument == nil {
		argument = []byte(`{}`)
	}

	_, err := s.Handle().ExecContext(
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

func (s *securityEventLogsStore) LogEvent(ctx context.Context, e *SecurityEvent) {
	// We don't want to begin logging authentication or authorization events in
	// on-premises installations yet.
	if !envvar.SourcegraphDotComMode() {
		return
	}

	if err := s.Insert(ctx, e); err != nil {
		j, _ := json.Marshal(e)
		log15.Error(string(e.Name), "event", string(j), "traceID", trace.ID(ctx), "error", err)
		// We want to capture in sentry as it includes a stack trace which will allow us
		// to track down the root cause.
		sentry.CaptureError(err, map[string]string{})
	}
}
