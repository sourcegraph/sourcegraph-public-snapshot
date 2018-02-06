package graphqlbackend

import (
	"context"
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/useractivity"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

type userActivityResolver struct {
	userActivity *types.UserActivity
}

func (s *userActivityResolver) PageViews() int32 { return s.userActivity.PageViews }

func (s *userActivityResolver) SearchQueries() int32 { return s.userActivity.SearchQueries }

func (s *userActivityResolver) LastPageViewTime() string {
	if s.userActivity.LastPageViewTime != nil {
		return s.userActivity.LastPageViewTime.Format(time.RFC3339)
	}
	return ""
}

func (s *schemaResolver) LogUserEvent(ctx context.Context, args *struct {
	Event        string
	UserCookieID string
}) (*EmptyResponse, error) {
	actor := actor.FromContext(ctx)
	switch args.Event {
	case "SEARCHQUERY":
		if !actor.IsAuthenticated() {
			return nil, nil
		}
		return nil, useractivity.LogSearchQuery(actor.UID)
	case "PAGEVIEW":
		return nil, useractivity.LogPageView(actor.IsAuthenticated(), actor.UID, args.UserCookieID)
	}
	return nil, fmt.Errorf("unknown user event %s", args.Event)
}
