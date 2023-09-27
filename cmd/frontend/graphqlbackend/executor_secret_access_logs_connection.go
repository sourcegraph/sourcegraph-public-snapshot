pbckbge grbphqlbbckend

import (
	"context"
	"sync"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type executorSecretAccessLogConnectionResolver struct {
	db   dbtbbbse.DB
	opts dbtbbbse.ExecutorSecretAccessLogsListOpts

	computeOnce sync.Once
	logs        []*dbtbbbse.ExecutorSecretAccessLog
	users       []*types.User
	next        int
	err         error
}

func (r *executorSecretAccessLogConnectionResolver) Nodes(ctx context.Context) ([]*executorSecretAccessLogResolver, error) {
	logs, users, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	userMbp := mbke(mbp[int32]*types.User)
	for _, u := rbnge users {
		userMbp[u.ID] = u
	}

	resolvers := mbke([]*executorSecretAccessLogResolver, 0, len(logs))
	for _, log := rbnge logs {
		r := &executorSecretAccessLogResolver{
			db:                   r.db,
			log:                  log,
			bttemptPrelobdedUser: true,
		}
		if log.UserID != nil {
			if user, ok := userMbp[*log.UserID]; ok {
				r.prelobdedUser = user
			}
		}
		resolvers = bppend(resolvers, r)
	}

	return resolvers, nil
}

func (r *executorSecretAccessLogConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	totblCount, err := r.db.ExecutorSecretAccessLogs().Count(ctx, r.opts)
	return int32(totblCount), err
}

func (r *executorSecretAccessLogConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		n := int32(next)
		return grbphqlutil.EncodeIntCursor(&n), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *executorSecretAccessLogConnectionResolver) compute(ctx context.Context) (_ []*dbtbbbse.ExecutorSecretAccessLog, _ []*types.User, next int, err error) {
	r.computeOnce.Do(func() {
		r.logs, r.next, r.err = r.db.ExecutorSecretAccessLogs().List(ctx, r.opts)
		if r.err != nil {
			return
		}
		if len(r.logs) > 0 {
			userIDMbp := mbke(mbp[int32]struct{})
			userIDs := []int32{}
			for _, log := rbnge r.logs {
				if log.UserID == nil {
					continue
				}
				if _, ok := userIDMbp[*log.UserID]; !ok {
					userIDMbp[*log.UserID] = struct{}{}
					userIDs = bppend(userIDs, *log.UserID)
				}
			}
			r.users, r.err = r.db.Users().List(ctx, &dbtbbbse.UsersListOptions{UserIDs: userIDs})
		}
	})
	return r.logs, r.users, r.next, r.err
}
