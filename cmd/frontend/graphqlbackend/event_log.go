package graphqlbackend

import (
	"context"
	"net/url"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type userEventLogResolver struct {
	db    dbutil.DB
	event *types.Event
}

func (s *userEventLogResolver) User(ctx context.Context) (*UserResolver, error) {
	if s.event.UserID != nil {
		user, err := UserByIDInt32(ctx, s.db, *s.event.UserID)
		if err != nil && errcode.IsNotFound(err) {
			// Don't throw an error if a user has been deleted.
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
	if s.event.URL == "" {
		return ""
	}

	// Check if the URL looks like a real URL
	u, err := url.Parse(s.event.URL)
	if err != nil ||
		(u.Scheme != "http" && u.Scheme != "https") {
		return ""
	}

	// Check if the URL belongs to the current site
	normalized := u.String()
	if !strings.HasPrefix(normalized, conf.ExternalURL()) {
		return ""
	}
	return normalized
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
