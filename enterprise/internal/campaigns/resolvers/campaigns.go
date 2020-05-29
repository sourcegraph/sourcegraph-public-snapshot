package resolvers

import (
	"context"
	"fmt"
	"path"
	"sync"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
)

var _ graphqlbackend.CampaignsConnectionResolver = &campaignsConnectionResolver{}

type campaignsConnectionResolver struct {
	store *ee.Store
	opts  ee.ListCampaignsOpts

	// cache results because they are used by multiple fields
	once      sync.Once
	campaigns []*campaigns.Campaign
	next      int64
	err       error
}

func (r *campaignsConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.CampaignResolver, error) {
	campaigns, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]graphqlbackend.CampaignResolver, 0, len(campaigns))
	for _, c := range campaigns {
		resolvers = append(resolvers, &campaignResolver{store: r.store, Campaign: c})
	}
	return resolvers, nil
}

func (r *campaignsConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	opts := ee.CountCampaignsOpts{ChangesetID: r.opts.ChangesetID, State: r.opts.State, HasPatchSet: r.opts.HasPatchSet}
	count, err := r.store.CountCampaigns(ctx, opts)
	return int32(count), err
}

func (r *campaignsConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	return graphqlutil.HasNextPage(next != 0), nil
}

func (r *campaignsConnectionResolver) compute(ctx context.Context) ([]*campaigns.Campaign, int64, error) {
	r.once.Do(func() {
		r.campaigns, r.next, r.err = r.store.ListCampaigns(ctx, r.opts)
	})
	return r.campaigns, r.next, r.err
}

var _ graphqlbackend.CampaignResolver = &campaignResolver{}

type campaignResolver struct {
	store *ee.Store
	*campaigns.Campaign
}

func (r *campaignResolver) ID() graphql.ID {
	return campaigns.MarshalCampaignID(r.Campaign.ID)
}

func (r *campaignResolver) Name() string {
	return r.Campaign.Name
}

func (r *campaignResolver) Description() *string {
	if r.Campaign.Description == "" {
		return nil
	}
	return &r.Campaign.Description
}

func (r *campaignResolver) Branch() *string {
	if r.Campaign.Branch == "" {
		return nil
	}
	return &r.Campaign.Branch
}

func (r *campaignResolver) Author(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.AuthorID)
}

func (r *campaignResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return currentUserCanAdministerCampaign(ctx, r.Campaign)
}

func (r *campaignResolver) URL(ctx context.Context) (string, error) {
	return path.Join("/campaigns", string(r.ID())), nil
}

func (r *campaignResolver) Namespace(ctx context.Context) (n graphqlbackend.NamespaceResolver, err error) {
	if r.NamespaceUserID != 0 {
		n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.NamespaceUserID)
	} else {
		n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.NamespaceOrgID)
	}

	return n, err
}

func (r *campaignResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.CreatedAt}
}

func (r *campaignResolver) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.Campaign.UpdatedAt}
}

func (r *campaignResolver) ClosedAt() *graphqlbackend.DateTime {
	if r.Campaign.ClosedAt.IsZero() {
		return nil
	}
	return &graphqlbackend.DateTime{Time: r.Campaign.ClosedAt}
}

func (r *campaignResolver) PublishedAt(ctx context.Context) (*graphqlbackend.DateTime, error) {
	if r.Campaign.PatchSetID == 0 {
		return &graphqlbackend.DateTime{Time: r.Campaign.CreatedAt}, nil
	}

	createdAt, err := r.store.GetLatestChangesetJobCreatedAt(ctx, r.Campaign.ID)
	if err != nil {
		return nil, err
	}
	if createdAt.IsZero() {
		return nil, nil
	}
	return &graphqlbackend.DateTime{Time: createdAt}, nil
}

func (r *campaignResolver) Changesets(
	ctx context.Context,
	args *graphqlbackend.ListChangesetsArgs,
) (graphqlbackend.ChangesetsConnectionResolver, error) {
	opts, err := listChangesetOptsFromArgs(args)
	if err != nil {
		return nil, err
	}
	opts.CampaignID = r.Campaign.ID
	return &changesetsConnectionResolver{
		store: r.store,
		opts:  opts,
	}, nil
}

func (r *campaignResolver) OpenChangesets(ctx context.Context) (graphqlbackend.ChangesetsConnectionResolver, error) {
	state := campaigns.ChangesetStateOpen
	return &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID:    r.Campaign.ID,
			ExternalState: &state,
			Limit:         -1,
		},
	}, nil
}

func (r *campaignResolver) Patches(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) graphqlbackend.PatchConnectionResolver {
	if r.Campaign.PatchSetID == 0 {
		return &emptyPatchConnectionResolver{}
	}

	return &patchesConnectionResolver{
		store: r.store,
		opts: ee.ListPatchesOpts{
			PatchSetID:                r.Campaign.PatchSetID,
			Limit:                     int(args.GetFirst()),
			OnlyWithDiff:              true,
			OnlyUnpublishedInCampaign: r.Campaign.ID,
		},
	}
}

func (r *campaignResolver) ChangesetCountsOverTime(
	ctx context.Context,
	args *graphqlbackend.ChangesetCountsArgs,
) ([]graphqlbackend.ChangesetCountsResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	resolvers := []graphqlbackend.ChangesetCountsResolver{}

	opts := ee.ListChangesetsOpts{CampaignID: r.Campaign.ID, Limit: -1}
	cs, _, err := r.store.ListChangesets(ctx, opts)
	if err != nil {
		return resolvers, err
	}

	weekAgo := time.Now().Add(-7 * 24 * time.Hour)
	start := r.Campaign.CreatedAt.UTC()
	if start.After(weekAgo) {
		start = weekAgo
	}
	if args.From != nil {
		start = args.From.Time.UTC()
	}

	end := time.Now().UTC()
	if args.To != nil && args.To.Time.Before(end) {
		end = args.To.Time.UTC()
	}

	changesetIDs := make([]int64, len(cs))
	for i, c := range cs {
		changesetIDs[i] = c.ID
	}

	eventsOpts := ee.ListChangesetEventsOpts{
		ChangesetIDs: changesetIDs,
		Limit:        -1,
	}
	es, _, err := r.store.ListChangesetEvents(ctx, eventsOpts)
	if err != nil {
		return resolvers, err
	}

	counts, err := ee.CalcCounts(start, end, cs, es...)
	if err != nil {
		return resolvers, err
	}

	for _, c := range counts {
		resolvers = append(resolvers, &changesetCountsResolver{counts: c})
	}

	return resolvers, nil
}

func (r *campaignResolver) PatchSet(ctx context.Context) (graphqlbackend.PatchSetResolver, error) {
	if r.Campaign.PatchSetID == 0 {
		return nil, nil
	}

	patchSet, err := r.store.GetPatchSet(ctx, ee.GetPatchSetOpts{ID: r.Campaign.PatchSetID})
	if err != nil {
		return nil, err
	}

	return &patchSetResolver{store: r.store, patchSet: patchSet}, nil
}

func (r *campaignResolver) RepositoryDiffs(
	ctx context.Context,
	args *graphqlutil.ConnectionArgs,
) (graphqlbackend.RepositoryComparisonConnectionResolver, error) {
	changesetsConnection := &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      int(args.GetFirst()),
		},
	}
	return &changesetDiffsConnectionResolver{changesetsConnection}, nil
}

func (r *campaignResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	changesetsConnection := &changesetsConnectionResolver{
		store: r.store,
		opts: ee.ListChangesetsOpts{
			CampaignID: r.Campaign.ID,
			Limit:      -1, // Get all changesets
		},
	}

	changesetDiffs := &changesetDiffsConnectionResolver{changesetsConnection}
	repoComparisons, err := changesetDiffs.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	totalStat := &graphqlbackend.DiffStat{}

	for _, repoComp := range repoComparisons {
		fileDiffs := repoComp.FileDiffs(&graphqlbackend.FileDiffsConnectionArgs{})
		s, err := fileDiffs.DiffStat(ctx)
		if err != nil {
			return nil, err
		}
		totalStat.AddDiffStat(s)
	}

	// We don't have a patch set, so we don't have patches and can return
	if r.Campaign.PatchSetID == 0 {
		return totalStat, nil
	}

	patchSetStat, err := patchSetDiffStat(ctx, r.store, ee.ListPatchesOpts{
		PatchSetID:                r.Campaign.PatchSetID,
		Limit:                     -1, // Fetch all patches in a patch set
		OnlyWithDiff:              true,
		OnlyUnpublishedInCampaign: r.Campaign.ID,
	})
	if err != nil {
		return nil, err
	}
	totalStat.AddDiffStat(patchSetStat)

	return totalStat, nil
}

func (r *campaignResolver) Status(ctx context.Context) (graphqlbackend.BackgroundProcessStatus, error) {
	canAdmin, err := currentUserCanAdministerCampaign(ctx, r.Campaign)
	if err != nil {
		return nil, err
	}

	if !canAdmin {
		// If the user doesn't have admin permissions for this campaign, we
		// don't need to filter out specific errors, but can simply exclude
		// _all_ errors.
		return r.store.GetCampaignStatus(ctx, ee.GetCampaignStatusOpts{
			ID:            r.Campaign.ID,
			ExcludeErrors: true,
		})
	}

	// TODO: Wow, this is horrible. We're loading way too many patches.
	// What we actually want is this:

	//   SELECT repo.id
	//   FROM   patches
	//   JOIN   repo           ON repo.id = patches.repo_id
	//   JOIN   changeset_jobs ON changeset_jobs.patch_id = patches.id
	//   WHERE patches.patch_set_id = patch.patch_set_id;

	// Or we do:
	//
	//   SELECT repo.id
	//   FROM   changeset_jobs
	//   JOIN   patches ON patches.id = changeset_jobs.patch_id
	//   JOIN   repo    ON patches.repo_id = repo.id
	//   WHERE  changeset_jobs.campaign_id = <campaign_id>

	// And then we put those repo IDs through `db.Repo.GetByIDs`, which
	// uses the authz filter what we then get back are the repos we have access
	// to.

	// And then we need to filter out the error messages of changeset_jobs that
	// are attached to patches that are attached to filtered out repositories in the
	// `GetCampaignStatus` query below.

	patches, _, err := r.store.ListPatches(ctx, ee.ListPatchesOpts{
		PatchSetID: r.Campaign.PatchSetID,
		Limit:      -1,
	})
	if err != nil {
		return nil, err
	}

	repoIDs := make([]api.RepoID, 0, len(patches))
	for _, p := range patches {
		repoIDs = append(repoIDs, p.RepoID)
	}

	// 🚨 SECURITY: We use db.Repos.GetByIDs to filter out repositories the
	// user doesn't have access to.
	accessibleRepos, err := db.Repos.GetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	accessibleRepoIDs := make(map[api.RepoID]struct{}, len(accessibleRepos))
	for _, r := range accessibleRepos {
		accessibleRepoIDs[r.ID] = struct{}{}
	}

	// We now check which repositories in `repoIDs` are not in `accessibleRepoIDs`.
	// We have to filter the error messages associated with those out.
	excludedRepos := make([]api.RepoID, 0, len(accessibleRepoIDs))
	for _, id := range repoIDs {
		if _, ok := accessibleRepoIDs[id]; !ok {
			excludedRepos = append(excludedRepos, id)
		}
	}

	return r.store.GetCampaignStatus(ctx, ee.GetCampaignStatusOpts{
		ID:                   r.Campaign.ID,
		ExcludeErrorsInRepos: excludedRepos,
	})
}

type changesetDiffsConnectionResolver struct {
	*changesetsConnectionResolver
}

func (r *changesetDiffsConnectionResolver) Nodes(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	changesets, err := r.changesetsConnectionResolver.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := make([]*graphqlbackend.RepositoryComparisonResolver, 0, len(changesets))
	for _, c := range changesets {
		changesetResolver, ok := c.(*changesetResolver)
		if !ok {
			return nil, fmt.Errorf("not a changeset resolver: %T", c)
		}

		comp, err := changesetResolver.Diff(ctx)
		if err != nil {
			return nil, err
		}
		if comp != nil {
			resolvers = append(resolvers, comp)
		}
	}
	return resolvers, nil
}

type emptyPatchConnectionResolver struct{}

func (r *emptyPatchConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.PatchInterfaceResolver, error) {
	return []graphqlbackend.PatchInterfaceResolver{}, nil
}

func (r *emptyPatchConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return 0, nil
}

func (r *emptyPatchConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil
}
