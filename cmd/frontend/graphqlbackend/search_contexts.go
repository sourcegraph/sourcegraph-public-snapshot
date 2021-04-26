package graphqlbackend

import (
	"context"
	"errors"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/search/searchcontexts"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

type namespaceFilterType string

const (
	namespaceFilterTypeInstance  namespaceFilterType = "INSTANCE"
	namespaceFilterTypeNamespace namespaceFilterType = "NAMESPACE"
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

type searchContextRepositoryRevisionsInputArgs struct {
	RepositoryID graphql.ID
	Revisions    []string
}

type createSearchContextArgs struct {
	SearchContext searchContextInputArgs
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
	First               int32
	After               *string
	Query               *string
	NamespaceFilterType *namespaceFilterType
	Namespace           *graphql.ID
}

type searchContextConnection struct {
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
	return graphqlutil.NextPageCursor(string(s.searchContexts[len(s.searchContexts)-1].DatabaseID())), nil
}

func marshalSearchContextID(searchContextSpec string) graphql.ID {
	return relay.MarshalID("SearchContext", searchContextSpec)
}

func unmarshalSearchContextID(id graphql.ID) (spec string, err error) {
	err = relay.UnmarshalSpec(id, &spec)
	return
}

func marshalSearchContextDatabaseID(id int64) graphql.ID {
	return relay.MarshalID("SearchContext", id)
}

func unmarshalSearchContextCursor(after *string) (int64, error) {
	var id int64
	if after == nil {
		id = 0
	} else {
		err := relay.UnmarshalSpec(graphql.ID(*after), &id)
		if err != nil {
			return -1, err
		}
	}
	return id, nil
}

func (r *searchContextResolver) ID() graphql.ID {
	return marshalSearchContextID(searchcontexts.GetSearchContextSpec(r.sc))
}

func (r *searchContextResolver) DatabaseID() graphql.ID {
	return marshalSearchContextDatabaseID(r.sc.ID)
}

func (r *searchContextResolver) Description(ctx context.Context) string {
	return r.sc.Description
}

func (r *searchContextResolver) AutoDefined(ctx context.Context) bool {
	return searchcontexts.IsAutoDefinedSearchContext(r.sc)
}

func (r *searchContextResolver) Spec(ctx context.Context) string {
	return searchcontexts.GetSearchContextSpec(r.sc)
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

func (r *schemaResolver) CreateSearchContext(ctx context.Context, args createSearchContextArgs) (*searchContextResolver, error) {
	var namespaceUserID, namespaceOrgID int32
	if args.SearchContext.Namespace != nil {
		err := UnmarshalNamespaceID(*args.SearchContext.Namespace, &namespaceUserID, &namespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	repositoryRevisions := make([]*types.SearchContextRepositoryRevisions, 0, len(args.Repositories))
	for _, repository := range args.Repositories {
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
	var namespaceFilter namespaceFilterType
	if args.NamespaceFilterType != nil {
		namespaceFilter = *args.NamespaceFilterType
	}

	if args.Namespace != nil && namespaceFilter != namespaceFilterTypeNamespace {
		return nil, errors.New("namespace can only be used if namespaceFilterType is NAMESPACE")
	}
	if args.Namespace == nil && namespaceFilter == namespaceFilterTypeNamespace {
		return nil, errors.New("namespace has to be non-nil if namespaceFilterType is NAMESPACE")
	}

	// Request one extra to determine if there are more pages
	newArgs := *args
	newArgs.First += 1

	// TODO(rok): Parse the query into namespace and search context name components
	var searchContextName string
	if newArgs.Query != nil {
		searchContextName = *newArgs.Query
	}

	afterCursor, err := unmarshalSearchContextCursor(newArgs.After)
	if err != nil {
		return nil, err
	}

	opts := database.ListSearchContextsOptions{Name: searchContextName, NoNamespace: namespaceFilter == namespaceFilterTypeInstance}
	if newArgs.Namespace != nil {
		err := UnmarshalNamespaceID(*newArgs.Namespace, &opts.NamespaceUserID, &opts.NamespaceOrgID)
		if err != nil {
			return nil, err
		}
	}

	searchContextsStore := database.SearchContexts(r.db)
	pageOpts := database.ListSearchContextsPageOptions{First: newArgs.First, AfterID: afterCursor}
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
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
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
