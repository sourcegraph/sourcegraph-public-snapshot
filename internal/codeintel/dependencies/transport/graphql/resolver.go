package graphql

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	livedependencies "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/live"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type resolver struct {
	svc    *dependencies.Service
	db     database.DB
	logger log.Logger
}

func New(db database.DB, logger log.Logger) graphqlbackend.DependenciesResolver {
	return &resolver{
		svc:    livedependencies.GetService(db, livedependencies.NewSyncer()),
		db:     db,
		logger: logger,
	}
}

func (r *resolver) LockfileIndexes(ctx context.Context, args *graphqlbackend.ListLockfileIndexesArgs) (graphqlbackend.LockfileIndexConnectionResolver, error) {
	// 🚨 SECURITY: For now we only allow site admins to query lockfile indexes.
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	p, err := validateArgs(args)
	if err != nil {
		return nil, err
	}

	lockfileIndexes, totalCount, err := r.svc.ListLockfileIndexes(ctx, dependencies.ListLockfileIndexesOpts{
		After: p.after,
		Limit: p.limit,
	})
	if err != nil {
		return nil, err
	}

	repoIDs := make([]api.RepoID, len(lockfileIndexes))
	for i, li := range lockfileIndexes {
		repoIDs[i] = api.RepoID(li.RepositoryID)
	}

	var repos []*types.Repo
	if len(repoIDs) > 0 {
		var err error
		repos, err = backend.NewRepos(r.logger, r.db).List(ctx, database.ReposListOptions{IDs: repoIDs})
		if err != nil {
			return nil, err
		}
	}

	reposByID := make(map[api.RepoID]*types.Repo, len(repos))
	for _, repo := range repos {
		reposByID[repo.ID] = repo
	}

	resolvers := make([]graphqlbackend.LockfileIndexResolver, 0, len(lockfileIndexes))
	for _, li := range lockfileIndexes {
		repo, ok := reposByID[api.RepoID(li.RepositoryID)]
		if !ok {
			return nil, errors.Newf("repository with ID %d not found", li.RepositoryID)
		}

		repoResolver := graphqlbackend.NewRepositoryResolver(r.db, repo)
		commit := graphqlbackend.NewGitCommitResolver(r.db, repoResolver, api.CommitID(li.Commit), nil)
		resolvers = append(resolvers, NewLockfileIndexResolver(li, repoResolver, commit))
	}

	nextOffset := graphqlutil.NextOffset(p.after, len(lockfileIndexes), totalCount)
	lockfileIndexesConnection := NewLockfileIndexConnectionConnection(resolvers, totalCount, nextOffset)

	return lockfileIndexesConnection, nil
}

const DefaultLockfileIndexesLimit = 50

type params struct {
	after int
	limit int
}

func validateArgs(args *graphqlbackend.ListLockfileIndexesArgs) (params, error) {
	var p params
	afterCount, err := graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return p, err
	}
	p.after = afterCount

	limit := DefaultLockfileIndexesLimit
	if args.First != 0 {
		limit = int(args.First)
	}
	p.limit = limit

	return p, nil
}
