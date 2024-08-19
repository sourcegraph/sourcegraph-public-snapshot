package graphqlbackend

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type repositoryContributorsArgs struct {
	RevisionRange *string
	AfterDate     *string
	Path          *string
}

func (r *RepositoryResolver) Contributors(args *struct {
	repositoryContributorsArgs
	gqlutil.ConnectionResolverArgs
}) (*gqlutil.ConnectionResolver[*repositoryContributorResolver], error) {
	var after time.Time
	if args.AfterDate != nil && *args.AfterDate != "" {
		var err error
		after, err = gitdomain.ParseGitDate(*args.AfterDate, time.Now)
		if err != nil {
			return nil, errors.Wrap(err, "failed to parse after date")
		}
	}

	connectionStore := &repositoryContributorConnectionStore{
		db:    r.db,
		args:  &args.repositoryContributorsArgs,
		after: after,
		repo:  r,
	}
	reverse := false
	connectionOptions := gqlutil.ConnectionResolverOptions{
		Reverse: &reverse,
	}
	return gqlutil.NewConnectionResolver[*repositoryContributorResolver](connectionStore, &args.ConnectionResolverArgs, &connectionOptions)
}

type repositoryContributorConnectionStore struct {
	db    database.DB
	args  *repositoryContributorsArgs
	after time.Time

	repo *RepositoryResolver

	// cache result because it is used by multiple fields
	once    sync.Once
	results []*gitdomain.ContributorCount
	err     error
}

func (s *repositoryContributorConnectionStore) MarshalCursor(node *repositoryContributorResolver, _ database.OrderBy) (*string, error) {
	position := strconv.Itoa(node.index)
	return &position, nil
}

func (s *repositoryContributorConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	c, err := strconv.Atoi(cursor)
	if err != nil {
		return nil, err
	}
	return []any{c}, nil
}

func (s *repositoryContributorConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	results, err := s.compute(ctx)
	return int32(len(results)), err
}

func (s *repositoryContributorConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*repositoryContributorResolver, error) {
	results, err := s.compute(ctx)
	if err != nil {
		return nil, err
	}

	var start int
	results, start, err = database.OffsetBasedCursorSlice(results, args)
	if err != nil {
		return nil, err
	}

	resolvers := make([]*repositoryContributorResolver, len(results))
	for i, contributor := range results {
		resolvers[i] = &repositoryContributorResolver{
			db:    s.db,
			name:  contributor.Name,
			email: contributor.Email,
			count: contributor.Count,
			repo:  s.repo,
			args:  *s.args,
			index: start + i,
		}
	}

	return resolvers, nil
}

func (s *repositoryContributorConnectionStore) compute(ctx context.Context) ([]*gitdomain.ContributorCount, error) {
	s.once.Do(func() {
		client := gitserver.NewClient("graphql.repocontributor")
		var opt gitserver.ContributorOptions
		if s.args.RevisionRange != nil {
			opt.Range = *s.args.RevisionRange
		}
		if s.args.Path != nil {
			opt.Path = *s.args.Path
		}
		opt.After = s.after
		s.results, s.err = client.ContributorCount(ctx, s.repo.RepoName(), opt)
	})
	return s.results, s.err
}
