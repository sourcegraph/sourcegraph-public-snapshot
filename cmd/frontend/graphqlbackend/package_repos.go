package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PackageRepoReferenceConnectionArgs struct {
	graphqlutil.ConnectionArgs
	After  *int
	Scheme *string
	Name   *string
}

func (r *schemaResolver) PackageRepoReferences(ctx context.Context, args *PackageRepoReferenceConnectionArgs) (*packageRepoReferenceConnectionResolver, error) {
	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	var opts dependencies.ListDependencyReposOpts

	if args.Scheme != nil {
		opts.Scheme = *args.Scheme
	}

	if args.Name != nil {
		opts.Name = reposource.PackageName(*args.Name)
	}

	if args.First != nil {
		opts.Limit = int(*args.First)
	}

	if args.After != nil {
		opts.After = *args.After
	}

	deps, total, err := depsService.ListPackageRepoRefs(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &packageRepoReferenceConnectionResolver{r.db, deps, total}, err
}

type packageRepoReferenceConnectionResolver struct {
	db    database.DB
	deps  []dependencies.PackageRepoReference
	total int
}

func (r *packageRepoReferenceConnectionResolver) Nodes(ctx context.Context) ([]*packageRepoReferenceResolver, error) {
	once := syncx.OnceValues(func() (map[api.RepoName]*types.Repo, error) {
		allNames := make([]string, 0, len(r.deps))
		for _, dep := range r.deps {
			name, err := dependencyRepoToRepoName(dep)
			if err != nil || string(name) == "" {
				continue
			}
			allNames = append(allNames, string(name))
		}

		repos, err := r.db.Repos().List(ctx, database.ReposListOptions{
			Names: allNames,
		})
		if err != nil {
			return nil, errors.Wrap(err, "error listing repos")
		}

		repoMappings := make(map[api.RepoName]*types.Repo, len(repos))
		for _, repo := range repos {
			repoMappings[repo.Name] = repo
		}
		return repoMappings, nil
	})

	resolvers := make([]*packageRepoReferenceResolver, 0, len(r.deps))
	for _, dep := range r.deps {
		resolvers = append(resolvers, &packageRepoReferenceResolver{r.db, dep, once})
	}

	return resolvers, nil
}

func (r *packageRepoReferenceConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.total), nil
}

func (r *packageRepoReferenceConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	if len(r.deps) == 0 {
		return graphqlutil.HasNextPage(false), nil
	}

	next := int32(r.deps[len(r.deps)-1].ID)
	return graphqlutil.EncodeIntCursor(&next), nil
}

type packageRepoReferenceResolver struct {
	db       database.DB
	dep      dependencies.PackageRepoReference
	allRepos func() (map[api.RepoName]*types.Repo, error)
}

func (r *packageRepoReferenceResolver) ID() graphql.ID {
	return relay.MarshalID("PackageRepoReference", r.dep.ID)
}

func (r *packageRepoReferenceResolver) Scheme() string {
	return r.dep.Scheme
}

func (r *packageRepoReferenceResolver) Name() string {
	return string(r.dep.Name)
}

func (r *packageRepoReferenceResolver) Versions() []*packageRepoReferenceVersionResolver {
	versions := make([]*packageRepoReferenceVersionResolver, 0, len(r.dep.Versions))
	for _, version := range r.dep.Versions {
		versions = append(versions, &packageRepoReferenceVersionResolver{version})
	}
	return versions
}

func (r *packageRepoReferenceResolver) Repository(ctx context.Context) (*RepositoryResolver, error) {
	repoName, err := dependencyRepoToRepoName(r.dep)
	if err != nil {
		return nil, err
	}

	repos, err := r.allRepos()
	if err != nil {
		return nil, err
	}

	if repo, ok := repos[repoName]; ok {
		return NewRepositoryResolver(r.db, gitserver.NewClient(), repo), nil
	}

	return nil, nil
}

type packageRepoReferenceVersionResolver struct {
	version dependencies.PackageRepoRefVersion
}

func (r *packageRepoReferenceVersionResolver) ID() graphql.ID {
	return relay.MarshalID("PackageRepoRefVersion", r.version.ID)
}

func (r *packageRepoReferenceVersionResolver) PackageRepoReferenceID() graphql.ID {
	return relay.MarshalID("PackageRepoReference", r.version.PackageRefID)
}

func (r *packageRepoReferenceVersionResolver) Version() string {
	return r.version.Version
}

func dependencyRepoToRepoName(dep dependencies.PackageRepoReference) (repoName api.RepoName, _ error) {
	switch dep.Scheme {
	case "python":
		repoName = reposource.ParsePythonPackageFromName(dep.Name).RepoName()
	case "scip-ruby":
		repoName = reposource.ParseRubyPackageFromName(dep.Name).RepoName()
	case "semanticdb":
		pkg, err := reposource.ParseMavenPackageFromName(dep.Name)
		if err != nil {
			return "", err
		}
		repoName = pkg.RepoName()
	case "npm":
		pkg, err := reposource.ParseNpmPackageFromPackageSyntax(dep.Name)
		if err != nil {
			return "", err
		}
		repoName = pkg.RepoName()
	case "rust-analyzer":
		repoName = reposource.ParseRustPackageFromName(dep.Name).RepoName()
	case "go":
		pkg, err := reposource.ParseGoDependencyFromName(dep.Name)
		if err != nil {
			return "", err
		}
		repoName = pkg.RepoName()
	}

	return repoName, nil
}
