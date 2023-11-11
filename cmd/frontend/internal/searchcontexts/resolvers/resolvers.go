package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func NewResolver(db database.DB) graphqlbackend.SearchContextsResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db database.DB
}

func (r *Resolver) NodeResolvers() map[string]graphqlbackend.NodeByIDFunc {
	return map[string]graphqlbackend.NodeByIDFunc{
		"SearchContext": func(ctx context.Context, id graphql.ID) (graphqlbackend.Node, error) {
			return r.SearchContextByID(ctx, id)
		},
	}
}

func marshalSearchContextID(searchContextSpec string) graphql.ID {
	return relay.MarshalID("SearchContext", searchContextSpec)
}

func unmarshalSearchContextID(id graphql.ID) (spec string, err error) {
	err = relay.UnmarshalSpec(id, &spec)
	return
}

func marshalSearchContextCursor(cursor int32) string {
	return string(relay.MarshalID(graphqlbackend.SearchContextCursorKind, cursor))
}

func (r *Resolver) SearchContextsToResolvers(searchContexts []*types.SearchContext) []graphqlbackend.SearchContextResolver {
	searchContextResolvers := make([]graphqlbackend.SearchContextResolver, len(searchContexts))
	for idx, searchContext := range searchContexts {
		searchContextResolvers[idx] = &searchContextResolver{searchContext, r.db}
	}
	return searchContextResolvers
}

func (r *Resolver) SearchContextBySpec(ctx context.Context, args graphqlbackend.SearchContextBySpecArgs) (graphqlbackend.SearchContextResolver, error) {
	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, args.Spec)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func (r *Resolver) CreateSearchContext(ctx context.Context, args graphqlbackend.CreateSearchContextArgs) (_ graphqlbackend.SearchContextResolver, err error) {
	var namespaceUserID, namespaceOrgID int32
	if args.SearchContext.Namespace != nil {
		err := graphqlbackend.UnmarshalNamespaceID(*args.SearchContext.Namespace, &namespaceUserID, &namespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	var repositoryRevisions []*types.SearchContextRepositoryRevisions
	if len(args.Repositories) > 0 {
		repositoryRevisions, err = r.repositoryRevisionsFromInputArgs(ctx, args.Repositories)
		if err != nil {
			return nil, err
		}
	}

	searchContext, err := searchcontexts.CreateSearchContextWithRepositoryRevisions(
		ctx,
		r.db,
		&types.SearchContext{
			Name:            args.SearchContext.Name,
			Description:     args.SearchContext.Description,
			Public:          args.SearchContext.Public,
			NamespaceUserID: namespaceUserID,
			NamespaceOrgID:  namespaceOrgID,
			Query:           args.SearchContext.Query,
		},
		repositoryRevisions,
	)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func (r *Resolver) UpdateSearchContext(ctx context.Context, args graphqlbackend.UpdateSearchContextArgs) (graphqlbackend.SearchContextResolver, error) {
	searchContextSpec, err := unmarshalSearchContextID(args.ID)
	if err != nil {
		return nil, err
	}

	var repositoryRevisions []*types.SearchContextRepositoryRevisions
	if len(args.Repositories) > 0 {
		repositoryRevisions, err = r.repositoryRevisionsFromInputArgs(ctx, args.Repositories)
		if err != nil {
			return nil, err
		}
	}

	original, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	updated := original // inherits the ID
	updated.Name = args.SearchContext.Name
	updated.Description = args.SearchContext.Description
	updated.Public = args.SearchContext.Public
	updated.Query = args.SearchContext.Query

	searchContext, err := searchcontexts.UpdateSearchContextWithRepositoryRevisions(
		ctx,
		r.db,
		updated,
		repositoryRevisions,
	)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func (r *Resolver) repositoryRevisionsFromInputArgs(ctx context.Context, args []graphqlbackend.SearchContextRepositoryRevisionsInputArgs) ([]*types.SearchContextRepositoryRevisions, error) {
	repoIDs := make([]api.RepoID, 0, len(args))
	for _, repository := range args {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(repository.RepositoryID)
		if err != nil {
			return nil, err
		}
		repoIDs = append(repoIDs, repoID)
	}
	idToRepo, err := r.db.Repos().GetReposSetByIDs(ctx, repoIDs...)
	if err != nil {
		return nil, err
	}

	repositoryRevisions := make([]*types.SearchContextRepositoryRevisions, 0, len(args))
	for _, repository := range args {
		repoID, err := graphqlbackend.UnmarshalRepositoryID(repository.RepositoryID)
		if err != nil {
			return nil, err
		}
		repo, ok := idToRepo[repoID]
		if !ok {
			return nil, errors.Errorf("cannot find repo with id: %q", repository.RepositoryID)
		}
		repositoryRevisions = append(repositoryRevisions, &types.SearchContextRepositoryRevisions{
			Repo:      types.MinimalRepo{ID: repo.ID, Name: repo.Name},
			Revisions: repository.Revisions,
		})
	}
	return repositoryRevisions, nil
}

func (r *Resolver) DeleteSearchContext(ctx context.Context, args graphqlbackend.DeleteSearchContextArgs) (*graphqlbackend.EmptyResponse, error) {
	searchContextSpec, err := unmarshalSearchContextID(args.ID)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	err = searchcontexts.DeleteSearchContext(ctx, r.db, searchContext)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) CreateSearchContextStar(ctx context.Context, args graphqlbackend.CreateSearchContextStarArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Make sure the current user has permission to star the search context.
	userID, err := graphqlbackend.UnmarshalUserID(args.UserID)
	if err != nil {
		return nil, err
	}

	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	searchContextSpec, err := unmarshalSearchContextID(args.SearchContextID)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	err = searchcontexts.CreateSearchContextStarForUser(ctx, r.db, searchContext, userID)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) DeleteSearchContextStar(ctx context.Context, args graphqlbackend.DeleteSearchContextStarArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Make sure the current user has permission to star the search context.
	userID, err := graphqlbackend.UnmarshalUserID(args.UserID)
	if err != nil {
		return nil, err
	}

	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	searchContextSpec, err := unmarshalSearchContextID(args.SearchContextID)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	err = searchcontexts.DeleteSearchContextStarForUser(ctx, r.db, searchContext, userID)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) SetDefaultSearchContext(ctx context.Context, args graphqlbackend.SetDefaultSearchContextArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Make sure the current user has permission to set the search context as default.
	userID, err := graphqlbackend.UnmarshalUserID(args.UserID)
	if err != nil {
		return nil, err
	}

	if err := auth.CheckSiteAdminOrSameUser(ctx, r.db, userID); err != nil {
		return nil, err
	}

	searchContextSpec, err := unmarshalSearchContextID(args.SearchContextID)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	err = searchcontexts.SetDefaultSearchContextForUser(ctx, r.db, searchContext, userID)
	if err != nil {
		return nil, err
	}

	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) DefaultSearchContext(ctx context.Context) (graphqlbackend.SearchContextResolver, error) {
	searchContext, err := r.db.SearchContexts().GetDefaultSearchContextForCurrentUser(ctx)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func unmarshalSearchContextCursor(cursor *string) (int32, error) {
	var after int32
	if cursor == nil {
		after = 0
	} else {
		err := relay.UnmarshalSpec(graphql.ID(*cursor), &after)
		if err != nil {
			return -1, err
		}
	}
	return after, nil
}

func (r *Resolver) SearchContexts(ctx context.Context, args *graphqlbackend.ListSearchContextsArgs) (graphqlbackend.SearchContextConnectionResolver, error) {
	orderBy := database.SearchContextsOrderBySpec
	if args.OrderBy == graphqlbackend.SearchContextsOrderByUpdatedAt {
		orderBy = database.SearchContextsOrderByUpdatedAt
	}

	// Request one extra to determine if there are more pages
	newArgs := *args
	newArgs.First += 1

	var namespaceName string
	var searchContextName string
	if newArgs.Query != nil {
		parsedSearchContextSpec := searchcontexts.ParseSearchContextSpec(*newArgs.Query)
		searchContextName = parsedSearchContextSpec.SearchContextName
		namespaceName = parsedSearchContextSpec.NamespaceName
	}

	afterCursor, err := unmarshalSearchContextCursor(newArgs.After)
	if err != nil {
		return nil, err
	}

	namespaceUserIDs := []int32{}
	namespaceOrgIDs := []int32{}
	noNamespace := false
	for _, namespace := range args.Namespaces {
		if namespace == nil {
			noNamespace = true
		} else {
			var namespaceUserID, namespaceOrgID int32
			err := graphqlbackend.UnmarshalNamespaceID(*namespace, &namespaceUserID, &namespaceOrgID)
			if err != nil {
				return nil, err
			}
			if namespaceUserID != 0 {
				namespaceUserIDs = append(namespaceUserIDs, namespaceUserID)
			}
			if namespaceOrgID != 0 {
				namespaceOrgIDs = append(namespaceOrgIDs, namespaceOrgID)
			}
		}
	}

	opts := database.ListSearchContextsOptions{
		NamespaceName:     namespaceName,
		Name:              searchContextName,
		NamespaceUserIDs:  namespaceUserIDs,
		NamespaceOrgIDs:   namespaceOrgIDs,
		NoNamespace:       noNamespace,
		OrderBy:           orderBy,
		OrderByDescending: args.Descending,
	}

	searchContextsStore := r.db.SearchContexts()
	pageOpts := database.ListSearchContextsPageOptions{First: newArgs.First, After: afterCursor}
	searchContexts, err := searchContextsStore.ListSearchContexts(ctx, pageOpts, opts)
	if err != nil {
		return nil, err
	}

	count, err := searchContextsStore.CountSearchContexts(ctx, opts)
	if err != nil {
		return nil, err
	}

	hasNextPage := false
	if len(searchContexts) == int(args.First)+1 {
		hasNextPage = true
		searchContexts = searchContexts[:len(searchContexts)-1]
	}

	return &searchContextConnectionResolver{
		afterCursor:    afterCursor,
		searchContexts: r.SearchContextsToResolvers(searchContexts),
		totalCount:     count,
		hasNextPage:    hasNextPage,
	}, nil
}

func (r *Resolver) IsSearchContextAvailable(ctx context.Context, args graphqlbackend.IsSearchContextAvailableArgs) (bool, error) {
	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, args.Spec)
	if err != nil {
		return false, err
	}

	if searchcontexts.IsInstanceLevelSearchContext(searchContext) {
		// Instance-level search contexts are available to everyone
		return true, nil
	}

	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() {
		return false, nil
	}

	if searchContext.NamespaceUserID != 0 {
		// Is search context created by the current user
		return a.UID == searchContext.NamespaceUserID, nil
	} else {
		// Is search context created by one of the users' organizations
		orgs, err := r.db.Orgs().GetByUserID(ctx, a.UID)
		if err != nil {
			return false, err
		}
		for _, org := range orgs {
			if org.ID == searchContext.NamespaceOrgID {
				return true, nil
			}
		}
		return false, nil
	}
}

func (r *Resolver) SearchContextByID(ctx context.Context, id graphql.ID) (graphqlbackend.SearchContextResolver, error) {
	searchContextSpec, err := unmarshalSearchContextID(id)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	return &searchContextResolver{searchContext, r.db}, nil
}

type searchContextResolver struct {
	sc *types.SearchContext
	db database.DB
}

func (r *searchContextResolver) ID() graphql.ID {
	return marshalSearchContextID(searchcontexts.GetSearchContextSpec(r.sc))
}

func (r *searchContextResolver) Name() string {
	return r.sc.Name
}

func (r *searchContextResolver) Description() string {
	return r.sc.Description
}

func (r *searchContextResolver) Public() bool {
	return r.sc.Public
}

func (r *searchContextResolver) AutoDefined() bool {
	return searchcontexts.IsAutoDefinedSearchContext(r.sc)
}

func (r *searchContextResolver) Spec() string {
	return searchcontexts.GetSearchContextSpec(r.sc)
}

func (r *searchContextResolver) UpdatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.sc.UpdatedAt}
}

func (r *searchContextResolver) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	if r.sc.NamespaceUserID != 0 {
		n, err := graphqlbackend.NamespaceByID(ctx, r.db, graphqlbackend.MarshalUserID(r.sc.NamespaceUserID))
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: n}, nil
	}
	if r.sc.NamespaceOrgID != 0 {
		n, err := graphqlbackend.NamespaceByID(ctx, r.db, graphqlbackend.MarshalOrgID(r.sc.NamespaceOrgID))
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.NamespaceResolver{Namespace: n}, nil
	}
	return nil, nil
}

func (r *searchContextResolver) ViewerCanManage(ctx context.Context) bool {
	hasWriteAccess := searchcontexts.ValidateSearchContextWriteAccessForCurrentUser(ctx, r.db, r.sc.NamespaceUserID, r.sc.NamespaceOrgID, r.sc.Public) == nil
	return !searchcontexts.IsAutoDefinedSearchContext(r.sc) && hasWriteAccess
}

func (r *searchContextResolver) ViewerHasAsDefault(ctx context.Context) bool {
	return r.sc.Default
}

func (r *searchContextResolver) ViewerHasStarred(ctx context.Context) bool {
	return r.sc.Starred
}

func (r *searchContextResolver) Repositories(ctx context.Context) ([]graphqlbackend.SearchContextRepositoryRevisionsResolver, error) {
	if searchcontexts.IsAutoDefinedSearchContext(r.sc) {
		return []graphqlbackend.SearchContextRepositoryRevisionsResolver{}, nil
	}

	repoRevs, err := r.db.SearchContexts().GetSearchContextRepositoryRevisions(ctx, r.sc.ID)
	if err != nil {
		return nil, err
	}

	searchContextRepositories := make([]graphqlbackend.SearchContextRepositoryRevisionsResolver, len(repoRevs))
	for idx, repoRev := range repoRevs {
		searchContextRepositories[idx] = &searchContextRepositoryRevisionsResolver{graphqlbackend.NewRepositoryResolver(r.db, gitserver.NewClient("graphql.searchcontext.repositories"), repoRev.Repo.ToRepo()), repoRev.Revisions}
	}
	return searchContextRepositories, nil
}

func (r *searchContextResolver) Query() string {
	return r.sc.Query
}

type searchContextConnectionResolver struct {
	afterCursor    int32
	searchContexts []graphqlbackend.SearchContextResolver
	totalCount     int32
	hasNextPage    bool
}

func (s *searchContextConnectionResolver) Nodes() []graphqlbackend.SearchContextResolver {
	return s.searchContexts
}

func (s *searchContextConnectionResolver) TotalCount() int32 {
	return s.totalCount
}

func (s *searchContextConnectionResolver) PageInfo() *graphqlutil.PageInfo {
	if len(s.searchContexts) == 0 || !s.hasNextPage {
		return graphqlutil.HasNextPage(false)
	}
	// The after value (offset) for the next page is computed from the current after value + the number of retrieved search contexts
	return graphqlutil.NextPageCursor(marshalSearchContextCursor(s.afterCursor + int32(len(s.searchContexts))))
}

type searchContextRepositoryRevisionsResolver struct {
	repository *graphqlbackend.RepositoryResolver
	revisions  []string
}

func (r *searchContextRepositoryRevisionsResolver) Repository() *graphqlbackend.RepositoryResolver {
	return r.repository
}

func (r *searchContextRepositoryRevisionsResolver) Revisions() []string {
	return r.revisions
}
