package graphqlbackend

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type searchContextsOrderBy string

const (
	searchContextCursorKind                              = "SearchContextCursor"
	searchContextsOrderByUpdatedAt searchContextsOrderBy = "SEARCH_CONTEXT_UPDATED_AT"
	searchContextsOrderBySpec      searchContextsOrderBy = "SEARCH_CONTEXT_SPEC"
)

type searchContextResolver struct {
	sc *types.SearchContext
	db dbutil.DB
}

type searchContextInputArgs struct {
	Name        string
	Description string
	Public      bool
	Namespace   *graphql.ID
}

type searchContextEditInputArgs struct {
	Name        string
	Description string
	Public      bool
}

type searchContextRepositoryRevisionsInputArgs struct {
	RepositoryID graphql.ID
	Revisions    []string
}

type createSearchContextArgs struct {
	SearchContext searchContextInputArgs
	Repositories  []searchContextRepositoryRevisionsInputArgs
}

type updateSearchContextArgs struct {
	ID            graphql.ID
	SearchContext searchContextEditInputArgs
	Repositories  []searchContextRepositoryRevisionsInputArgs
}

type searchContextRepositoryRevisionsResolver struct {
	repository *RepositoryResolver
	revisions  []string
}

func (r *searchContextRepositoryRevisionsResolver) Repository(ctx context.Context) *RepositoryResolver {
	return r.repository
}

func (r *searchContextRepositoryRevisionsResolver) Revisions(ctx context.Context) []string {
	return r.revisions
}

type listSearchContextsArgs struct {
	First      int32
	After      *string
	Query      *string
	Namespaces []*graphql.ID
	OrderBy    searchContextsOrderBy
	Descending bool
}

type searchContextConnection struct {
	afterCursor    int32
	searchContexts []*searchContextResolver
	totalCount     int32
	hasNextPage    bool
}

func (s *searchContextConnection) Nodes(ctx context.Context) ([]*searchContextResolver, error) {
	return s.searchContexts, nil
}

func (s *searchContextConnection) TotalCount(ctx context.Context) (int32, error) {
	return s.totalCount, nil
}

func (s *searchContextConnection) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if len(s.searchContexts) == 0 || !s.hasNextPage {
		return graphqlutil.HasNextPage(false), nil
	}
	// The after value (offset) for the next page is computed from the current after value + the number of retrieved search contexts
	return graphqlutil.NextPageCursor(marshalSearchContextCursor(s.afterCursor + int32(len(s.searchContexts)))), nil
}

func marshalSearchContextID(searchContextSpec string) graphql.ID {
	return relay.MarshalID("SearchContext", searchContextSpec)
}

func unmarshalSearchContextID(id graphql.ID) (spec string, err error) {
	err = relay.UnmarshalSpec(id, &spec)
	return
}

func marshalSearchContextCursor(cursor int32) string {
	return string(relay.MarshalID(searchContextCursorKind, cursor))
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

func (r *searchContextResolver) UpdatedAt(ctx context.Context) DateTime {
	return DateTime{Time: r.sc.UpdatedAt}
}

func (r *searchContextResolver) Namespace(ctx context.Context) (*NamespaceResolver, error) {
	if r.sc.NamespaceUserID != 0 {
		n, err := NamespaceByID(ctx, r.db, MarshalUserID(r.sc.NamespaceUserID))
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}
	if r.sc.NamespaceOrgID != 0 {
		n, err := NamespaceByID(ctx, r.db, MarshalOrgID(r.sc.NamespaceOrgID))
		if err != nil {
			return nil, err
		}
		return &NamespaceResolver{n}, nil
	}
	return nil, nil
}

func (r *searchContextResolver) ViewerCanManage(ctx context.Context) bool {
	hasWriteAccess := searchcontexts.ValidateSearchContextWriteAccessForCurrentUser(ctx, r.db, r.sc.NamespaceUserID, r.sc.NamespaceOrgID, r.sc.Public) == nil
	return !searchcontexts.IsAutoDefinedSearchContext(r.sc) && hasWriteAccess
}

func (r *searchContextResolver) Repositories(ctx context.Context) ([]*searchContextRepositoryRevisionsResolver, error) {
	if searchcontexts.IsAutoDefinedSearchContext(r.sc) {
		return []*searchContextRepositoryRevisionsResolver{}, nil
	}

	repoRevs, err := database.SearchContexts(r.db).GetSearchContextRepositoryRevisions(ctx, r.sc.ID)
	if err != nil {
		return nil, err
	}

	searchContextRepositories := make([]*searchContextRepositoryRevisionsResolver, len(repoRevs))
	for idx, repoRev := range repoRevs {
		searchContextRepositories[idx] = &searchContextRepositoryRevisionsResolver{NewRepositoryResolver(r.db, repoRev.Repo.ToRepo()), repoRev.Revisions}
	}
	return searchContextRepositories, nil
}

func (r *schemaResolver) AutoDefinedSearchContexts(ctx context.Context) ([]*searchContextResolver, error) {
	searchContexts, err := searchcontexts.GetAutoDefinedSearchContexts(ctx, r.db)
	if err != nil {
		return nil, err
	}
	return searchContextsToResolvers(searchContexts, r.db), nil
}

func (r *schemaResolver) SearchContextBySpec(ctx context.Context, args *struct {
	Spec string
}) (*searchContextResolver, error) {
	searchContext, err := searchcontexts.ResolveSearchContextSpec(ctx, r.db, args.Spec)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func (r *schemaResolver) CreateSearchContext(ctx context.Context, args createSearchContextArgs) (*searchContextResolver, error) {
	var namespaceUserID, namespaceOrgID int32
	if args.SearchContext.Namespace != nil {
		err := UnmarshalNamespaceID(*args.SearchContext.Namespace, &namespaceUserID, &namespaceOrgID)
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

func (r *schemaResolver) UpdateSearchContext(ctx context.Context, args updateSearchContextArgs) (*searchContextResolver, error) {
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

func (r *schemaResolver) repositoryRevisionsFromInputArgs(ctx context.Context, args []searchContextRepositoryRevisionsInputArgs) ([]*types.SearchContextRepositoryRevisions, error) {
	repositoryRevisions := make([]*types.SearchContextRepositoryRevisions, 0, len(args))
	for _, repository := range args {
		repoResolver, err := r.repositoryByID(ctx, repository.RepositoryID)
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

func (r *schemaResolver) DeleteSearchContext(ctx context.Context, args struct {
	ID graphql.ID
}) (*EmptyResponse, error) {
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

	return &EmptyResponse{}, nil
}

func (r *schemaResolver) SearchContexts(ctx context.Context, args *listSearchContextsArgs) (*searchContextConnection, error) {
	orderBy := database.SearchContextsOrderBySpec
	if args.OrderBy == searchContextsOrderByUpdatedAt {
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
			err := UnmarshalNamespaceID(*namespace, &namespaceUserID, &namespaceOrgID)
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

	return &searchContextConnection{
		afterCursor:    afterCursor,
		searchContexts: searchContextsToResolvers(searchContexts, r.db),
		totalCount:     count,
		hasNextPage:    hasNextPage,
	}, nil
}

func searchContextsToResolvers(searchContexts []*types.SearchContext, db dbutil.DB) []*searchContextResolver {
	searchContextResolvers := make([]*searchContextResolver, len(searchContexts))
	for idx, searchContext := range searchContexts {
		searchContextResolvers[idx] = &searchContextResolver{searchContext, db}
	}
	return searchContextResolvers
}

func (r *schemaResolver) IsSearchContextAvailable(ctx context.Context, args struct {
	Spec string
}) (bool, error) {
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

func resolveVersionContext(versionContext string) (*schema.VersionContext, error) {
	for _, vc := range conf.Get().ExperimentalFeatures.VersionContexts {
		if vc.Name == versionContext {
			return vc, nil
		}
	}
	return nil, errors.New("version context not found")
}

func (r *schemaResolver) ConvertVersionContextToSearchContext(ctx context.Context, args *struct {
	Name string
}) (*searchContextResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, errors.New("converting a version context to a search context is limited to site admins")
	}
	versionContext, err := resolveVersionContext(args.Name)
	if err != nil {
		return nil, err
	}

	searchContext, err := searchcontexts.ConvertVersionContextToSearchContext(ctx, r.db, versionContext)
	if err != nil {
		return nil, err
	}
	return &searchContextResolver{searchContext, r.db}, nil
}

func (r *schemaResolver) SearchContextByID(ctx context.Context, id graphql.ID) (*searchContextResolver, error) {
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
