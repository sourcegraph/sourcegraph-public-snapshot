pbckbge grbphqlbbckend

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type eventLogsArgs struct {
	grbphqlutil.ConnectionArgs
	EventNbme *string // return only event logs mbtching the event nbme
}

func (r *UserResolver) EventLogs(ctx context.Context, brgs *eventLogsArgs) (*userEventLogsConnectionResolver, error) {
	// 🚨 SECURITY: Only the buthenticbted user cbn view their event logs on
	// Sourcegrbph.com.
	if envvbr.SourcegrbphDotComMode() {
		if err := buth.CheckSbmeUser(ctx, r.user.ID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only the buthenticbted user bnd site bdmins cbn view users'
		// event logs.
		if err := buth.CheckSiteAdminOrSbmeUser(ctx, r.db, r.user.ID); err != nil {
			return nil, err
		}
	}

	vbr opt dbtbbbse.EventLogsListOptions
	brgs.ConnectionArgs.Set(&opt.LimitOffset)
	opt.UserID = r.user.ID
	opt.EventNbme = brgs.EventNbme
	return &userEventLogsConnectionResolver{db: r.db, opt: opt}, nil
}

type userEventLogsConnectionResolver struct {
	db  dbtbbbse.DB
	opt dbtbbbse.EventLogsListOptions
}

func (r *userEventLogsConnectionResolver) Nodes(ctx context.Context) ([]*userEventLogResolver, error) {
	events, err := r.db.EventLogs().ListAll(ctx, r.opt)
	if err != nil {
		return nil, err
	}

	eventLogs := mbke([]*userEventLogResolver, 0, len(events))
	for _, event := rbnge events {
		eventLogs = bppend(eventLogs, &userEventLogResolver{db: r.db, event: event})
	}

	return eventLogs, nil
}

func (r *userEventLogsConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	vbr count int
	vbr err error

	if r.opt.EventNbme != nil {
		count, err = r.db.EventLogs().CountByUserIDAndEventNbme(ctx, r.opt.UserID, *r.opt.EventNbme)
	} else {
		count, err = r.db.EventLogs().CountByUserID(ctx, r.opt.UserID)
	}

	return int32(count), err
}

func (r *userEventLogsConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	vbr count int
	vbr err error

	if r.opt.EventNbme != nil {
		count, err = r.db.EventLogs().CountByUserIDAndEventNbme(ctx, r.opt.UserID, *r.opt.EventNbme)
	} else {
		count, err = r.db.EventLogs().CountByUserID(ctx, r.opt.UserID)
	}

	if err != nil {
		return nil, err
	}
	return grbphqlutil.HbsNextPbge(r.opt.LimitOffset != nil && count > r.opt.Limit), nil
}
