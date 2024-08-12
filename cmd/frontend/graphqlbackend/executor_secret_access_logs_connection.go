package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type executorSecretAccessLogConnectionResolver struct {
	db   database.DB
	opts database.ExecutorSecretAccessLogsListOpts

	computeOnce sync.Once
	logs        []*database.ExecutorSecretAccessLog
	users       []*types.User
	next        int
	err         error
}

func (r *executorSecretAccessLogConnectionResolver) Nodes(ctx context.Context) ([]*executorSecretAccessLogResolver, error) {
	logs, users, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	userMap := make(map[int32]*types.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	resolvers := make([]*executorSecretAccessLogResolver, 0, len(logs))
	for _, log := range logs {
		r := &executorSecretAccessLogResolver{
			db:                   r.db,
			log:                  log,
			attemptPreloadedUser: true,
		}
		if log.UserID != nil {
			if user, ok := userMap[*log.UserID]; ok {
				r.preloadedUser = user
			}
		}
		resolvers = append(resolvers, r)
	}

	return resolvers, nil
}

func (r *executorSecretAccessLogConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	totalCount, err := r.db.ExecutorSecretAccessLogs().Count(ctx, r.opts)
	return int32(totalCount), err
}

func (r *executorSecretAccessLogConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	_, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		n := int32(next)
		return gqlutil.EncodeIntCursor(&n), nil
	}
	return gqlutil.HasNextPage(false), nil
}

func (r *executorSecretAccessLogConnectionResolver) compute(ctx context.Context) (_ []*database.ExecutorSecretAccessLog, _ []*types.User, next int, err error) {
	r.computeOnce.Do(func() {
		r.logs, r.next, r.err = r.db.ExecutorSecretAccessLogs().List(ctx, r.opts)
		if r.err != nil {
			return
		}
		if len(r.logs) > 0 {
			userIDMap := make(map[int32]struct{})
			userIDs := []int32{}
			for _, log := range r.logs {
				if log.UserID == nil {
					continue
				}
				if _, ok := userIDMap[*log.UserID]; !ok {
					userIDMap[*log.UserID] = struct{}{}
					userIDs = append(userIDs, *log.UserID)
				}
			}
			r.users, r.err = r.db.Users().List(ctx, &database.UsersListOptions{UserIDs: userIDs})
		}
	})
	return r.logs, r.users, r.next, r.err
}
