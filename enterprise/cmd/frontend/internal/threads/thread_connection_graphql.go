package threads

import (
	"context"
	"strings"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func (GraphQLResolver) Threads(ctx context.Context, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	opt, err := threadConnectionArgsToListOptions(ctx, arg)
	if err != nil {
		return nil, err
	}
	return &threadConnection{opt: opt}, nil
}

func (GraphQLResolver) ThreadsForRepository(ctx context.Context, repositoryID graphql.ID, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	repo, err := graphqlbackend.RepositoryByID(ctx, repositoryID)
	if err != nil {
		return nil, err
	}
	opt, err := threadConnectionArgsToListOptions(ctx, arg)
	if err != nil {
		return nil, err
	}
	opt.RepositoryIDs = []api.RepoID{repo.DBID()}
	return &threadConnection{opt: opt}, nil
}

func ThreadsByIDs(ctx context.Context, threadIDs []int64, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	opt, err := threadConnectionArgsToListOptions(ctx, arg)
	if err != nil {
		return nil, err
	}
	opt.ThreadIDs = threadIDs
	return &threadConnection{opt: opt}, nil
}

func threadConnectionArgsToListOptions(ctx context.Context, arg *graphqlbackend.ThreadConnectionArgs) (dbThreadsListOptions, error) {
	var opt dbThreadsListOptions
	arg.Set(&opt.LimitOffset)
	if arg.Filters != nil {
		f := *arg.Filters
		if f.States != nil {
			if len(*f.States) != 1 {
				return opt, errors.New("ThreadFilters only supports having exactly 1 state (or null)")
			}
			opt.States = append(opt.States, string((*f.States)[0]))
		}
		if f.Repositories != nil {
			for _, repoGQLID := range *f.Repositories {
				// TODO!(sqs): security check perms
				dbID, err := graphqlbackend.UnmarshalRepositoryID(repoGQLID)
				if err != nil {
					return opt, err
				}
				opt.RepositoryIDs = append(opt.RepositoryIDs, dbID)
			}
		}
		// TODO!(sqs): hacky repo parser
		if f.Query != nil {
			for _, token := range strings.Fields(*f.Query) {
				switch {
				case strings.HasPrefix(token, "repo:"):
					repoName := api.RepoName(token[len("repo:"):])
					// TODO!(sqs): security check perms
					repo, err := backend.Repos.GetByName(ctx, repoName)
					if err != nil {
						return opt, err
					}
					opt.RepositoryIDs = append(opt.RepositoryIDs, repo.ID)
				case strings.HasPrefix(token, "label:"):
					labelName := token[len("label:"):]
					opt.LabelNames = append(opt.LabelNames, labelName)
				case strings.HasPrefix(token, "is:"):
					state := token[len("is:"):]
					switch strings.ToLower(state) {
					case strings.ToLower(string(graphqlbackend.ThreadStateOpen)):
						opt.States = append(opt.States, string(graphqlbackend.ThreadStateOpen))
					case strings.ToLower(string(graphqlbackend.ThreadStateMerged)):
						opt.States = append(opt.States, string(graphqlbackend.ThreadStateMerged))
					case strings.ToLower(string(graphqlbackend.ThreadStateClosed)):
						// Consider merged threads to be closed because that's usually what the user intends.
						opt.States = append(opt.States, string(graphqlbackend.ThreadStateMerged), string(graphqlbackend.ThreadStateClosed))
					}
				default:
					// Treat token as a query.
					if opt.Query != "" {
						opt.Query += " "
					}
					opt.Query += token
				}
			}
		}
	}
	return opt, nil
}

type threadConnection struct {
	opt dbThreadsListOptions

	once    sync.Once
	threads []*DBThread
	err     error
}

func (r *threadConnection) compute(ctx context.Context) ([]*DBThread, error) {
	r.once.Do(func() {
		opt2 := r.opt
		if opt2.LimitOffset != nil {
			tmp := *opt2.LimitOffset
			opt2.LimitOffset = &tmp
			opt2.Limit++ // so we can detect if there is a next page
		}

		r.threads, r.err = dbThreads{}.List(ctx, opt2)
	})
	return r.threads, r.err
}

func (r *threadConnection) Nodes(ctx context.Context) ([]graphqlbackend.Thread, error) {
	dbThreads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if r.opt.LimitOffset != nil && len(dbThreads) > r.opt.LimitOffset.Limit {
		dbThreads = dbThreads[:r.opt.LimitOffset.Limit]
	}
	return toThreads(dbThreads), nil
}

func toThreads(dbThreads []*DBThread) []graphqlbackend.Thread {
	threads := make([]graphqlbackend.Thread, len(dbThreads))
	for i, DBThread := range dbThreads {
		threads[i] = newGQLThread(DBThread)
	}
	return threads
}

func (r *threadConnection) TotalCount(ctx context.Context) (int32, error) {
	count, err := dbThreads{}.Count(ctx, r.opt)
	return int32(count), err
}

func (r *threadConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	threads, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(r.opt.LimitOffset != nil && len(threads) > r.opt.Limit), nil
}

func (r *threadConnection) Filters(ctx context.Context) (graphqlbackend.ThreadConnectionFilters, error) {
	return newThreadConnectionFiltersFromDB(ctx, r.opt)
}
