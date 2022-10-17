package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
)

type userEventLogResolver struct {
	db    database.DB
	event *database.Event
}

func (s *userEventLogResolver) User(ctx context.Context) (*UserResolver, error) {
	user, err := UserByIDInt32(ctx, s.db, int32(s.event.UserID))
	if err != nil && errcode.IsNotFound(err) {
		// Don't throw an error if a user has been deleted.
		return nil, nil
	}
	return user, err
}

func (s *userEventLogResolver) Name() string {
	return s.event.Name
}

func (s *userEventLogResolver) AnonymousUserID() string {
	return s.event.AnonymousUserID
}

func (s *userEventLogResolver) URL() string {
	// 🚨 SECURITY: It is important to sanitize event URL before responding to the
	// client to prevent malicious data being rendered in browser.
	return database.SanitizeEventURL(s.event.URL)
}

func (s *userEventLogResolver) Source() string {
	return s.event.Source
}

func (s *userEventLogResolver) Argument() *string {
	if s.event.Argument == nil {
		return nil
	}
	st := string(s.event.Argument)
	return &st
}

func (s *userEventLogResolver) Version() string {
	return s.event.Version
}

func (s *userEventLogResolver) Timestamp() gqlutil.DateTime {
	return gqlutil.DateTime{Time: s.event.Timestamp}
}
