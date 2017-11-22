package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/localstore"
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
	user, err := localstore.Users.GetByAuth0ID(ctx, actor.UID)
	if err != nil {
		return nil, err
	}
	switch args.Event {
	case "SEARCHQUERY":
		return nil, localstore.UserActivity.LogSearchQuery(ctx, user.ID)
	case "PAGEVIEW":
		return nil, localstore.UserActivity.LogPageView(ctx, user.ID)
	}
	return nil, fmt.Errorf("unknown user event %s", args.Event)
}
