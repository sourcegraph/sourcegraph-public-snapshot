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
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewResolver returns a new Resolver that uses the given database.
func NewResolver(logger log.Logger, db database.DB) graphqlbackend.SavedSearchesResolver {
	return &Resolver{logger: logger, db: db}
}

type Resolver struct {
	logger log.Logger
	db     database.DB
}

func (r *Resolver) Now() time.Time {
	return r.db.CodeMonitors().Now()
}

const SavedSearchKind = "SavedSearch"

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		SavedSearchKind: func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.SavedSearchByID(ctx, id)
		},
	}
}

type savedSearchResolver struct {
	db database.DB
	s  types.SavedSearch
}

func marshalSavedSearchID(savedSearchID int32) graphql.ID {
	return relay.MarshalID(SavedSearchKind, savedSearchID)
}

func unmarshalSavedSearchID(id graphql.ID) (savedSearchID int32, err error) {
	err = relay.UnmarshalSpec(id, &savedSearchID)
	return
}

func checkActorCanViewSavedSearch(ctx context.Context, db database.DB, ss *types.SavedSearch) error {
	// 🚨 SECURITY: Public (non-secret) saved searches can be viewed by anyone.
	if !ss.VisibilitySecret {
		return nil
	}

	// 🚨 SECURITY: Make sure the current user has permission to get the secret saved search.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, db, ss.Owner); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) SavedSearchByID(ctx context.Context, id graphql.ID) (graphqlbackend.SavedSearchResolver, error) {
	intID, err := unmarshalSavedSearchID(id)
	if err != nil {
		return nil, err
	}

	ss, err := r.db.SavedSearches().GetByID(ctx, intID)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}

	// 🚨 SECURITY: Check whether the actor can view the saved search.
	if err := checkActorCanViewSavedSearch(ctx, r.db, ss); err != nil {
		return nil, err
	}

	savedSearch := &savedSearchResolver{
		db: r.db,
		s:  *ss,
	}
	return savedSearch, nil
}

func (r *savedSearchResolver) ID() graphql.ID {
	return marshalSavedSearchID(r.s.ID)
}

func (r *savedSearchResolver) Description() string { return r.s.Description }

func (r *savedSearchResolver) Query() string { return r.s.Query }

func (r *savedSearchResolver) Draft() bool { return r.s.Draft }

func (r *savedSearchResolver) Owner(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
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

func (r *savedSearchResolver) Visibility() graphqlbackend.SavedSearchVisibility {
	if r.s.VisibilitySecret {
		return graphqlbackend.SavedSearchVisibilitySecret
	}
	return graphqlbackend.SavedSearchVisibilityPublic
}

func (r *savedSearchResolver) CreatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if userID := r.s.CreatedByUser; userID != nil {
		return graphqlbackend.UserByIDInt32(ctx, r.db, *userID)
	}
	return nil, nil
}

func (r *savedSearchResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.s.CreatedAt}
}

func (r *savedSearchResolver) UpdatedBy(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	if userID := r.s.UpdatedByUser; userID != nil {
		return graphqlbackend.UserByIDInt32(ctx, r.db, *userID)
	}
	return nil, nil
}

func (r *savedSearchResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.s.UpdatedAt}
}

func (r *savedSearchResolver) URL() string {
	return "/saved-searches/" + string(r.ID())
}

func (r *savedSearchResolver) ViewerCanAdminister(ctx context.Context) bool {
	// 🚨 SECURITY: If the visibility is public, then the user can see it, but they can only
	// administer it if they are authorized for the namespace (as an org member or their own user
	// account).
	err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, r.s.Owner)
	return err == nil
}

func (r *Resolver) toSavedSearchResolver(entry types.SavedSearch) *savedSearchResolver {
	return &savedSearchResolver{db: r.db, s: entry}
}

func (r *Resolver) SavedSearches(ctx context.Context, args graphqlbackend.SavedSearchesArgs) (*graphqlbackend.SavedSearchConnectionResolver, error) {
	connectionStore := &savedSearchesConnectionStore{db: r.db}

	if args.Query != nil {
		connectionStore.listArgs.Query = *args.Query
	}
	connectionStore.listArgs.HideDrafts = !args.IncludeDrafts

	if args.Owner != nil {
		// 🚨 SECURITY: Make sure the current user has permission to view saved searches of the
		// specified owner.
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
			// For anonymous visitors, just show all public saved searches.
			connectionStore.listArgs.PublicOnly = true
		}
	}

	// 🚨 SECURITY: Only site admins can list all non-public saved searches.
	if connectionStore.listArgs.Owner == nil && connectionStore.listArgs.AffiliatedUser == nil && !connectionStore.listArgs.PublicOnly {
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, errors.Wrap(err, "must specify owner or viewerIsAffiliated args")
		}
	}

	var orderBy database.SavedSearchesOrderBy
	switch args.OrderBy {
	case graphqlbackend.SavedSearchesOrderByDescription:
		orderBy = database.SavedSearchesOrderByDescription
	case graphqlbackend.SavedSearchesOrderByUpdatedAt:
		orderBy = database.SavedSearchesOrderByUpdatedAt
	default:
		// Don't expose SavedSearchesOrderByID option to the GraphQL API. This is not a security
		// thing, it's just to avoid allowing clients to depend on our implementation details.
		return nil, errors.New("invalid orderBy")
	}

	opts := gqlutil.ConnectionResolverOptions{}
	opts.OrderBy, opts.Ascending = orderBy.ToOptions()

	return gqlutil.NewConnectionResolver(connectionStore, &args.ConnectionResolverArgs, &opts)
}

func (r *Resolver) CreateSavedSearch(ctx context.Context, args *graphqlbackend.CreateSavedSearchArgs) (graphqlbackend.SavedSearchResolver, error) {
	// 🚨 SECURITY: Make sure the current user has permission to create a saved search in the
	// specified owner namespace.
	namespace, err := graphqlbackend.CheckAuthorizedForNamespace(ctx, r.db, args.Input.Owner)
	if err != nil {
		return nil, err
	}

	if !queryHasPatternType(args.Input.Query) {
		return nil, errMissingPatternType
	}

	// 🚨 SECURITY: Check that the user can create public-visibility items on this instance.
	if !args.Input.Visibility.IsSecret() {
		if err := graphqlbackend.ViewerCanChangeLibraryItemVisibilityToPublic(ctx, r.db); err != nil {
			return nil, err
		}
	}

	ss, err := r.db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description:      args.Input.Description,
		Query:            args.Input.Query,
		Draft:            args.Input.Draft,
		Owner:            *namespace,
		VisibilitySecret: args.Input.Visibility.IsSecret(),
	})
	if err != nil {
		return nil, err
	}

	return r.toSavedSearchResolver(*ss), nil
}

func (r *Resolver) UpdateSavedSearch(ctx context.Context, args *graphqlbackend.UpdateSavedSearchArgs) (graphqlbackend.SavedSearchResolver, error) {
	id, err := unmarshalSavedSearchID(args.ID)
	if err != nil {
		return nil, err
	}

	old, err := r.db.SavedSearches().GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "get existing saved search")
	}

	// 🚨 SECURITY: Make sure the current user has permission to update a saved search for the
	// specified owner namespace.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, old.Owner); err != nil {
		return nil, err
	}

	if !queryHasPatternType(args.Input.Query) {
		return nil, errMissingPatternType
	}

	ss, err := r.db.SavedSearches().Update(ctx, &types.SavedSearch{
		ID:          id,
		Description: args.Input.Description,
		Query:       args.Input.Query,
		Draft:       args.Input.Draft,
		Owner:       old.Owner, // use transferSavedSearchOwnership to update the owner
	})
	if err != nil {
		return nil, err
	}

	return r.toSavedSearchResolver(*ss), nil
}

func (r *Resolver) TransferSavedSearchOwnership(ctx context.Context, args *graphqlbackend.TransferSavedSearchOwnershipArgs) (graphqlbackend.SavedSearchResolver, error) {
	id, err := unmarshalSavedSearchID(args.ID)
	if err != nil {
		return nil, err
	}
	ss, err := r.db.SavedSearches().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Make sure the current user has permission to administer a saved search for
	// *BOTH* the current and new owner namespaces.
	//
	// Check the user can administer saved searches in the current owner's namespace:
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, ss.Owner); err != nil {
		return nil, err
	}
	// 🚨 SECURITY: ...and check the user can administer saved searches in the new owner's
	// namespace:
	newOwner, err := graphqlbackend.CheckAuthorizedForNamespace(ctx, r.db, args.NewOwner)
	if err != nil {
		return nil, err
	}

	ss, err = r.db.SavedSearches().UpdateOwner(ctx, id, *newOwner)
	if err != nil {
		return nil, err
	}
	return r.toSavedSearchResolver(*ss), nil
}

func (r *Resolver) ChangeSavedSearchVisibility(ctx context.Context, args *graphqlbackend.ChangeSavedSearchVisibilityArgs) (graphqlbackend.SavedSearchResolver, error) {
	// 🚨 SECURITY: Check that the user can change the visibility on this instance.
	if err := graphqlbackend.ViewerCanChangeLibraryItemVisibilityToPublic(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmarshalSavedSearchID(args.ID)
	if err != nil {
		return nil, err
	}
	ss, err := r.db.SavedSearches().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Check the user can administer saved searches in the owner's namespace.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, ss.Owner); err != nil {
		return nil, err
	}

	ss, err = r.db.SavedSearches().UpdateVisibility(ctx, id, args.NewVisibility.IsSecret())
	if err != nil {
		return nil, err
	}
	return r.toSavedSearchResolver(*ss), nil
}

func (r *Resolver) DeleteSavedSearch(ctx context.Context, args *graphqlbackend.DeleteSavedSearchArgs) (*graphqlbackend.EmptyResponse, error) {
	id, err := unmarshalSavedSearchID(args.ID)
	if err != nil {
		return nil, err
	}
	ss, err := r.db.SavedSearches().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Make sure the current user has permission to delete a saved search for the
	// specified owner namespace.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, ss.Owner); err != nil {
		return nil, err
	}

	if err := r.db.SavedSearches().Delete(ctx, id); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

var patternType = lazyregexp.New(`(?i)\bpatternType:(literal|regexp|structural|standard|keyword)\b`)

func queryHasPatternType(query string) bool {
	return patternType.Match([]byte(query))
}

var errMissingPatternType = errors.New("a `patternType:` filter is required in the query for all saved searches. `patternType` can be \"keyword\", \"standard\", \"literal\", or \"regexp\"")
