package graphqlbackend

import (
	"context"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type repositoryArgs struct {
	graphqlutil.ConnectionArgs
	Query *string // Search query
	Names *[]string

	Cloned     bool
	NotCloned  bool
	Indexed    bool
	NotIndexed bool

	CloneStatus *string
	FailedFetch bool
	Corrupted   bool

	ExternalService *graphql.ID

	OrderBy    string
	Descending bool
	After      *string
}

func (args *repositoryArgs) toReposListOptions() (database.ReposListOptions, error) {
	opt := database.ReposListOptions{
		OrderBy: database.RepoListOrderBy{{
			Field:      ToDBRepoListColumn(args.OrderBy),
			Descending: args.Descending,
		}},
	}
	if args.Names != nil {
		opt.Names = *args.Names
	}
	if args.Query != nil {
		opt.Query = *args.Query
	}
	if args.After != nil {
		cursor, err := UnmarshalRepositoryCursor(args.After)
		if err != nil {
			return opt, err
		}
		opt.Cursors = append(opt.Cursors, cursor)
	} else {
		cursor := types.Cursor{
			Column: string(ToDBRepoListColumn(args.OrderBy)),
		}

		if args.Descending {
			cursor.Direction = "prev"
		} else {
			cursor.Direction = "next"
		}

		opt.Cursors = append(opt.Cursors, &cursor)
	}
	args.Set(&opt.LimitOffset)

	if args.CloneStatus != nil {
		opt.CloneStatus = types.ParseCloneStatusFromGraphQL(*args.CloneStatus)
	}

	opt.FailedFetch = args.FailedFetch
	opt.OnlyCorrupted = args.Corrupted

	if !args.Cloned && !args.NotCloned {
		return database.ReposListOptions{}, errors.New("excluding cloned and not cloned repos leaves an empty set")
	}
	if !args.Cloned {
		opt.NoCloned = true
	}
	if !args.NotCloned {
		// notCloned is true by default.
		// this condition is valid only if it has been
		// explicitly set to false by the client.
		opt.OnlyCloned = true
	}

	if !args.Indexed && !args.NotIndexed {
		return database.ReposListOptions{}, errors.New("excluding indexed and not indexed repos leaves an empty set")
	}
	if !args.Indexed {
		opt.NoIndexed = true
	}
	if !args.NotIndexed {
		opt.OnlyIndexed = true
	}

	if args.ExternalService != nil {
		extSvcID, err := UnmarshalExternalServiceID(*args.ExternalService)
		if err != nil {
			return opt, err
		}
		opt.ExternalServiceIDs = append(opt.ExternalServiceIDs, extSvcID)
	}

	return opt, nil
}

func (r *schemaResolver) Repositories(args *repositoryArgs) (*repositoryConnectionResolver, error) {
	opt, err := args.toReposListOptions()

	if err != nil {
		return nil, err
	}

	return &repositoryConnectionResolver{
		db:         r.db,
		logger:     r.logger.Scoped("repositoryConnectionResolver", "resolves connections to a repository"),
		opt:        opt,
		indexed:    args.Indexed,
		notIndexed: args.NotIndexed,
	}, nil
}

type TotalCountArgs struct {
	Precise bool
}

type RepositoryConnectionResolver interface {
	Nodes(ctx context.Context) ([]*RepositoryResolver, error)
	TotalCount(ctx context.Context, args *TotalCountArgs) (*int32, error)
	PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error)
}

var _ RepositoryConnectionResolver = &repositoryConnectionResolver{}

type repositoryConnectionResolver struct {
	logger     log.Logger
	db         database.DB
	opt        database.ReposListOptions
	indexed    bool
	notIndexed bool

	// cache results because they are used by multiple fields
	once  sync.Once
	repos []*types.Repo
	err   error
}

func (r *repositoryConnectionResolver) compute(ctx context.Context) ([]*types.Repo, error) {
	r.once.Do(func() {
		opt2 := r.opt

		if envvar.SourcegraphDotComMode() {
			// 🚨 SECURITY: Don't allow non-admins to perform huge queries on Sourcegraph.com.
			if isSiteAdmin := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db) == nil; !isSiteAdmin {
				if opt2.LimitOffset == nil {
					opt2.LimitOffset = &database.LimitOffset{Limit: 1000}
				}
			}
		}

		reposClient := backend.NewRepos(r.logger, r.db, gitserver.NewClient(r.db))
		for {
			// Cursor-based pagination requires that we fetch limit+1 records, so
			// that we know whether or not there's an additional page (or more)
			// beyond the current one. We reset the limit immediately afterward for
			// any subsequent calculations.
			if opt2.LimitOffset != nil {
				opt2.LimitOffset.Limit++
			}
			repos, err := reposClient.List(ctx, opt2)
			if err != nil {
				r.err = err
				return
			}
			if opt2.LimitOffset != nil {
				opt2.LimitOffset.Limit--
			}
			reposFromDB := len(repos)

			r.repos = append(r.repos, repos...)

			if opt2.LimitOffset == nil {
				break
			} else {
				// check if we filtered some repos and if we need to get more from the DB
				if len(repos) >= opt2.Limit || reposFromDB < opt2.Limit {
					break
				}
				opt2.Offset += opt2.Limit
			}
		}
	})

	return r.repos, r.err
}

func (r *repositoryConnectionResolver) Nodes(ctx context.Context) ([]*RepositoryResolver, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*RepositoryResolver, 0, len(repos))
	client := gitserver.NewClient(r.db)
	for i, repo := range repos {
		if r.opt.LimitOffset != nil && i == r.opt.Limit {
			break
		}

		resolvers = append(resolvers, NewRepositoryResolver(r.db, client, repo))
	}
	return resolvers, nil
}

func (r *repositoryConnectionResolver) TotalCount(ctx context.Context, args *TotalCountArgs) (countptr *int32, err error) {
	if r.opt.UserID != 0 {
		// 🚨 SECURITY: If filtering by user, restrict to that user
		if err := auth.CheckSameUser(ctx, r.opt.UserID); err != nil {
			return nil, err
		}
	} else if r.opt.OrgID != 0 {
		if err := auth.CheckOrgAccess(ctx, r.db, r.opt.OrgID); err != nil {
			return nil, err
		}
	} else {
		// 🚨 SECURITY: Only site admins can list all repos, because a total repository
		// count does not respect repository permissions.
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, err
		}
	}

	// Counting repositories is slow on Sourcegraph.com. Don't wait very long for an exact count.
	if !args.Precise && envvar.SourcegraphDotComMode() {
		if len(r.opt.Query) < 4 {
			return nil, nil
		}

		var cancel func()
		ctx, cancel = context.WithTimeout(ctx, 300*time.Millisecond)
		defer cancel()
		defer func() {
			if ctx.Err() == context.DeadlineExceeded {
				countptr = nil
				err = nil
			}
		}()
	}

	i32ptr := func(v int32) *int32 { return &v }
	count, err := r.db.Repos().Count(ctx, r.opt)
	return i32ptr(int32(count)), err
}

func (r *repositoryConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	repos, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if len(repos) == 0 || r.opt.LimitOffset == nil || len(repos) <= r.opt.Limit || len(r.opt.Cursors) == 0 {
		return graphqlutil.HasNextPage(false), nil
	}

	cursor := r.opt.Cursors[0]

	var value string
	switch cursor.Column {
	case string(database.RepoListName):
		value = string(repos[len(repos)-1].Name)
	case string(database.RepoListCreatedAt):
		value = repos[len(repos)-1].CreatedAt.Format("2006-01-02 15:04:05.999999")
	}
	return graphqlutil.NextPageCursor(MarshalRepositoryCursor(
		&types.Cursor{
			Column:    cursor.Column,
			Value:     value,
			Direction: cursor.Direction,
		},
	)), nil
}

func ToDBRepoListColumn(ob string) database.RepoListColumn {
	switch ob {
	case "REPO_URI", "REPOSITORY_NAME":
		return database.RepoListName
	case "REPO_CREATED_AT", "REPOSITORY_CREATED_AT":
		return database.RepoListCreatedAt
	case "SIZE":
		return database.RepoListSize
	default:
		return ""
	}
}
