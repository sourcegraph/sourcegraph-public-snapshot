package resolvers

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
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

	// 🚨 SECURITY: Make sure the current user has permission to get the saved search.
	if err := graphqlbackend.CheckAuthorizedForNamespaceByIDs(ctx, r.db, ss.Owner); err != nil {
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

func (r *savedSearchResolver) URL() string {
	return "/saved-searches/" + string(r.ID())
}

func (r *savedSearchResolver) ViewerCanAdminister(ctx context.Context) (bool, error) {
	// Right now, any user who can see a saved search can edit/administer it, but in the future we
	// can add more access levels.
	return true, nil
}

func (r *Resolver) toSavedSearchResolver(entry types.SavedSearch) *savedSearchResolver {
	return &savedSearchResolver{db: r.db, s: entry}
}

func (r *Resolver) SavedSearches(ctx context.Context, args graphqlbackend.SavedSearchesArgs) (*graphqlbackend.SavedSearchConnectionResolver, error) {
	connectionStore := &savedSearchesConnectionStore{db: r.db}

	if args.Query != nil {
		connectionStore.listArgs.Query = *args.Query
	}

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
		if currentUser == nil {
			// 🚨 SECURITY: Just in case, ensure the user is signed in.
			return nil, auth.ErrNotAuthenticated
		}
		connectionStore.listArgs.AffiliatedUser = &currentUser.ID
	}

	// 🚨 SECURITY: Only site admins can list all saved searches.
	if connectionStore.listArgs.Owner == nil && connectionStore.listArgs.AffiliatedUser == nil {
		if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
			return nil, errors.Wrap(err, "must specify owner or viewerIsAffiliated args")
		}
	}

	// Don't expose SavedSearchesOrderByID option to the GraphQL API. This is not a security thing,
	// it's just to avoid allowing clients to depend on our implementation details.
	orderBy := database.SavedSearchesOrderByUpdatedAt
	if args.OrderBy != graphqlbackend.SavedSearchesOrderByDescription {
		orderBy = database.SavedSearchesOrderByDescription
	}

	opts := graphqlutil.ConnectionResolverOptions{MaxPageSize: 1000}
	opts.OrderBy, opts.Ascending = orderBy.ToOptions()

	return graphqlutil.NewConnectionResolver(connectionStore, &args.ConnectionResolverArgs, &opts)
}

type savedSearchesConnectionStore struct {
	db       database.DB
	listArgs database.SavedSearchListArgs
}

func (s *savedSearchesConnectionStore) MarshalCursor(node graphqlbackend.SavedSearchResolver, orderBy database.OrderBy) (*string, error) {
	// TODO!(sqs): will refactor to not use raw db column name strings before merging
	var value string
	var column string
	if len(orderBy) > 0 {
		column = orderBy[0].Field
	} else {
		column = "id"
	}
	switch column {
	case "id":
		value = string(node.ID())
	case "description":
		value = node.Description()
	case "updated_at":
		value = node.(*savedSearchResolver).s.UpdatedAt.Format(time.RFC3339)
	default:
		return nil, errors.New("unexpected orderBy")
	}
	return &value, nil
}

func (s *savedSearchesConnectionStore) UnmarshalCursor(cursor string, orderBy database.OrderBy) ([]any, error) {
	// TODO!(sqs): will refactor to not use raw db column name strings before merging
	var cursorValue []any
	var column string
	if len(orderBy) > 0 {
		column = orderBy[0].Field
	} else {
		column = "id"
	}
	switch column {
	case "id":
		nodeID, err := unmarshalSavedSearchID(graphql.ID(cursor))
		if err != nil {
			return nil, err
		}
		cursorValue = []any{nodeID}
	case "description":
		cursorValue = []any{cursor}
	case "updated_at":
		updatedAt, err := time.Parse(time.RFC3339, cursor)
		if err != nil {
			return nil, err
		}
		cursorValue = []any{updatedAt}
	default:
		return nil, errors.New("unexpected orderBy")
	}
	return cursorValue, nil
}

func (s *savedSearchesConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.SavedSearches().Count(ctx, s.listArgs)
	return int32(count), err
}

func (s *savedSearchesConnectionStore) ComputeNodes(ctx context.Context, pgArgs *database.PaginationArgs) ([]graphqlbackend.SavedSearchResolver, error) {
	dbResults, err := s.db.SavedSearches().List(ctx, s.listArgs, pgArgs)
	if err != nil {
		return nil, err
	}

	var results []graphqlbackend.SavedSearchResolver
	for _, savedSearch := range dbResults {
		results = append(results, &savedSearchResolver{db: s.db, s: *savedSearch})
	}

	return results, nil
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

	ss, err := r.db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description: args.Input.Description,
		Query:       args.Input.Query,
		Owner:       *namespace,
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
		Owner:       old.Owner,
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
