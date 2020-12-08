package resolvers

import (
	"context"
	"database/sql"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

func marshalChangesetSpecRandID(id string) graphql.ID {
	return relay.MarshalID("ChangesetSpec", id)
}

func unmarshalChangesetSpecID(id graphql.ID) (changesetSpecRandID string, err error) {
	err = relay.UnmarshalSpec(id, &changesetSpecRandID)
	return
}

var _ graphqlbackend.ChangesetSpecResolver = &changesetSpecResolver{}

type changesetSpecResolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory

	changesetSpec *campaigns.ChangesetSpec

	previewer ChangesetSpecPreviewer
	repo      *types.Repo

	planOnce sync.Once
	plan     *ee.ReconcilerPlan
	planErr  error
}

func NewChangesetSpecResolver(ctx context.Context, store *ee.Store, cf *httpcli.Factory, changesetSpec *campaigns.ChangesetSpec) (*changesetSpecResolver, error) {
	resolver := &changesetSpecResolver{
		store:         store,
		httpFactory:   cf,
		changesetSpec: changesetSpec,
		previewer: &changesetSpecPreviewer{
			store:          store,
			campaignSpecID: changesetSpec.CampaignSpecID,
		},
	}

	// 🚨 SECURITY: db.Repos.GetByIDs uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	// In case we don't find a repository, it might be because it's deleted
	// or because the user doesn't have access.
	rs, err := db.Repos.GetByIDs(ctx, changesetSpec.RepoID)
	if err != nil {
		return nil, err
	}

	// Not found is ok, the resolver will disguise as a HiddenChangesetResolver.
	if len(rs) == 1 {
		resolver.repo = rs[0]
	}

	return resolver, nil
}

func NewChangesetSpecResolverWithRepo(store *ee.Store, cf *httpcli.Factory, repo *types.Repo, changesetSpec *campaigns.ChangesetSpec) *changesetSpecResolver {
	return &changesetSpecResolver{
		store:         store,
		httpFactory:   cf,
		repo:          repo,
		changesetSpec: changesetSpec,
		previewer: &changesetSpecPreviewer{
			store:          store,
			campaignSpecID: changesetSpec.CampaignSpecID,
		},
	}
}

func (r *changesetSpecResolver) WithPreviewer(previewer ChangesetSpecPreviewer) *changesetSpecResolver {
	r.previewer = previewer
	return r
}

func (r *changesetSpecResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalChangesetSpecRandID(r.changesetSpec.RandID)
}

func (r *changesetSpecResolver) Type() campaigns.ChangesetSpecDescriptionType {
	return r.changesetSpec.Spec.Type()
}

func (r *changesetSpecResolver) Description(ctx context.Context) (graphqlbackend.ChangesetDescription, error) {
	descriptionResolver := &changesetDescriptionResolver{
		desc: r.changesetSpec.Spec,
		// Note: r.repo can never be nil, because Description is a VisibleChangesetSpecResolver-only field.
		repoResolver: graphqlbackend.NewRepositoryResolver(r.repo),
		diffStat:     r.changesetSpec.DiffStat(),
	}

	return descriptionResolver, nil
}

func (r *changesetSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: r.changesetSpec.ExpiresAt()}
}

func (r *changesetSpecResolver) Operations(ctx context.Context) ([]campaigns.ReconcilerOperation, error) {
	if !r.repoAccessible() {
		// If the repo is inaccessible, no operations would be taken, since the changeset is not created/updated.
		return []campaigns.ReconcilerOperation{}, nil
	}
	plan, err := r.computePlan(ctx)
	if err != nil {
		return nil, err
	}
	ops := plan.Ops.ExecutionOrder()
	return ops, nil
}

func (r *changesetSpecResolver) Delta(ctx context.Context) (graphqlbackend.ChangesetSpecDeltaResolver, error) {
	if !r.repoAccessible() {
		// If the repo is inaccessible, no comparison is made, since the changeset is not created/updated.
		return &changesetSpecDeltaResolver{}, nil
	}
	plan, err := r.computePlan(ctx)
	if err != nil {
		return nil, err
	}
	if plan.Delta == nil {
		return &changesetSpecDeltaResolver{}, nil
	}
	return &changesetSpecDeltaResolver{delta: *plan.Delta}, nil
}

func (r *changesetSpecResolver) computePlan(ctx context.Context) (*ee.ReconcilerPlan, error) {
	r.planOnce.Do(func() {
		r.plan, r.planErr = r.previewer.PlanForChangesetSpec(ctx, r.changesetSpec)
	})
	return r.plan, r.planErr
}

func (r *changesetSpecResolver) Changeset(ctx context.Context) (graphqlbackend.ChangesetResolver, error) {
	changeset, err := r.previewer.ChangesetForChangesetSpec(ctx, r.changesetSpec.ID)
	if err != nil {
		return nil, err
	}
	if changeset == nil {
		return nil, nil
	}
	return NewChangesetResolver(r.store, r.httpFactory, changeset, r.repo), nil
}

func (r *changesetSpecResolver) repoAccessible() bool {
	// If the repository is not nil, it's accessible
	return r.repo != nil
}

func (r *changesetSpecResolver) ToHiddenChangesetSpec() (graphqlbackend.HiddenChangesetSpecResolver, bool) {
	if r.repoAccessible() {
		return nil, false
	}

	return r, true
}

func (r *changesetSpecResolver) ToVisibleChangesetSpec() (graphqlbackend.VisibleChangesetSpecResolver, bool) {
	if !r.repoAccessible() {
		return nil, false
	}

	return r, true
}

var _ graphqlbackend.ChangesetDescription = &changesetDescriptionResolver{}

// changesetDescriptionResolver implements both ChangesetDescription
// interfaces: ExistingChangesetReferenceResolver and
// GitBranchChangesetDescriptionResolver.
type changesetDescriptionResolver struct {
	repoResolver *graphqlbackend.RepositoryResolver
	desc         *campaigns.ChangesetSpecDescription
	diffStat     diff.Stat
}

func (r *changesetDescriptionResolver) ToExistingChangesetReference() (graphqlbackend.ExistingChangesetReferenceResolver, bool) {
	if r.desc.IsImportingExisting() {
		return r, true
	}
	return nil, false
}
func (r *changesetDescriptionResolver) ToGitBranchChangesetDescription() (graphqlbackend.GitBranchChangesetDescriptionResolver, bool) {
	if r.desc.IsBranch() {
		return r, true
	}
	return nil, false
}

func (r *changesetDescriptionResolver) BaseRepository() *graphqlbackend.RepositoryResolver {
	return r.repoResolver
}
func (r *changesetDescriptionResolver) ExternalID() string { return r.desc.ExternalID }
func (r *changesetDescriptionResolver) BaseRef() string    { return git.AbbreviateRef(r.desc.BaseRef) }
func (r *changesetDescriptionResolver) BaseRev() string    { return r.desc.BaseRev }
func (r *changesetDescriptionResolver) HeadRepository() *graphqlbackend.RepositoryResolver {
	return r.repoResolver
}
func (r *changesetDescriptionResolver) HeadRef() string { return git.AbbreviateRef(r.desc.HeadRef) }
func (r *changesetDescriptionResolver) Title() string   { return r.desc.Title }
func (r *changesetDescriptionResolver) Body() string    { return r.desc.Body }
func (r *changesetDescriptionResolver) Published() campaigns.PublishedValue {
	return r.desc.Published
}

func (r *changesetDescriptionResolver) DiffStat() *graphqlbackend.DiffStat {
	return graphqlbackend.NewDiffStat(r.diffStat)
}

func (r *changesetDescriptionResolver) Diff(ctx context.Context) (graphqlbackend.PreviewRepositoryComparisonResolver, error) {
	diff, err := r.desc.Diff()
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewPreviewRepositoryComparisonResolver(ctx, r.repoResolver, r.desc.BaseRev, diff)
}

func (r *changesetDescriptionResolver) Commits() []graphqlbackend.GitCommitDescriptionResolver {
	var resolvers []graphqlbackend.GitCommitDescriptionResolver
	for _, c := range r.desc.Commits {
		resolvers = append(resolvers, &gitCommitDescriptionResolver{
			message:     c.Message,
			diff:        c.Diff,
			authorName:  c.AuthorName,
			authorEmail: c.AuthorEmail,
		})
	}
	return resolvers
}

var _ graphqlbackend.GitCommitDescriptionResolver = &gitCommitDescriptionResolver{}

type gitCommitDescriptionResolver struct {
	message     string
	diff        string
	authorName  string
	authorEmail string
}

func (r *gitCommitDescriptionResolver) Author() *graphqlbackend.PersonResolver {
	return graphqlbackend.NewPersonResolver(
		r.authorName,
		r.authorEmail,
		// Try to find the corresponding Sourcegraph user.
		true,
	)
}
func (r *gitCommitDescriptionResolver) Message() string { return r.message }
func (r *gitCommitDescriptionResolver) Subject() string {
	return graphqlbackend.GitCommitSubject(r.message)
}
func (r *gitCommitDescriptionResolver) Body() *string {
	body := graphqlbackend.GitCommitBody(r.message)
	if body == "" {
		return nil
	}
	return &body
}
func (r *gitCommitDescriptionResolver) Diff() string { return r.diff }

type changesetSpecDeltaResolver struct {
	delta ee.ChangesetSpecDelta
}

var _ graphqlbackend.ChangesetSpecDeltaResolver = &changesetSpecDeltaResolver{}

func (c *changesetSpecDeltaResolver) TitleChanged() bool {
	return c.delta.TitleChanged
}
func (c *changesetSpecDeltaResolver) BodyChanged() bool {
	return c.delta.BodyChanged
}
func (c *changesetSpecDeltaResolver) Undraft() bool {
	return c.delta.Undraft
}
func (c *changesetSpecDeltaResolver) BaseRefChanged() bool {
	return c.delta.BaseRefChanged
}
func (c *changesetSpecDeltaResolver) DiffChanged() bool {
	return c.delta.DiffChanged
}
func (c *changesetSpecDeltaResolver) CommitMessageChanged() bool {
	return c.delta.CommitMessageChanged
}
func (c *changesetSpecDeltaResolver) AuthorNameChanged() bool {
	return c.delta.AuthorNameChanged
}
func (c *changesetSpecDeltaResolver) AuthorEmailChanged() bool {
	return c.delta.AuthorEmailChanged
}

type ChangesetSpecPreviewer interface {
	PlanForChangesetSpec(context.Context, *campaigns.ChangesetSpec) (*ee.ReconcilerPlan, error)
	ChangesetForChangesetSpec(context.Context, int64) (*campaigns.Changeset, error)
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

// Type guard.
var _ ChangesetSpecPreviewer = &changesetSpecPreviewer{}

// PlanForChangesetSpec computes the ReconcilerPlan for the given changeset spec, based on the current state of the database.
func (c *changesetSpecPreviewer) PlanForChangesetSpec(ctx context.Context, changesetSpec *campaigns.ChangesetSpec) (*ee.ReconcilerPlan, error) {
	// To get the plan, we first need to determine the changeset_spec => changeset mapping.
	mapping, err := c.mappingForChangesetSpec(ctx, changesetSpec.ID)
	if err != nil {
		return nil, err
	}
	campaign, err := c.computeCampaign(ctx)
	if err != nil {
		return nil, err
	}
	// And then dry-run the rewirer to simulate how the changeset would look like after an _apply_ operation.
	rewirer := ee.NewChangesetRewirer(ee.RewirerMappings{mapping}, campaign, repos.NewDBStore(c.store.DB(), sql.TxOptions{}))
	changesets, err := rewirer.Rewire(ctx)
	if err != nil {
		return nil, err
	}

	if len(changesets) != 1 {
		return nil, errors.New("rewirer did not return changeset")
	}
	changeset := changesets[0]

	// Detached changesets would still appear here, but since they'll never match one of the new specs, they don't actually appear here.
	// Once we have a way to have changeset specs for detached changesets, this would be the place to do a "will be detached" check.
	// TBD: How we represent that in the API.

	// The rewirer takes previous and current spec into account to determine actions to take,
	// so we need to find out which specs we need to pass to the planner.
	var previousSpec, currentSpec *campaigns.ChangesetSpec
	if changeset.PreviousSpecID != 0 {
		previousSpec, err = c.store.GetChangesetSpecByID(ctx, changeset.PreviousSpecID)
		if err != nil {
			return nil, err
		}
	}
	if changeset.CurrentSpecID != 0 {
		// If the current spec was not unset by the rewirer, it will be this resolvers spec.
		currentSpec = changesetSpec
	}
	return ee.DetermineReconcilerPlan(previousSpec, currentSpec, changeset)
}

// ChangesetForChangesetSpec returns the changeset for the target changeset of the changeset spec. It can return nil, if no changeset
// exists yet.
func (c *changesetSpecPreviewer) ChangesetForChangesetSpec(ctx context.Context, changesetSpecID int64) (*campaigns.Changeset, error) {
	mapping, err := c.mappingForChangesetSpec(ctx, changesetSpecID)
	if err != nil {
		return nil, err
	}
	return mapping.Changeset, nil
}

func (c *changesetSpecPreviewer) mappingForChangesetSpec(ctx context.Context, id int64) (*ee.RewirerMapping, error) {
	mappingByID, err := c.computeMappings(ctx)
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

func (c *changesetSpecPreviewer) computeMappings(ctx context.Context) (map[int64]*ee.RewirerMapping, error) {
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
