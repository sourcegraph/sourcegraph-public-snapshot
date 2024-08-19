package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewResolver returns a new Resolver that uses the given database.
func NewResolver(logger log.Logger, db database.DB) graphqlbackend.PromptsResolver {
	return &Resolver{logger: logger, db: db}
}

type Resolver struct {
	logger log.Logger
	db     database.DB
}

func (r *Resolver) Now() time.Time {
	return r.db.CodeMonitors().Now()
}

const PromptKind = "Prompt"

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		PromptKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.PromptByID(ctx, id)
		},
	}
}

type promptResolver struct {
	db database.DB
	s  types.Prompt
}

func marshalPromptID(promptID int32) graphql.ID {
	return relay.MarshalID(PromptKind, promptID)
}

func unmarshalPromptID(id graphql.ID) (promptID int32, err error) {
	err = relay.UnmarshalSpec(id, &promptID)
	return
}

func checkActorCanViewPrompt(ctx context.Context, db database.DB, p *types.Prompt) error {
	// 🚨 SECURITY: Public (non-secret) prompts can be viewed by anyone.
	if !p.VisibilitySecret {
		return nil
	}

	// 🚨 SECURITY: Make sure the current user has permission to get the secret prompt.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, db, p.Owner); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) PromptByID(ctx context.Context, id graphql.ID) (graphqlbackend.PromptResolver, error) {
	intID, err := unmarshalPromptID(id)
	if err != nil {
		return nil, err
	}

	p, err := r.db.Prompts().GetByID(ctx, intID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check whether the actor can view the saved search.
	if err := checkActorCanViewPrompt(ctx, r.db, p); err != nil {
		return nil, err
	}

	prompt := &promptResolver{
		db: r.db,
		s:  *p,
	}
	return prompt, nil
}

func (r *promptResolver) ID() graphql.ID {
	return marshalPromptID(r.s.ID)
}

func (r *promptResolver) Name() string { return r.s.Name }

func (r *promptResolver) Description() string { return r.s.Description }

func (r *promptResolver) Definition() graphqlbackend.PromptDefinitionResolver {
	return graphqlbackend.PromptDefinitionResolver{Text_: r.s.DefinitionText}
}

func (r *promptResolver) Draft() bool { return r.s.Draft }

func (r *promptResolver) Owner(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	if r.s.Owner.User != nil {
		n, err := graphqlbackend.NamespaceByID(ctx, r.db, graphqlbackend.MarshalUserID(*r.s.Owner.User))
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: n}, nil
	}
	if r.s.Owner.Org != nil {
		n, err := graphqlbackend.NamespaceByID(ctx, r.db, graphqlbackend.MarshalOrgID(*r.s.Owner.Org))
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: n}, nil
	}
	return nil, errors.New("no owner")
}

func (r *promptResolver) Visibility() graphqlbackend.PromptVisibility {
	if r.s.VisibilitySecret {
		return graphqlbackend.PromptVisibilitySecret
	}
	return graphqlbackend.PromptVisibilityPublic
}

func (r *promptResolver) NameWithOwner(ctx context.Context) (string, error) {
	owner, err := r.Owner(ctx)
	if err != nil {
		return "", err
	}
	return owner.NamespaceName() + "/" + r.s.Name, nil
}

func (r *promptResolver) CreatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if userID := r.s.CreatedByUser; userID != nil {
		return graphqlbackend.UserByIDInt32(ctx, r.db, *userID)
	}
	return nil, nil
}

func (r *promptResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.s.CreatedAt}
}

func (r *promptResolver) UpdatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if userID := r.s.UpdatedByUser; userID != nil {
		return graphqlbackend.UserByIDInt32(ctx, r.db, *userID)
	}
	return nil, nil
}

func (r *promptResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.s.UpdatedAt}
}

func (r *promptResolver) URL() string {
	return "/prompts/" + string(r.ID())
}

func (r *promptResolver) ViewerCanAdminister(ctx context.Context) bool {
	// 🚨 SECURITY: If the visibility is public, then the user can see it, but they can only
	// administer it if they are authorized for the namespace (as an org member or their own user
	// account).
	err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, r.s.Owner)
	return err == nil
}

func (r *Resolver) toPromptResolver(entry types.Prompt) *promptResolver {
	return &promptResolver{db: r.db, s: entry}
}

func (r *Resolver) Prompts(ctx context.Context, args graphqlbackend.PromptsArgs) (*graphqlbackend.PromptConnectionResolver, error) {
	connectionStore := &promptsConnectionStore{db: r.db}

	if args.Query != nil {
		connectionStore.listArgs.Query = *args.Query
	}
	connectionStore.listArgs.HideDrafts = !args.IncludeDrafts

	if args.Owner != nil {
		// 🚨 SECURITY: Make sure the current user has permission to view prompts of the specified
		// owner.
		owner, err := graphqlbackend.CheckAuthorizedForNamespace(ctx, r.db, *args.Owner)
		if err != nil {
			return nil, err
		}
		connectionStore.listArgs.Owner = owner
	}

	if args.ViewerIsAffiliated != nil && *args.ViewerIsAffiliated {
		// 🚨 SECURITY: The auth check is implicit here because `viewerIsAffiliated` is a bool and
		// only the current user can be used, and the actor *is* the current user.
		currentUser, err := auth.CurrentUser(ctx, r.db)
		if err != nil {
			return nil, err
		}
		if currentUser != nil {
			connectionStore.listArgs.AffiliatedUser = &currentUser.ID
		} else {
			// For anonymous visitors, just show all public prompts.
			connectionStore.listArgs.PublicOnly = true
		}
	}

	// 🚨 SECURITY: Only site admins can list all non-public prompts.
	if connectionStore.listArgs.Owner == nil && connectionStore.listArgs.AffiliatedUser == nil && !connectionStore.listArgs.PublicOnly {
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, errors.Wrap(err, "must specify owner or viewerIsAffiliated args")
		}
	}

	var orderBy database.PromptsOrderBy
	switch args.OrderBy {
	case graphqlbackend.PromptsOrderByNameWithOwner:
		orderBy = database.PromptsOrderByNameWithOwner
	case graphqlbackend.PromptsOrderByUpdatedAt:
		orderBy = database.PromptsOrderByUpdatedAt
	default:
		// Don't expose PromptsOrderByID option to the GraphQL API. This is not a security
		// thing, it's just to avoid allowing clients to depend on our implementation details.
		return nil, errors.New("invalid orderBy")
	}

	opts := gqlutil.ConnectionResolverOptions{}
	opts.OrderBy, opts.Ascending = orderBy.ToOptions()

	return gqlutil.NewConnectionResolver(connectionStore, &args.ConnectionResolverArgs, &opts)
}

func (r *Resolver) CreatePrompt(ctx context.Context, args *graphqlbackend.CreatePromptArgs) (graphqlbackend.PromptResolver, error) {
	// 🚨 SECURITY: Make sure the current user has permission to create a prompt in the
	// specified owner namespace.
	namespace, err := graphqlbackend.CheckAuthorizedForNamespace(ctx, r.db, args.Input.Owner)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check that the user can create public-visibility items on this instance.
	if !args.Input.Visibility.IsSecret() {
		if err := graphqlbackend.ViewerCanChangeLibraryItemVisibilityToPublic(ctx, r.db); err != nil {
			return nil, err
		}
	}

	p, err := r.db.Prompts().Create(ctx, &types.Prompt{
		Name:             args.Input.Name,
		Description:      args.Input.Description,
		DefinitionText:   args.Input.DefinitionText,
		Draft:            args.Input.Draft,
		Owner:            *namespace,
		VisibilitySecret: args.Input.Visibility.IsSecret(),
	})
	if err != nil {
		return nil, err
	}

	return r.toPromptResolver(*p), nil
}

func (r *Resolver) UpdatePrompt(ctx context.Context, args *graphqlbackend.UpdatePromptArgs) (graphqlbackend.PromptResolver, error) {
	id, err := unmarshalPromptID(args.ID)
	if err != nil {
		return nil, err
	}

	old, err := r.db.Prompts().GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get existing prompt")
	}

	// 🚨 SECURITY: Make sure the current user has permission to update a prompt for the
	// specified owner namespace.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, old.Owner); err != nil {
		return nil, err
	}

	p, err := r.db.Prompts().Update(ctx, &types.Prompt{
		ID:             id,
		Name:           args.Input.Name,
		Description:    args.Input.Description,
		DefinitionText: args.Input.DefinitionText,
		Draft:          args.Input.Draft,
		Owner:          old.Owner, // use transferPromptOwnership to update the owner
	})
	if err != nil {
		return nil, err
	}

	return r.toPromptResolver(*p), nil
}

func (r *Resolver) TransferPromptOwnership(ctx context.Context, args *graphqlbackend.TransferPromptOwnershipArgs) (graphqlbackend.PromptResolver, error) {
	id, err := unmarshalPromptID(args.ID)
	if err != nil {
		return nil, err
	}
	p, err := r.db.Prompts().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Make sure the current user has permission to administer a prompt for
	// *BOTH* the current and new owner namespaces.
	//
	// Check the user can administer prompts in the current owner's namespace:
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, p.Owner); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: ...and check the user can administer prompts in the new owner's
	// namespace:
	newOwner, err := graphqlbackend.CheckAuthorizedForNamespace(ctx, r.db, args.NewOwner)
	if err != nil {
		return nil, err
	}

	p, err = r.db.Prompts().UpdateOwner(ctx, id, *newOwner)
	if err != nil {
		return nil, err
	}
	return r.toPromptResolver(*p), nil
}

func (r *Resolver) ChangePromptVisibility(ctx context.Context, args *graphqlbackend.ChangePromptVisibilityArgs) (graphqlbackend.PromptResolver, error) {
	// 🚨 SECURITY: Check that the user can change the visibility on this instance.
	if err := graphqlbackend.ViewerCanChangeLibraryItemVisibilityToPublic(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmarshalPromptID(args.ID)
	if err != nil {
		return nil, err
	}
	p, err := r.db.Prompts().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check the user can administer prompts in the owner's namespace.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, p.Owner); err != nil {
		return nil, err
	}

	p, err = r.db.Prompts().UpdateVisibility(ctx, id, args.NewVisibility.IsSecret())
	if err != nil {
		return nil, err
	}
	return r.toPromptResolver(*p), nil
}

func (r *Resolver) DeletePrompt(ctx context.Context, args *graphqlbackend.DeletePromptArgs) (*graphqlbackend.EmptyResponse, error) {
	id, err := unmarshalPromptID(args.ID)
	if err != nil {
		return nil, err
	}
	p, err := r.db.Prompts().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Make sure the current user has permission to delete a prompt for the
	// specified owner namespace.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, p.Owner); err != nil {
		return nil, err
	}

	if err := r.db.Prompts().Delete(ctx, id); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}
