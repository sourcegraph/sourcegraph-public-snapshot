package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

type userEventLogResolver struct {
	event *types.Event
}

func (s *userEventLogResolver) User(ctx context.Context) (*UserResolver, error) {
	if s.event.UserID != nil {
		user, err := UserByIDInt32(ctx, *s.event.UserID)
		if err != nil && errcode.IsNotFound(err) {
			// This can happen if the user has been deleted, see issue #4888 and #6454
			return nil, nil
		}
		return user, err
	}
	return nil, nil
}

func (s *userEventLogResolver) Name() string {
	return s.event.Name
}

func (s *userEventLogResolver) AnonymousUserID() string {
	return s.event.AnonymousUserID
}

func (s *userEventLogResolver) URL() string {
	return s.event.URL
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
