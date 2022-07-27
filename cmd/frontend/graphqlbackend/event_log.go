package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type userEventLogResolver struct {
	db    database.DB
	event *types.Event
}

func (s *userEventLogResolver) User(ctx context.Context) (*UserResolver, error) {
	user, err := UserByIDInt32(ctx, s.db, s.event.UserID)
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
	if s.event.Argument == "" {
		return nil
	}
	return &s.event.Argument
}

func (s *userEventLogResolver) Version() string {
	return s.event.Version
}

func (s *userEventLogResolver) Timestamp() DateTime {
	return DateTime{Time: s.event.Timestamp}
}
