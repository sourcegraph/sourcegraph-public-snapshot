package resolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func NewRootResolver(db database.DB) gql.CatalogRootResolver {
	return &rootResolver{db: db}
}

type rootResolver struct {
	db database.DB
}

func (r *rootResolver) Catalog(context.Context) (gql.CatalogResolver, error) {
	return &catalogResolver{db: r.db}, nil
}

func (r *rootResolver) NodeResolvers() map[string]gql.NodeByIDFunc {
	return map[string]gql.NodeByIDFunc{
		"CatalogComponent": func(ctx context.Context, id graphql.ID) (gql.Node, error) {
			components := dummyData(r.db)
			for _, c := range components {
				if c.ID() == id {
					return c, nil
				}
			}
			return nil, nil
		},
	}
}

// TODO(sqs): dummy data
func dummyData(db database.DB) []*catalogComponentResolver {
	const (
		sourceRepo   = "github.com/sourcegraph/sourcegraph"
		sourceCommit = "2ada4911722e2c812cc4f1bbfb6d5d1756891392"
	)
	components := []*catalogComponentResolver{
		{
			kind:         "SERVICE",
			name:         "frontend",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/frontend",
		},
		{
			kind:         "SERVICE",
			name:         "gitserver",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/gitserver",
		},
		{
			kind:         "SERVICE",
			name:         "repo-updater",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/repo-updater",
		},
		{
			kind:         "SERVICE",
			name:         "executor",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "enterprise/cmd/executor",
		},
		{
			kind:         "SERVICE",
			name:         "precise-code-intel-worker",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "enterprise/cmd/precise-code-intel-worker",
		},

		{
			kind:         "SERVICE",
			name:         "github-proxy",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/github-proxy",
		},

		{
			kind:         "SERVICE",
			name:         "query-runner",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/query-runner",
		},

		{
			kind:         "SERVICE",
			name:         "worker",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/worker",
		},

		{
			kind:         "SERVICE",
			name:         "server",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/server",
		},

		{
			kind:         "SERVICE",
			name:         "symbols",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/symbols",
		},

		{
			kind:         "SERVICE",
			name:         "sitemap",
			sourceRepo:   sourceRepo,
			sourceCommit: sourceCommit,
			sourcePath:   "cmd/sitemap",
		},
	}
	for _, c := range components {
		c.db = db
	}
	return components
}

type catalogResolver struct {
	db database.DB
}

func (r *catalogResolver) Components(ctx context.Context, args *gql.CatalogComponentsArgs) (gql.CatalogComponentConnectionResolver, error) {
	components := dummyData(r.db)

	var keep []gql.CatalogComponentResolver
	for _, c := range components {
		if args.Query == nil || strings.Contains(c.name, *args.Query) {
			keep = append(keep, c)
		}
	}

	return &catalogComponentConnectionResolver{
		components: keep,
	}, nil
}

type catalogComponentConnectionResolver struct {
	components []gql.CatalogComponentResolver
}

func (r *catalogComponentConnectionResolver) Nodes(ctx context.Context) ([]gql.CatalogComponentResolver, error) {
	return r.components, nil
}

func (r *catalogComponentConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(len(r.components)), nil
}

func (r *catalogComponentConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	return graphqlutil.HasNextPage(false), nil // TODO(sqs)
}

type catalogComponentResolver struct {
	kind   gql.CatalogComponentKind
	name   string
	system *string

	sourceRepo, sourceCommit, sourcePath string

	db database.DB
}

func (r *catalogComponentResolver) ID() graphql.ID {
	return relay.MarshalID("CatalogComponent", r.name) // TODO(sqs)
}

func (r *catalogComponentResolver) Kind() gql.CatalogComponentKind {
	return r.kind
}

func (r *catalogComponentResolver) Name() string {
	return r.name
}

func (r *catalogComponentResolver) Owner(context.Context) (*gql.PersonResolver, error) {
	return nil, nil
}

func (r *catalogComponentResolver) System() *string {
	return r.system
}

func (r *catalogComponentResolver) Tags() []string {
	return []string{"my-tag1", "my-tag2"}
}

func (r *catalogComponentResolver) sourceRepoResolver(ctx context.Context) (*gql.RepositoryResolver, error) {
	// 🚨 SECURITY: database.Repos.Get uses the authzFilter under the hood and
	// filters out repositories that the user doesn't have access to.
	repo, err := r.db.Repos().GetByName(ctx, api.RepoName(r.sourceRepo))
	if err != nil {
		return nil, err
	}

	return gql.NewRepositoryResolver(r.db, repo), nil
}

func (r *catalogComponentResolver) SourceLocation(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	repoResolver, err := r.sourceRepoResolver(ctx)
	if err != nil {
		return nil, err
	}
	commitResolver := gql.NewGitCommitResolver(r.db, repoResolver, api.CommitID(r.sourceCommit), nil)
	return gql.NewGitTreeEntryResolver(r.db, commitResolver, gql.CreateFileInfo(r.sourcePath, false)), nil
}
