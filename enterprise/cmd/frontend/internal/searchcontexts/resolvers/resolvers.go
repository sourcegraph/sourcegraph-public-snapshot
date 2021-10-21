package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func NewResolver(db dbutil.DB) graphqlbackend.SearchContextsResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db dbutil.DB
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

func (r *Resolver) AutoDefinedSearchContexts(ctx context.Context) ([]graphqlbackend.SearchContextResolver, error) {
	searchContexts, err := searchcontexts.GetAutoDefinedSearchContexts(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return r.SearchContextsToResolvers(searchContexts), nil
}

func (r *Resolver) SearchContextBySpec(ctx context.Context, args graphqlbackend.SearchContextBySpecArgs) (graphqlbackend.SearchContextResolver, error) {
	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, args.Spec)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func (r *Resolver) CreateSearchContext(ctx context.Context, args graphqlbackend.CreateSearchContextArgs) (graphqlbackend.SearchContextResolver, error) {
	var namespaceUserID, namespaceOrgID int32
	if args.SearchContext.Namespace != nil {
		err := graphqlbackend.UnmarshalNamespaceID(*args.SearchContext.Namespace, &namespaceUserID, &namespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	repositoryRevisions, err := r.repositoryRevisionsFromInputArgs(ctx, args.Repositories)
	if err != nil {
		return nil, err
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

	repositoryRevisions, err := r.repositoryRevisionsFromInputArgs(ctx, args.Repositories)
	if err != nil {
		return nil, err
	}

	original, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, searchContextSpec)
	if err != nil {
		return nil, err
	}

	updated := original // inherits the ID
	updated.Name = args.SearchContext.Name
	updated.Description = args.SearchContext.Description
	updated.Public = args.SearchContext.Public

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

func repositoryByID(ctx context.Context, id graphql.ID, db dbutil.DB) (*graphqlbackend.RepositoryResolver, error) {
	var repoID api.RepoID
	if err := relay.UnmarshalSpec(id, &repoID); err != nil {
		return nil, err
	}
	repo, err := database.Repos(db).Get(ctx, repoID)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewRepositoryResolver(db, repo), nil
}

func (r *Resolver) repositoryRevisionsFromInputArgs(ctx context.Context, args []graphqlbackend.SearchContextRepositoryRevisionsInputArgs) ([]*types.SearchContextRepositoryRevisions, error) {
	repositoryRevisions := make([]*types.SearchContextRepositoryRevisions, 0, len(args))
	for _, repository := range args {
		repoResolver, err := repositoryByID(ctx, repository.RepositoryID, r.db)
		if err != nil {
			return nil, err
		}
		repositoryRevisions = append(repositoryRevisions, &types.SearchContextRepositoryRevisions{
			Repo: types.RepoName{
				ID:   repoResolver.IDInt32(),
				Name: repoResolver.RepoName(),
			},
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

	searchContextsStore := database.SearchContexts(r.db)
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
		orgs, err := database.Orgs(r.db).GetByUserID(ctx, a.UID)
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
	db dbutil.DB
}

func (r *searchContextResolver) ID() graphql.ID {
	return marshalSearchContextID(searchcontexts.GetSearchContextSpec(r.sc))
}

func (r *searchContextResolver) Name(ctx context.Context) string {
	return r.sc.Name
}

func (r *searchContextResolver) Description(ctx context.Context) string {
	return r.sc.Description
}

func (r *searchContextResolver) Public(ctx context.Context) bool {
	return r.sc.Public
}

func (r *searchContextResolver) AutoDefined(ctx context.Context) bool {
	return searchcontexts.IsAutoDefinedSearchContext(r.sc)
}

func (r *searchContextResolver) Spec() string {
	return searchcontexts.GetSearchContextSpec(r.sc)
}

func (r *searchContextResolver) UpdatedAt(ctx context.Context) graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.sc.UpdatedAt}
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

func (r *searchContextResolver) Repositories(ctx context.Context) ([]graphqlbackend.SearchContextRepositoryRevisionsResolver, error) {
	if searchcontexts.IsAutoDefinedSearchContext(r.sc) {
		return []graphqlbackend.SearchContextRepositoryRevisionsResolver{}, nil
	}

	repoRevs, err := database.SearchContexts(r.db).GetSearchContextRepositoryRevisions(ctx, r.sc.ID)
	if err != nil {
		return nil, err
	}

	searchContextRepositories := make([]graphqlbackend.SearchContextRepositoryRevisionsResolver, len(repoRevs))
	for idx, repoRev := range repoRevs {
		searchContextRepositories[idx] = &searchContextRepositoryRevisionsResolver{graphqlbackend.NewRepositoryResolver(r.db, repoRev.Repo.ToRepo()), repoRev.Revisions}
	}
	return searchContextRepositories, nil
}

type searchContextConnectionResolver struct {
	afterCursor    int32
	searchContexts []graphqlbackend.SearchContextResolver
	totalCount     int32
	hasNextPage    bool
}

func (s *searchContextConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.SearchContextResolver, error) {
	return s.searchContexts, nil
}

func (s *searchContextConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return s.totalCount, nil
}

func (s *searchContextConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if len(s.searchContexts) == 0 || !s.hasNextPage {
		return graphqlutil.HasNextPage(false), nil
	}
	// The after value (offset) for the next page is computed from the current after value + the number of retrieved search contexts
	return graphqlutil.NextPageCursor(marshalSearchContextCursor(s.afterCursor + int32(len(s.searchContexts)))), nil
}

type searchContextRepositoryRevisionsResolver struct {
	repository *graphqlbackend.RepositoryResolver
	revisions  []string
}

func (r *searchContextRepositoryRevisionsResolver) Repository(ctx context.Context) *graphqlbackend.RepositoryResolver {
	return r.repository
}

func (r *searchContextRepositoryRevisionsResolver) Revisions(ctx context.Context) []string {
	return r.revisions
}
