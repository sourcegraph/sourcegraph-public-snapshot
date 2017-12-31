package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/db"
)

type userActivityResolver struct {
	userActivity *sourcegraph.UserActivity
}

func (s *userActivityResolver) ID() int32 {
	return s.userActivity.ID
}

func (s *userActivityResolver) UserID() int32 {
	return s.userActivity.UserID
}

func (s *userActivityResolver) PageViews() int32 {
	return s.userActivity.PageViews
}

func (s *userActivityResolver) SearchQueries() int32 {
	return s.userActivity.SearchQueries
}

func (s *userActivityResolver) CreatedAt() string {
	t := s.userActivity.CreatedAt.Format(time.RFC3339) // ISO
	return t
}

func (s *userActivityResolver) UpdatedAt() string {
	t := s.userActivity.UpdatedAt.Format(time.RFC3339) // ISO
	return t
}

func (s *schemaResolver) LogUserEvent(ctx context.Context, args *struct {
	Event string
}) (*EmptyResponse, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, errors.New("must be authenticated")
	}
	switch args.Event {
	case "SEARCHQUERY":
		return nil, db.UserActivity.LogSearchQuery(ctx, actor.UID)
	case "PAGEVIEW":
		return nil, db.UserActivity.LogPageView(ctx, actor.UID)
	}
	return nil, fmt.Errorf("unknown user event %s", args.Event)
}
