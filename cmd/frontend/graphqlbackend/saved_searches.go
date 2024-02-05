package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type savedSearchResolver struct {
	db database.DB
	s  types.SavedSearch
}

func marshalSavedSearchID(savedSearchID int32) graphql.ID {
	return relay.MarshalID("SavedSearch", savedSearchID)
}

func unmarshalSavedSearchID(id graphql.ID) (savedSearchID int32, err error) {
	err = relay.UnmarshalSpec(id, &savedSearchID)
	return
}

func (r *schemaResolver) savedSearchByID(ctx context.Context, id graphql.ID) (*savedSearchResolver, error) {
	intID, err := unmarshalSavedSearchID(id)
	if err != nil {
		return nil, err
	}

	ss, err := r.db.SavedSearches().GetByID(ctx, intID)
	if err != nil {
		return nil, err
	}

	// 🚨 SECURITY: Make sure the current user has permission to get the saved
	// search.
	if ss.Config.UserID != nil {
		if *ss.Config.UserID != actor.FromContext(ctx).UID {
			return nil, &auth.InsufficientAuthorizationError{
				Message: "current user has insufficient privileges to view saved search",
			}
		}
	} else if ss.Config.OrgID != nil {
		if err := auth.CheckOrgAccess(ctx, r.db, *ss.Config.OrgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("failed to get saved search: no Org ID or User ID associated with saved search")
	}

	savedSearch := &savedSearchResolver{
		db: r.db,
		s: types.SavedSearch{
			ID:              intID,
			Description:     ss.Config.Description,
			Query:           ss.Config.Query,
			Notify:          ss.Config.Notify,
			NotifySlack:     ss.Config.NotifySlack,
			UserID:          ss.Config.UserID,
			OrgID:           ss.Config.OrgID,
			SlackWebhookURL: ss.Config.SlackWebhookURL,
		},
	}
	return savedSearch, nil
}

func (r savedSearchResolver) ID() graphql.ID {
	return marshalSavedSearchID(r.s.ID)
}

func (r savedSearchResolver) Notify() bool {
	return r.s.Notify
}

func (r savedSearchResolver) NotifySlack() bool {
	return r.s.NotifySlack
}

func (r savedSearchResolver) Description() string { return r.s.Description }

func (r savedSearchResolver) Query() string { return r.s.Query }

func (r savedSearchResolver) Namespace(ctx context.Context) (*NamespaceResolver, error) {
	if r.s.OrgID != nil {
		n, err := NamespaceByID(ctx, r.db, MarshalOrgID(*r.s.OrgID))
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}
	if r.s.UserID != nil {
		n, err := NamespaceByID(ctx, r.db, MarshalUserID(*r.s.UserID))
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}
	return nil, nil
}

func (r savedSearchResolver) SlackWebhookURL() *string { return r.s.SlackWebhookURL }

func (r *schemaResolver) toSavedSearchResolver(entry types.SavedSearch) *savedSearchResolver {
	return &savedSearchResolver{db: r.db, s: entry}
}

type savedSearchesArgs struct {
	graphqlutil.ConnectionResolverArgs
	Namespace graphql.ID
}

func (r *schemaResolver) SavedSearches(ctx context.Context, args savedSearchesArgs) (*graphqlutil.ConnectionResolver[*savedSearchResolver], error) {
	var userID, orgID int32
	if err := UnmarshalNamespaceID(args.Namespace, &userID, &orgID); err != nil {
		return nil, err
	}

	if userID != 0 {
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
			return nil, err
		}
	} else if orgID != 0 {
		if err := auth.CheckOrgAccessOrSiteAdmin(ctx, r.db, orgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("User or Organisation namespace must be provided.")
	}

	connectionStore := &savedSearchesConnectionStore{
		db:     r.db,
		userID: &userID,
		orgID:  &orgID,
	}

	return graphqlutil.NewConnectionResolver[*savedSearchResolver](connectionStore, &args.ConnectionResolverArgs, nil)
}

type savedSearchesConnectionStore struct {
	db     database.DB
	userID *int32
	orgID  *int32
}

func (s *savedSearchesConnectionStore) MarshalCursor(node *savedSearchResolver, _ database.OrderBy) (*string, error) {
	cursor := string(node.ID())

	return &cursor, nil
}

func (s *savedSearchesConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	nodeID, err := unmarshalSavedSearchID(graphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	return []any{nodeID}, nil
}

func (s *savedSearchesConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	count, err := s.db.SavedSearches().CountSavedSearchesByOrgOrUser(ctx, s.userID, s.orgID)
	if err != nil {
		return 0, err
	}

	return int32(count), nil
}

func (s *savedSearchesConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*savedSearchResolver, error) {
	allSavedSearches, err := s.db.SavedSearches().ListSavedSearchesByOrgOrUser(ctx, s.userID, s.orgID, args)
	if err != nil {
		return nil, err
	}

	var savedSearches []*savedSearchResolver
	for _, savedSearch := range allSavedSearches {
		savedSearches = append(savedSearches, &savedSearchResolver{db: s.db, s: *savedSearch})
	}

	return savedSearches, nil
}

func (r *schemaResolver) SendSavedSearchTestNotification(ctx context.Context, args *struct {
	ID graphql.ID
}) (*EmptyResponse, error) {
	return &EmptyResponse{}, nil
}

func (r *schemaResolver) CreateSavedSearch(ctx context.Context, args *struct {
	Description string
	Query       string
	NotifyOwner bool
	NotifySlack bool
	OrgID       *graphql.ID
	UserID      *graphql.ID
}) (*savedSearchResolver, error) {
	var userID, orgID *int32
	// 🚨 SECURITY: Make sure the current user has permission to create a saved search for the specified user or org.
	if args.UserID != nil {
		u, err := unmarshalSavedSearchID(*args.UserID)
		if err != nil {
			return nil, err
		}
		userID = &u
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, u); err != nil {
			return nil, err
		}
	} else if args.OrgID != nil {
		o, err := unmarshalSavedSearchID(*args.OrgID)
		if err != nil {
			return nil, err
		}
		orgID = &o
		if err := auth.CheckOrgAccessOrSiteAdmin(ctx, r.db, o); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("failed to create saved search: no Org ID or User ID associated with saved search")
	}

	if !queryHasPatternType(args.Query) {
		return nil, errMissingPatternType
	}

	ss, err := r.db.SavedSearches().Create(ctx, &types.SavedSearch{
		Description: args.Description,
		Query:       args.Query,
		Notify:      args.NotifyOwner,
		NotifySlack: args.NotifySlack,
		UserID:      userID,
		OrgID:       orgID,
	})
	if err != nil {
		return nil, err
	}

	return r.toSavedSearchResolver(*ss), nil
}

func (r *schemaResolver) UpdateSavedSearch(ctx context.Context, args *struct {
	ID          graphql.ID
	Description string
	Query       string
	NotifyOwner bool
	NotifySlack bool
	OrgID       *graphql.ID
	UserID      *graphql.ID
}) (*savedSearchResolver, error) {
	id, err := unmarshalSavedSearchID(args.ID)
	if err != nil {
		return nil, err
	}

	old, err := r.db.SavedSearches().GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "fetch old saved search")
	}

	// 🚨 SECURITY: Make sure the current user has permission to update a saved search for the specified user or org.
	if old.Config.UserID != nil {
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, *old.Config.UserID); err != nil {
			return nil, err
		}
	} else if old.Config.OrgID != nil {
		if err := auth.CheckOrgAccessOrSiteAdmin(ctx, r.db, *old.Config.OrgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("failed to update saved search: no Org ID or User ID associated with saved search")
	}

	if !queryHasPatternType(args.Query) {
		return nil, errMissingPatternType
	}

	ss, err := r.db.SavedSearches().Update(ctx, &types.SavedSearch{
		ID:          id,
		Description: args.Description,
		Query:       args.Query,
		Notify:      args.NotifyOwner,
		NotifySlack: args.NotifySlack,
		UserID:      old.Config.UserID,
		OrgID:       old.Config.OrgID,
	})
	if err != nil {
		return nil, err
	}

	return r.toSavedSearchResolver(*ss), nil
}

func (r *schemaResolver) DeleteSavedSearch(ctx context.Context, args *struct {
	ID graphql.ID
}) (*EmptyResponse, error) {
	id, err := unmarshalSavedSearchID(args.ID)
	if err != nil {
		return nil, err
	}
	ss, err := r.db.SavedSearches().GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	// 🚨 SECURITY: Make sure the current user has permission to delete a saved search for the specified user or org.
	if ss.Config.UserID != nil {
		if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, *ss.Config.UserID); err != nil {
			return nil, err
		}
	} else if ss.Config.OrgID != nil {
		if err := auth.CheckOrgAccessOrSiteAdmin(ctx, r.db, *ss.Config.OrgID); err != nil {
			return nil, err
		}
	} else {
		return nil, errors.New("failed to delete saved search: no Org ID or User ID associated with saved search")
	}
	err = r.db.SavedSearches().Delete(ctx, id)
	if err != nil {
		return nil, err
	}
	return &EmptyResponse{}, nil
}

var patternType = lazyregexp.New(`(?i)\bpatternType:(literal|regexp|structural|standard)\b`)

func queryHasPatternType(query string) bool {
	return patternType.Match([]byte(query))
}

var errMissingPatternType = errors.New("a `patternType:` filter is required in the query for all saved searches. `patternType` can be \"standard\", \"literal\", \"regexp\" or \"structural\"")
