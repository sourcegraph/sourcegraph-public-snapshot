package resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/search"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/service"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/campaigns/store"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/campaigns"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
)

func marshalCampaignSpecRandID(id string) graphql.ID {
	return relay.MarshalID("CampaignSpec", id)
}

func unmarshalCampaignSpecID(id graphql.ID) (campaignSpecRandID string, err error) {
	err = relay.UnmarshalSpec(id, &campaignSpecRandID)
	return
}

var _ graphqlbackend.CampaignSpecResolver = &campaignSpecResolver{}

type campaignSpecResolver struct {
	store *store.Store

	campaignSpec       *campaigns.CampaignSpec
	preloadedNamespace *graphqlbackend.NamespaceResolver

	// We cache the namespace on the resolver, since it's accessed more than once.
	namespaceOnce sync.Once
	namespace     *graphqlbackend.NamespaceResolver
	namespaceErr  error
}

func (r *campaignSpecResolver) ID() graphql.ID {
	// 🚨 SECURITY: This needs to be the RandID! We can't expose the
	// sequential, guessable ID.
	return marshalCampaignSpecRandID(r.campaignSpec.RandID)
}

func (r *campaignSpecResolver) OriginalInput() (string, error) {
	return r.campaignSpec.RawSpec, nil
}

func (r *campaignSpecResolver) ParsedInput() (graphqlbackend.JSONValue, error) {
	return graphqlbackend.JSONValue{Value: r.campaignSpec.Spec}, nil
}

func (r *campaignSpecResolver) ChangesetSpecs(ctx context.Context, args *graphqlbackend.ChangesetSpecsConnectionArgs) (graphqlbackend.ChangesetSpecConnectionResolver, error) {
	opts := store.ListChangesetSpecsOpts{}
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts.Limit = int(args.First)
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.Cursor = int64(id)
	}

	return &changesetSpecConnectionResolver{
		store:          r.store,
		opts:           opts,
		campaignSpecID: r.campaignSpec.ID,
	}, nil
}

func (r *campaignSpecResolver) ApplyPreview(ctx context.Context, args *graphqlbackend.ChangesetApplyPreviewConnectionArgs) (graphqlbackend.ChangesetApplyPreviewConnectionResolver, error) {
	if err := validateFirstParamDefaults(args.First); err != nil {
		return nil, err
	}
	opts := store.GetRewirerMappingsOpts{
		LimitOffset: &database.LimitOffset{
			Limit: int(args.First),
		},
		CurrentState: args.CurrentState,
	}
	if args.After != nil {
		id, err := strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
		opts.LimitOffset.Offset = id
	}
	if args.Search != nil {
		var err error
		opts.TextSearch, err = search.ParseTextSearch(*args.Search)
		if err != nil {
			return nil, errors.Wrap(err, "parsing search")
		}
	}

	return &changesetApplyPreviewConnectionResolver{
		store:          r.store,
		opts:           opts,
		action:         args.Action,
		campaignSpecID: r.campaignSpec.ID,
	}, nil
}

func (r *campaignSpecResolver) Description() graphqlbackend.CampaignDescriptionResolver {
	return &campaignDescriptionResolver{
		name:        r.campaignSpec.Spec.Name,
		description: r.campaignSpec.Spec.Description,
	}
}

func (r *campaignSpecResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.campaignSpec.UserID)
	if errcode.IsNotFound(err) {
		return nil, nil
	}
	return user, err
}

func (r *campaignSpecResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return r.computeNamespace(ctx)
}

func (r *campaignSpecResolver) computeNamespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	r.namespaceOnce.Do(func() {
		if r.preloadedNamespace != nil {
			r.namespace = r.preloadedNamespace
			return
		}
		var (
			err error
			n   = &graphqlbackend.NamespaceResolver{}
		)

		if r.campaignSpec.NamespaceUserID != 0 {
			n.Namespace, err = graphqlbackend.UserByIDInt32(ctx, r.store.DB(), r.campaignSpec.NamespaceUserID)
		} else {
			n.Namespace, err = graphqlbackend.OrgByIDInt32(ctx, r.store.DB(), r.campaignSpec.NamespaceOrgID)
		}

		if errcode.IsNotFound(err) {
			r.namespace = nil
			r.namespaceErr = errors.New("namespace of campaign spec has been deleted")
			return
		}

		r.namespace = n
		r.namespaceErr = err
	})
	return r.namespace, r.namespaceErr
}

func (r *campaignSpecResolver) ApplyURL(ctx context.Context) (string, error) {
	n, err := r.computeNamespace(ctx)
	if err != nil {
		return "", err
	}
	return campaignsApplyURL(n, r), nil
}

func (r *campaignSpecResolver) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.campaignSpec.CreatedAt}
}

func (r *campaignSpecResolver) ExpiresAt() *graphqlbackend.DateTime {
	return &graphqlbackend.DateTime{Time: r.campaignSpec.ExpiresAt()}
}

func (r *campaignSpecResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	return checkSiteAdminOrSameUser(ctx, r.campaignSpec.UserID)
}

type campaignDescriptionResolver struct {
	name, description string
}

func (r *campaignDescriptionResolver) Name() string {
	return r.name
}

func (r *campaignDescriptionResolver) Description() string {
	return r.description
}

func (r *campaignSpecResolver) DiffStat(ctx context.Context) (*graphqlbackend.DiffStat, error) {
	specsConnection := &changesetSpecConnectionResolver{
		store:          r.store,
		campaignSpecID: r.campaignSpec.ID,
	}

	specs, err := specsConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	totalStat := &graphqlbackend.DiffStat{}
	for _, spec := range specs {
		// If we can't convert it, that means it's hidden from the user and we
		// can simply skip it.
		if _, ok := spec.ToVisibleChangesetSpec(); !ok {
			continue
		}

		resolver, ok := spec.(*changesetSpecResolver)
		if !ok {
			// This should never happen.
			continue
		}

		stat := resolver.changesetSpec.DiffStat()
		totalStat.AddStat(stat)
	}

	return totalStat, nil
}

func (r *campaignSpecResolver) AppliesToCampaign(ctx context.Context) (graphqlbackend.CampaignResolver, error) {
	svc := service.New(r.store)
	campaign, err := svc.GetCampaignMatchingCampaignSpec(ctx, r.campaignSpec)
	if err != nil {
		return nil, err
	}
	if campaign == nil {
		return nil, nil
	}

	return &campaignResolver{
		store:    r.store,
		Campaign: campaign,
	}, nil
}

func (r *campaignSpecResolver) SupersedingCampaignSpec(ctx context.Context) (graphqlbackend.CampaignSpecResolver, error) {
	namespace, err := r.computeNamespace(ctx)
	if err != nil {
		return nil, err
	}

	svc := service.New(r.store)
	newest, err := svc.GetNewestCampaignSpec(ctx, r.store, r.campaignSpec, actor.FromContext(ctx).UID)
	if err != nil {
		return nil, err
	}

	// If this is the newest spec, then we can just return nil.
	if newest == nil || newest.ID == r.campaignSpec.ID {
		return nil, nil
	}

	// If this spec and the new spec have different creators, we shouldn't
	// return this as a superseding spec.
	if newest.UserID != r.campaignSpec.UserID {
		return nil, nil
	}

	// Create our new resolver, reusing as many fields as we can from this one.
	resolver := &campaignSpecResolver{
		store:              r.store,
		campaignSpec:       newest,
		preloadedNamespace: namespace,
	}

	return resolver, nil
}

func (r *campaignSpecResolver) ViewerCampaignsCodeHosts(ctx context.Context, args *graphqlbackend.ListViewerCampaignsCodeHostsArgs) (graphqlbackend.CampaignsCodeHostConnectionResolver, error) {
	actor := actor.FromContext(ctx)
	if !actor.IsAuthenticated() {
		return nil, backend.ErrNotAuthenticated
	}

	// Short path for site-admins when OnlyWithoutCredential is true: It will always be an empty list.
	if args.OnlyWithoutCredential {
		if authErr := backend.CheckCurrentUserIsSiteAdmin(ctx); authErr == nil {
			// For site-admins never return anything
			return &emptyCampaignsCodeHostConnectionResolver{}, nil
		} else if authErr != nil && authErr != backend.ErrMustBeSiteAdmin {
			return nil, authErr
		}
	}

	specs, _, err := r.store.ListChangesetSpecs(ctx, store.ListChangesetSpecsOpts{CampaignSpecID: r.campaignSpec.ID})
	if err != nil {
		return nil, err
	}

	offset := 0
	if args.After != nil {
		offset, err = strconv.Atoi(*args.After)
		if err != nil {
			return nil, err
		}
	}

	return &campaignsCodeHostConnectionResolver{
		userID:                actor.UID,
		onlyWithoutCredential: args.OnlyWithoutCredential,
		store:                 r.store,
		opts: store.ListCodeHostsOpts{
			RepoIDs: specs.RepoIDs(),
		},
		limitOffset: database.LimitOffset{
			Limit:  int(args.First),
			Offset: offset,
		},
	}, nil
}
