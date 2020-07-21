package resolvers

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	ee "github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/db"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

var ErrIDIsZero = errors.New("invalid node id")

// Resolver is the GraphQL resolver of all things related to Campaigns.
type Resolver struct {
	store       *ee.Store
	httpFactory *httpcli.Factory
}

// NewResolver returns a new Resolver whose store uses the given db
func NewResolver(db *sql.DB) graphqlbackend.CampaignsResolver {
	return &Resolver{store: ee.NewStore(db)}
}

func allowReadAccess(ctx context.Context) error {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if readAccess := conf.CampaignsReadAccessEnabled(); readAccess {
		return nil
	}

	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) ChangesetByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access changesets.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	changesetID, err := unmarshalChangesetID(id)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, nil
	}

	changeset, err := r.store.GetChangeset(ctx, ee.GetChangesetOpts{ID: changesetID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: db.Repos.Get uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	repo, err := db.Repos.Get(ctx, changeset.RepoID)
	if err != nil && !errcode.IsNotFound(err) {
		return nil, err
	}

	return &changesetResolver{
		store:                r.store,
		httpFactory:          r.httpFactory,
		changeset:            changeset,
		attemptedPreloadRepo: true,
		preloadedRepo:        repo,
	}, nil
}

func (r *Resolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	campaignID, err := campaigns.UnmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, nil
	}

	campaign, err := r.store.GetCampaign(ctx, ee.GetCampaignOpts{ID: campaignID})
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) CampaignSpecByID(ctx context.Context, id graphql.ID) (graphqlbackend.CampaignSpecResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	campaignSpecRandID, err := unmarshalCampaignSpecID(id)
	if err != nil {
		return nil, err
	}

	if campaignSpecRandID == "" {
		return nil, nil
	}

	opts := ee.GetCampaignSpecOpts{RandID: campaignSpecRandID}
	campaignSpec, err := r.store.GetCampaignSpec(ctx, opts)
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &campaignSpecResolver{store: r.store, httpFactory: r.httpFactory, campaignSpec: campaignSpec}, nil
}

func (r *Resolver) ChangesetSpecByID(ctx context.Context, id graphql.ID) (graphqlbackend.ChangesetSpecResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}

	changesetSpecRandID, err := unmarshalChangesetSpecID(id)
	if err != nil {
		return nil, err
	}

	if changesetSpecRandID == "" {
		return nil, nil
	}

	opts := ee.GetChangesetSpecOpts{RandID: changesetSpecRandID}
	changesetSpec, err := r.store.GetChangesetSpec(ctx, opts)
	if err != nil {
		if err == ee.ErrNoResults {
			return nil, nil
		}
		return nil, err
	}

	return &changesetSpecResolver{
		store:         r.store,
		httpFactory:   r.httpFactory,
		changesetSpec: changesetSpec,
		repoCtx:       ctx,
	}, nil
}

func (r *Resolver) CreateCampaign(ctx context.Context, args *graphqlbackend.CreateCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, _ := trace.New(ctx, "Resolver.CreateCampaign", fmt.Sprintf("CampaignSpec %s", args.CampaignSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	// TODO(sqs): Implement createCampaign when we've implemented applyCampaign and are happy about
	// how it works.
	return nil, errors.New("createCampaign is not yet implemented (use applyCampaign instead)")
}

func (r *Resolver) ApplyCampaign(ctx context.Context, args *graphqlbackend.ApplyCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.ApplyCampaign", fmt.Sprintf("CampaignSpec %s", args.CampaignSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	opts := ee.ApplyCampaignOpts{}

	opts.CampaignSpecRandID, err = unmarshalCampaignSpecID(args.CampaignSpec)
	if err != nil {
		return nil, err
	}

	if opts.CampaignSpecRandID == "" {
		return nil, ErrIDIsZero
	}

	if args.EnsureCampaign != nil {
		opts.EnsureCampaignID, err = campaigns.UnmarshalCampaignID(*args.EnsureCampaign)
		if err != nil {
			return nil, err
		}
	}

	svc := ee.NewService(r.store, r.httpFactory)
	campaign, err := svc.ApplyCampaign(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) CreateCampaignSpec(ctx context.Context, args *graphqlbackend.CreateCampaignSpecArgs) (graphqlbackend.CampaignSpecResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateCampaignSpec", fmt.Sprintf("Namespace %s, Spec %q", args.Namespace, args.CampaignSpec))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()
	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}

	// 🚨 SECURITY: Only site admins may create campaign specs for now.
	if !user.SiteAdmin {
		return nil, backend.ErrMustBeSiteAdmin
	}

	opts := ee.CreateCampaignSpecOpts{RawSpec: args.CampaignSpec}

	switch relay.UnmarshalKind(args.Namespace) {
	case "User":
		err = relay.UnmarshalSpec(args.Namespace, &opts.NamespaceUserID)
	case "Org":
		err = relay.UnmarshalSpec(args.Namespace, &opts.NamespaceOrgID)
	default:
		err = errors.Errorf("Invalid namespace %q", args.Namespace)
	}

	if err != nil {
		return nil, err
	}

	for _, graphqlID := range args.ChangesetSpecs {
		randID, err := unmarshalChangesetSpecID(graphqlID)
		if err != nil {
			return nil, err
		}
		opts.ChangesetSpecRandIDs = append(opts.ChangesetSpecRandIDs, randID)
	}

	svc := ee.NewService(r.store, r.httpFactory)
	campaignSpec, err := svc.CreateCampaignSpec(ctx, opts)
	if err != nil {
		return nil, err
	}

	specResolver := &campaignSpecResolver{
		store:        r.store,
		httpFactory:  r.httpFactory,
		campaignSpec: campaignSpec,
	}

	return specResolver, nil
}

func (r *Resolver) CreateChangesetSpec(ctx context.Context, args *graphqlbackend.CreateChangesetSpecArgs) (graphqlbackend.ChangesetSpecResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.CreateChangesetSpec", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	user, err := db.Users.GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "%v", backend.ErrNotAuthenticated)
	}

	// 🚨 SECURITY: Only site admins may create campaign specs for now.
	if !user.SiteAdmin {
		return nil, backend.ErrMustBeSiteAdmin
	}

	svc := ee.NewService(r.store, r.httpFactory)
	spec, err := svc.CreateChangesetSpec(ctx, args.ChangesetSpec, user.ID)
	if err != nil {
		return nil, err
	}

	resolver := &changesetSpecResolver{
		store:         r.store,
		httpFactory:   r.httpFactory,
		changesetSpec: spec,
		repoCtx:       ctx,
	}
	return resolver, nil
}

func (r *Resolver) MoveCampaign(ctx context.Context, args *graphqlbackend.MoveCampaignArgs) (graphqlbackend.CampaignResolver, error) {
	var err error
	tr, ctx := trace.New(ctx, "Resolver.MoveCampaign", fmt.Sprintf("Campaign %s", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	var opts ee.MoveCampaignOpts

	if args.NewName != nil {
		opts.NewName = *args.NewName
	}

	if args.NewNamespace != nil {
		newNamespace := *args.NewNamespace
		switch relay.UnmarshalKind(newNamespace) {
		case "User":
			err = relay.UnmarshalSpec(newNamespace, &opts.NewNamespaceUserID)
		case "Org":
			err = relay.UnmarshalSpec(newNamespace, &opts.NewNamespaceOrgID)
		default:
			err = errors.Errorf("Invalid namespace %q", newNamespace)
		}

		if err != nil {
			return nil, err
		}
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: MoveCampaign checks whether the current user is authorized.
	campaign, err := svc.MoveCampaign(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) DeleteCampaign(ctx context.Context, args *graphqlbackend.DeleteCampaignArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.DeleteCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, err
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: DeleteCampaign checks whether current user is authorized.
	err = svc.DeleteCampaign(ctx, campaignID)
	return &graphqlbackend.EmptyResponse{}, err
}

func (r *Resolver) Campaigns(ctx context.Context, args *graphqlbackend.ListCampaignArgs) (graphqlbackend.CampaignsConnectionResolver, error) {
	// 🚨 SECURITY: Only site admins or users when read-access is enabled may access campaign.
	if err := allowReadAccess(ctx); err != nil {
		return nil, err
	}
	opts := ee.ListCampaignsOpts{}
	state, err := parseCampaignState(args.State)
	if err != nil {
		return nil, err
	}
	opts.State = state
	if args.First != nil {
		opts.Limit = int(*args.First)
	}
	authErr := backend.CheckCurrentUserIsSiteAdmin(ctx)
	if authErr != nil && authErr != backend.ErrMustBeSiteAdmin {
		return nil, err
	}
	isSiteAdmin := authErr != backend.ErrMustBeSiteAdmin
	if !isSiteAdmin {
		if args.ViewerCanAdminister != nil && *args.ViewerCanAdminister {
			actor := actor.FromContext(ctx)
			opts.OnlyForAuthor = actor.UID
		}
	}
	return &campaignsConnectionResolver{
		store:       r.store,
		httpFactory: r.httpFactory,
		opts:        opts,
	}, nil
}

// listChangesetOptsFromArgs turns the graphqlbackend.ListChangesetsArgs into
// ListChangesetsOpts.
// If the args do not include a filter that would reveal sensitive information
// about a changeset the user doesn't have access to, the second return value
// is false.
func listChangesetOptsFromArgs(args *graphqlbackend.ListChangesetsArgs) (opts ee.ListChangesetsOpts, optsSafe bool, err error) {
	if args == nil {
		return opts, true, nil
	}

	safe := true

	if args.First != nil {
		opts.Limit = int(*args.First)
	}

	if args.State != nil {
		state := *args.ExternalState
		if !state.Valid() {
			return opts, false, errors.New("changeset state not valid")
		}
		opts.ExternalState = &state
	}
	if args.ReviewState != nil {
		state := *args.ReviewState
		if !state.Valid() {
			return opts, false, errors.New("changeset review state not valid")
		}
		opts.ExternalReviewState = &state
		// If the user filters by ReviewState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}
	if args.CheckState != nil {
		state := *args.CheckState
		if !state.Valid() {
			return opts, false, errors.New("changeset check state not valid")
		}
		opts.ExternalCheckState = &state
		// If the user filters by CheckState we cannot include hidden
		// changesets, since that would leak information.
		safe = false
	}

	return opts, safe, nil
}

func (r *Resolver) CloseCampaign(ctx context.Context, args *graphqlbackend.CloseCampaignArgs) (_ graphqlbackend.CampaignResolver, err error) {
	tr, ctx := trace.New(ctx, "Resolver.CloseCampaign", fmt.Sprintf("Campaign: %q", args.Campaign))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	campaignID, err := campaigns.UnmarshalCampaignID(args.Campaign)
	if err != nil {
		return nil, errors.Wrap(err, "unmarshaling campaign id")
	}

	if campaignID == 0 {
		return nil, ErrIDIsZero
	}

	svc := ee.NewService(r.store, r.httpFactory)
	// 🚨 SECURITY: CloseCampaign checks whether current user is authorized.
	campaign, err := svc.CloseCampaign(ctx, campaignID, args.CloseChangesets)
	if err != nil {
		return nil, errors.Wrap(err, "closing campaign")
	}

	return &campaignResolver{store: r.store, httpFactory: r.httpFactory, Campaign: campaign}, nil
}

func (r *Resolver) SyncChangeset(ctx context.Context, args *graphqlbackend.SyncChangesetArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	tr, ctx := trace.New(ctx, "Resolver.SyncChangeset", fmt.Sprintf("Changeset: %q", args.Changeset))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	changesetID, err := unmarshalChangesetID(args.Changeset)
	if err != nil {
		return nil, err
	}

	if changesetID == 0 {
		return nil, ErrIDIsZero
	}

	// 🚨 SECURITY: EnqueueChangesetSync checks whether current user is authorized.
	svc := ee.NewService(r.store, r.httpFactory)
	if err = svc.EnqueueChangesetSync(ctx, changesetID); err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func parseCampaignState(s *string) (campaigns.CampaignState, error) {
	if s == nil {
		return campaigns.CampaignStateAny, nil
	}
	switch *s {
	case "OPEN":
		return campaigns.CampaignStateOpen, nil
	case "CLOSED":
		return campaigns.CampaignStateClosed, nil
	default:
		return campaigns.CampaignStateAny, fmt.Errorf("unknown state %q", *s)
	}
}

func checkSiteAdminOrSameUser(ctx context.Context, userID int32) (bool, error) {
	// 🚨 SECURITY: Only site admins or the authors of a campaign have campaign
	// admin rights.
	if err := backend.CheckSiteAdminOrSameUser(ctx, userID); err != nil {
		if _, ok := err.(*backend.InsufficientAuthorizationError); ok {
			return false, nil
		}

		return false, err
	}
	return true, nil
}
