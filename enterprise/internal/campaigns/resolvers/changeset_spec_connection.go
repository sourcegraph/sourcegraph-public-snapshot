package resolvers

import (
	"context"
	"database/sql"
	"strconv"
	"sync"

	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

var _ graphqlbackend.ChangesetSpecConnectionResolver = &changesetSpecConnectionResolver{}

type changesetSpecConnectionResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	opts           ee.ListChangesetSpecsOpts
	campaignSpecID int64

	// Cache results because they are used by multiple fields
	once           sync.Once
	changesetSpecs campaigns.ChangesetSpecs
	reposByID      map[api.RepoID]*types.Repo
	next           int64
	err            error
}

func (r *changesetSpecConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	count, err := r.store.CountChangesetSpecs(ctx, ee.CountChangesetSpecsOpts{
		CampaignSpecID: r.campaignSpecID,
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

	fetcher := &changesetSpecPreviewer{
		store:          r.store,
		campaignSpecID: r.campaignSpecID,
	}

	resolvers := make([]graphqlbackend.ChangesetSpecResolver, 0, len(changesetSpecs))
	for _, c := range changesetSpecs {
		repo := reposByID[c.RepoID]
		// If it's not in reposByID the repository was filtered out by the
		// authz-filter.
		// In that case we'll set it anyway to nil and changesetSpecResolver
		// will treat it as "hidden".

		resolvers = append(resolvers, NewChangesetSpecResolverWithRepo(r.store, r.httpFactory, repo, c).WithRewirerMappingFetcher(fetcher))
	}

	return resolvers, nil
}

func (r *changesetSpecConnectionResolver) compute(ctx context.Context) (campaigns.ChangesetSpecs, map[api.RepoID]*types.Repo, int64, error) {
	r.once.Do(func() {
		opts := r.opts
		opts.CampaignSpecID = r.campaignSpecID
		r.changesetSpecs, r.next, r.err = r.store.ListChangesetSpecs(ctx, opts)
		if r.err != nil {
			return
		}

		// 🚨 SECURITY: db.Repos.GetRepoIDsSet uses the authzFilter under the hood and
		// filters out repositories that the user doesn't have access to.
		r.reposByID, r.err = db.Repos.GetReposSetByIDs(ctx, r.changesetSpecs.RepoIDs()...)
	})

	return r.changesetSpecs, r.reposByID, r.next, r.err
}

type changesetSpecPreviewer struct {
	store          *ee.Store
	campaignSpecID int64

	mappingOnce sync.Once
	mappingByID map[int64]*ee.RewirerMapping
	mappingErr  error

	campaignOnce sync.Once
	campaign     *campaigns.Campaign
	campaignErr  error
}

func (c *changesetSpecPreviewer) PlanForChangesetSpec(ctx context.Context, changesetSpec *campaigns.ChangesetSpec) (*ee.ReconcilerPlan, error) {
	mapping, err := c.mappingForChangesetSpec(ctx, changesetSpec.ID)
	if err != nil {
		return nil, err
	}
	campaign, err := c.computeCampaign(ctx)
	if err != nil {
		return nil, err
	}
	rewirer := ee.NewChangesetRewirer(ee.RewirerMappings{mapping}, campaign, repos.NewDBStore(c.store.DB(), sql.TxOptions{}))
	changesets, err := rewirer.Rewire(ctx)
	if err != nil {
		return nil, err
	}

	var changeset *campaigns.Changeset
	if len(changesets) != 1 {
		return nil, errors.New("rewirer did not return changeset")
	} else {
		changeset = changesets[0]
	}

	// Detached changesets would still appear here, but since they'll never match one of the new specs, they don't actually appear here.
	// Once we have a way to have changeset specs for detached changesets, this would be the place to do a "will be detached" check.
	// TBD: How we represent that in the API.

	var previousSpec, currentSpec *campaigns.ChangesetSpec
	if changeset.PreviousSpecID != 0 {
		previousSpec, err = c.store.GetChangesetSpecByID(ctx, changeset.PreviousSpecID)
	}
	if changeset.CurrentSpecID != 0 {
		currentSpec = changesetSpec
	}
	return ee.DetermineReconcilerPlan(previousSpec, currentSpec, changeset)
}

// ChangesetForChangesetSpec can return nil
func (c *changesetSpecPreviewer) ChangesetForChangesetSpec(ctx context.Context, changesetSpecID int64) (*campaigns.Changeset, error) {
	mapping, err := c.mappingForChangesetSpec(ctx, changesetSpecID)
	if err != nil {
		return nil, err
	}
	return mapping.Changeset, nil
}

func (c *changesetSpecPreviewer) mappingForChangesetSpec(ctx context.Context, id int64) (*ee.RewirerMapping, error) {
	mappingByID, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}
	mapping, ok := mappingByID[id]
	if !ok {
		return nil, errors.New("couldn't find mapping for changeset")
	}

	return mapping, nil
}

func (c *changesetSpecPreviewer) computeCampaign(ctx context.Context) (*campaigns.Campaign, error) {
	c.campaignOnce.Do(func() {
		svc := ee.NewService(c.store, nil)
		campaignSpec, err := c.store.GetCampaignSpec(ctx, ee.GetCampaignSpecOpts{ID: c.campaignSpecID})
		if err != nil {
			c.campaignErr = err
			return
		}
		c.campaign, _, c.campaignErr = svc.ReconcileCampaign(ctx, campaignSpec)
	})
	return c.campaign, c.campaignErr
}

func (c *changesetSpecPreviewer) compute(ctx context.Context) (map[int64]*ee.RewirerMapping, error) {
	c.mappingOnce.Do(func() {
		campaign, err := c.computeCampaign(ctx)
		if err != nil {
			c.mappingErr = err
			return
		}
		mappings, err := c.store.GetRewirerMappings(ctx, ee.GetRewirerMappingsOpts{
			CampaignSpecID: c.campaignSpecID,
			CampaignID:     campaign.ID,
		})
		if err != nil {
			c.mappingErr = err
			return
		}
		if err := mappings.Hydrate(ctx, c.store); err != nil {
			c.mappingErr = err
			return
		}

		c.mappingByID = make(map[int64]*ee.RewirerMapping)
		for _, m := range mappings {
			if m.ChangesetSpecID == 0 {
				continue
			}
			c.mappingByID[m.ChangesetSpecID] = m
		}
	})
	return c.mappingByID, c.mappingErr
}
