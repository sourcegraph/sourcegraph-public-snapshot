package graphqlbackend

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
)

func (r *UserResolver) EventLogs(ctx context.Context, args *struct {
	graphqlutil.ConnectionArgs
}) (*userEventLogsConnectionResolver, error) {
	// 🚨 SECURITY: Event logs can only be viewed by the user or site admin.
	if err := backend.CheckSiteAdminOrSameUser(ctx, r.user.ID); err != nil {
		return nil, err
	}
	var opt db.EventLogsListOptions
	args.ConnectionArgs.Set(&opt.LimitOffset)
	return &userEventLogsConnectionResolver{opt: opt, userID: r.user.ID}, nil
}

type userEventLogsConnectionResolver struct {
	opt    db.EventLogsListOptions
	userID int32
}

func (r *userEventLogsConnectionResolver) Nodes(ctx context.Context) ([]*userEventLogResolver, error) {
	events, err := db.EventLogs.GetAllByUserID(ctx, r.userID)
	if err != nil {
		return nil, err
	}

	eventLogs := make([]*userEventLogResolver, 0, len(events))
	for _, event := range events {
		eventLogs = append(eventLogs, &userEventLogResolver{event: event})
	}

	return eventLogs, nil
}

func (r *userEventLogsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := db.EventLogs.CountByUserID(ctx, r.userID)
	return int32(count), err
}

func (r *userEventLogsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	count, err := db.EventLogs.CountByUserID(ctx, r.userID)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && count > r.opt.Limit), nil
}
