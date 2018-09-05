package graphqlbackend

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/useractivity"
	"github.com/sourcegraph/sourcegraph/pkg/types"
)

func (r *schemaResolver) Users(args *struct {
	connectionArgs
	Query        *string
	Tag          *string
	ActivePeriod *string
}) *userConnectionResolver {
	var opt db.UsersListOptions
	if args.Query != nil {
		opt.Query = *args.Query
	}
	if args.Tag != nil {
		opt.Tag = *args.Tag
	}
	args.connectionArgs.set(&opt.LimitOffset)
	return &userConnectionResolver{opt: opt, activePeriod: args.ActivePeriod}
}

type userConnectionResolver struct {
	opt          db.UsersListOptions
	activePeriod *string

	// cache results because they are used by multiple fields
	once       sync.Once
	users      []*types.User
	totalCount int
	err        error
}

// compute caches results from the more expensive user list creation that occurs when activePeriod
// is set to a specific length of time.
//
// Because user activity data isn't stored in PostgreSQL (but rather in
// Redis), adding this parameter requires accessing a second data store.
func (r *userConnectionResolver) compute(ctx context.Context) ([]*types.User, int, error) {
	if r.activePeriod == nil {
		return nil, 0, errors.New("activePeriod must not be nil")
	}
	if r.activePeriod != nil && envvar.SourcegraphDotComMode() {
		return nil, 0, errors.New("site analytics is not available on sourcegraph.com")
	}
	r.once.Do(func() {
		var userIDs []string
		var err error

		switch *r.activePeriod {
		case "TODAY":
			_, userIDs, _, err = useractivity.ListUsersToday()
		case "THIS_WEEK":
			_, userIDs, _, err = useractivity.ListUsersThisWeek()
		case "THIS_MONTH":
			_, userIDs, _, err = useractivity.ListUsersThisMonth()
		default:
			err = fmt.Errorf("unknown user event %s", *r.activePeriod)
		}
		if err != nil {
			r.err = err
			return
		}

		r.opt.UserIDs, err = sliceAtoi(userIDs)
		if err != nil {
			r.err = err
			return
		}

		r.users, err = db.Users.List(ctx, &r.opt)
		if err != nil {
			r.err = err
			return
		}
		r.totalCount, r.err = db.Users.Count(ctx, r.opt)
	})
	return r.users, r.totalCount, r.err
}

func (r *userConnectionResolver) Nodes(ctx context.Context) ([]*userResolver, error) {
	var users []*types.User
	var err error
	if r.useCache() {
		users, _, err = r.compute(ctx)
	} else {
		users, err = db.Users.List(ctx, &r.opt)
	}
	if err != nil {
		return nil, err
	}

	var l []*userResolver
	for _, user := range users {
		l = append(l, &userResolver{
			user: user,
		})
	}
	return l, nil
}

func (r *userConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	var count int
	var err error
	if r.useCache() {
		_, count, err = r.compute(ctx)
	} else {
		count, err = db.Users.Count(ctx, r.opt)
	}
	return int32(count), err
}

func (r *userConnectionResolver) useCache() bool {
	return r.activePeriod != nil && *r.activePeriod != "ALL_TIME"
}

// staticUserConnectionResolver implements the GraphQL type UserConnection based on an underlying
// list of users that is computed statically.
type staticUserConnectionResolver struct {
	users []*types.User
}

func (r *staticUserConnectionResolver) Nodes() []*userResolver {
	resolvers := make([]*userResolver, len(r.users))
	for i, user := range r.users {
		resolvers[i] = &userResolver{user: user}
	}
	return resolvers
}

func (r *staticUserConnectionResolver) TotalCount() int32 { return int32(len(r.users)) }

func sliceAtoi(sa []string) ([]int32, error) {
	si := make([]int32, 0, len(sa))
	for _, a := range sa {
		i, err := strconv.Atoi(a)
		if err != nil {
			return nil, err
		}
		si = append(si, int32(i))
	}
	return si, nil
}
