package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

var _ graphqlbackend.ChangesetSpecConnectionResolver = &changesetSpecConnectionResolver{}

type changesetSpecConnectionResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	opts ee.ListChangesetSpecsOpts

	// Cache results because they are used by multiple fields
	once           sync.Once
	changesetSpecs []*campaigns.ChangesetSpec
	reposByID      map[api.RepoID]*types.Repo
	next           int64
	err            error
}

func (r *changesetSpecConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountChangesetSpecs(ctx, ee.CountChangesetSpecsOpts{
		CampaignSpecID: r.opts.CampaignSpecID,
	})
	if err != nil {
		return 0, err
	}
	return int32(count), nil
}

func (r *changesetSpecConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if next != 0 {
		// We don't use the RandID for pagination, because we can't paginate database
		// entries based on the RandID.
		return graphqlutil.NextPageCursor(strconv.Itoa(int(next))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *changesetSpecConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetSpecResolver, error) {
	changesetSpecs, reposByID, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := make([]graphqlbackend.ChangesetSpecResolver, 0, len(changesetSpecs))
	for _, c := range changesetSpecs {
		repo, _ := reposByID[c.RepoID]
		// If it's not in reposByID the repository was filtered out by the
		// authz-filter.
		// In that case we'll set it anyway to nil and changesetSpecResolver
		// will treat it as "hidden".

		resolvers = append(resolvers, &changesetSpecResolver{
			store:         r.store,
			httpFactory:   r.httpFactory,
			changesetSpec: c,

			preloadedRepo:        repo,
			attemptedPreloadRepo: true,
			repoCtx:              ctx,
		})
	}

	return resolvers, nil
}

func (r *changesetSpecConnectionResolver) compute(ctx context.Context) ([]*campaigns.ChangesetSpec, map[api.RepoID]*types.Repo, int64, error) {
	r.once.Do(func() {
		r.changesetSpecs, r.next, r.err = r.store.ListChangesetSpecs(ctx, r.opts)
		if r.err != nil {
			return
		}

		repoIDs := make([]api.RepoID, len(r.changesetSpecs))
		for i, c := range r.changesetSpecs {
			repoIDs[i] = c.RepoID
		}

		// 🚨 SECURITY: db.Repos.GetByIDs uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		rs, err := db.Repos.GetByIDs(ctx, repoIDs...)
		if err != nil {
			r.err = err
			return
		}

		r.reposByID = make(map[api.RepoID]*types.Repo, len(rs))
		for _, repo := range rs {
			r.reposByID[repo.ID] = repo
		}
	})

	return r.changesetSpecs, r.reposByID, r.next, r.err
}
