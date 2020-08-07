package resolvers

import (
	"context"
	"sync"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/types"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type changesetsConnectionResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	opts ee.ListChangesetsOpts
	// 🚨 SECURITY: If the given opts do not reveal hidden information about a
	// changeset by including the changeset in the result set, this should be
	// set to true.
	optsSafe bool

	// cache results because they are used by multiple fields
	once       sync.Once
	changesets campaigns.Changesets
	reposByID  map[api.RepoID]*types.Repo
	err        error

	// allAccessibleChangesets contains all changesets in this connection,
	// without any pagination.
	// We need them for TotalCount and Stats and we need to load all, without a
	// limit, because some might be filtered out by the authzFilter.
	//
	// NOTE: In the future, as an optimization, we can combine this with
	// `changesets`, since changesets is a subset of `allAccessibleChangesets`.
	allAccessibleChangesetsOnce sync.Once
	allAccessibleChangesets     campaigns.Changesets
	allAccessibleChangesetsErr  error
}

func (r *changesetsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.ChangesetResolver, error) {
	changesets, reposByID, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	syncData, err := r.store.ListChangesetSyncData(ctx, ee.ListChangesetSyncDataOpts{ChangesetIDs: changesets.IDs()})
	if err != nil {
		return nil, err
	}
	scheduledSyncs := make(map[int64]time.Time)
	for _, d := range syncData {
		scheduledSyncs[d.ChangesetID] = ee.NextSync(time.Now, d)
	}

	resolvers := make([]graphqlbackend.ChangesetResolver, 0)
	for _, c := range changesets {
		nextSyncAt, isPreloaded := scheduledSyncs[c.ID]
		var preloadedNextSyncAt *time.Time
		if isPreloaded {
			preloadedNextSyncAt = &nextSyncAt
		}

		repo, repoFound := reposByID[c.RepoID]
		// If it's not in reposByID the repository was filtered out by the
		// authz-filter. In that case we want to return a changesetResolver
		// that doesn't reveal all information.
		// But if the filter opts would leak information about the hidden
		// changesets, we skip the hidden changeset.
		if !repoFound && !r.optsSafe {
			continue
		}

		resolvers = append(resolvers, &changesetResolver{
			store:                      r.store,
			httpFactory:                r.httpFactory,
			changeset:                  c,
			preloadedRepo:              repo,
			attemptedPreloadRepo:       true,
			attemptedPreloadNextSyncAt: true,
			preloadedNextSyncAt:        preloadedNextSyncAt,
		})
	}

	return resolvers, nil
}

func (r *changesetsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	cs, err := r.computeAllAccessibleChangesets(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(cs)), nil
}

func (r *changesetsConnectionResolver) Stats(ctx context.Context) (graphqlbackend.ChangesetsConnectionStatsResolver, error) {
	cs, err := r.computeAllAccessibleChangesets(ctx)
	if err != nil {
		return nil, err
	}
	return newChangesetConnectionStats(cs), nil
}

// computeAllChangesets loads all changesets matched by r.opts, but without a
// limit.
// If r.optsSafe is true, it returns all of them. If not, it filters out the
// ones to which the user doesn't have access.
func (r *changesetsConnectionResolver) computeAllAccessibleChangesets(ctx context.Context) (campaigns.Changesets, error) {
	r.allAccessibleChangesetsOnce.Do(func() {
		opts := r.opts
		opts.Limit = -1

		cs, _, err := r.store.ListChangesets(ctx, opts)
		if err != nil {
			r.allAccessibleChangesetsErr = err
			return
		}

		// 🚨 SECURITY: If the opts do not leak information, we can return the
		// number of changesets. Otherwise we have to filter the changesets by
		// accessible repos.
		if r.optsSafe {
			r.allAccessibleChangesets = cs
			return
		}

		// 🚨 SECURITY: db.Repos.GetByIDs uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		rs, err := db.Repos.GetByIDs(ctx, cs.RepoIDs()...)
		if err != nil {
			r.allAccessibleChangesetsErr = err
			return
		}

		accessibleRepoIDs := map[api.RepoID]struct{}{}
		for _, r := range rs {
			accessibleRepoIDs[r.ID] = struct{}{}
		}

		var accessibleChangesets []*campaigns.Changeset
		for _, c := range cs {
			if _, ok := accessibleRepoIDs[c.RepoID]; !ok {
				continue
			}
			accessibleChangesets = append(accessibleChangesets, c)
		}

		r.allAccessibleChangesets = accessibleChangesets
	})

	return r.allAccessibleChangesets, r.allAccessibleChangesetsErr
}

func (r *changesetsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	page, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	all, err := r.computeAllAccessibleChangesets(ctx)
	if err != nil {
		return nil, err
	}

	return graphqlutil.HasNextPage(len(page) < len(all)), nil
}

func (r *changesetsConnectionResolver) compute(ctx context.Context) (campaigns.Changesets, map[api.RepoID]*types.Repo, error) {
	r.once.Do(func() {
		r.changesets, _, r.err = r.store.ListChangesets(ctx, r.opts)
		if r.err != nil {
			return
		}

		repoIDs := r.changesets.RepoIDs()

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

	return r.changesets, r.reposByID, r.err
}

func newChangesetConnectionStats(cs []*campaigns.Changeset) *changesetsConnectionStatsResolver {
	stats := &changesetsConnectionStatsResolver{
		total: int32(len(cs)),
	}

	for _, c := range cs {
		if c.PublicationState.Unpublished() {
			stats.unpublished++
			continue
		}

		switch c.ExternalState {
		case campaigns.ChangesetExternalStateClosed:
			stats.closed++
		case campaigns.ChangesetExternalStateMerged:
			stats.merged++
		case campaigns.ChangesetExternalStateOpen:
			stats.open++
		}
	}

	return stats
}

type changesetsConnectionStatsResolver struct {
	unpublished, open, merged, closed, total int32
}

func (r *changesetsConnectionStatsResolver) Unpublished() int32 {
	return r.unpublished
}
func (r *changesetsConnectionStatsResolver) Open() int32 {
	return r.open
}
func (r *changesetsConnectionStatsResolver) Merged() int32 {
	return r.merged
}
func (r *changesetsConnectionStatsResolver) Closed() int32 {
	return r.closed
}
func (r *changesetsConnectionStatsResolver) Total() int32 {
	return r.total
}
