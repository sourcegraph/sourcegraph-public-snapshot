package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"time"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/backend"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/types"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/useractivity"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/actor"
)

func (r *userResolver) Activity(ctx context.Context) (*userActivityResolver, error) {
	// 🚨 SECURITY:  only admins are allowed to use this endpoint
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if r.user == nil {
		return nil, errors.New("Could not resolve activity on nil user")
	}
	activity, err := useractivity.GetByUserID(r.user.ID)
	if err != nil {
		return nil, err
	}
	return &userActivityResolver{activity}, nil
}

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
