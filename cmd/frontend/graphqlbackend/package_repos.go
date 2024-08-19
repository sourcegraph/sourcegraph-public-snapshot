package graphqlbackend

import (
	"context"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type PackageRepoReferenceConnectionArgs struct {
	gqlutil.ConnectionArgs
	After *string
	Kind  *string
	Name  *string
}

var externalServiceToPackageSchemeMap = map[string]string{
	extsvc.KindJVMPackages:    dependencies.JVMPackagesScheme,
	extsvc.KindNpmPackages:    dependencies.NpmPackagesScheme,
	extsvc.KindGoPackages:     dependencies.GoPackagesScheme,
	extsvc.KindPythonPackages: dependencies.PythonPackagesScheme,
	extsvc.KindRustPackages:   dependencies.RustPackagesScheme,
	extsvc.KindRubyPackages:   dependencies.RubyPackagesScheme,
}

var packageSchemeToExternalServiceMap = map[string]string{
	dependencies.JVMPackagesScheme:    extsvc.KindJVMPackages,
	dependencies.NpmPackagesScheme:    extsvc.KindNpmPackages,
	dependencies.GoPackagesScheme:     extsvc.KindGoPackages,
	dependencies.PythonPackagesScheme: extsvc.KindPythonPackages,
	dependencies.RustPackagesScheme:   extsvc.KindRustPackages,
	dependencies.RubyPackagesScheme:   extsvc.KindRubyPackages,
}

func (r *schemaResolver) PackageRepoReferences(ctx context.Context, args *PackageRepoReferenceConnectionArgs) (_ *packageRepoReferenceConnectionResolver, err error) {
	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	opts := dependencies.ListDependencyReposOpts{
		IncludeBlocked: true,
	}

	if args.Kind != nil {
		packageScheme, ok := externalServiceToPackageSchemeMap[*args.Kind]
		if !ok {
			return nil, errors.Errorf("unknown package scheme %q", *args.Kind)
		}
		opts.Scheme = packageScheme
	}

	if args.Name != nil {
		opts.Name = reposource.PackageName(*args.Name)
	}

	opts.Limit = int(args.GetFirst())

	if args.After != nil {
		if err := relay.UnmarshalSpec(graphql.ID(*args.After), &opts.After); err != nil {
			return nil, err
		}
	}

	deps, total, hasMore, err := depsService.ListPackageRepoRefs(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &packageRepoReferenceConnectionResolver{r.db, deps, hasMore, total}, err
}

type packageRepoReferenceConnectionResolver struct {
	db      database.DB
	deps    []dependencies.PackageRepoReference
	hasMore bool
	total   int
}

func (r *packageRepoReferenceConnectionResolver) Nodes(ctx context.Context) ([]*packageRepoReferenceResolver, error) {
	once := sync.OnceValues(func() (map[api.RepoName]*types.Repo, error) {
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

func (r *packageRepoReferenceConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	if len(r.deps) == 0 || !r.hasMore {
		return gqlutil.HasNextPage(false), nil
	}

	next := r.deps[len(r.deps)-1].ID
	cursor := string(relay.MarshalID("PackageRepoReference", next))
	return gqlutil.NextPageCursor(cursor), nil
}

type packageRepoReferenceVersionConnectionResolver struct {
	versions []dependencies.PackageRepoRefVersion
	hasMore  bool
	total    int
}

func (r *packageRepoReferenceVersionConnectionResolver) Nodes(ctx context.Context) (vs []*packageRepoReferenceVersionResolver) {
	for _, version := range r.versions {
		vs = append(vs, &packageRepoReferenceVersionResolver{
			version: version,
		})
	}
	return
}

func (r *packageRepoReferenceVersionConnectionResolver) TotalCount(ctx context.Context) (int32, error) {
	return int32(r.total), nil
}

func (r *packageRepoReferenceVersionConnectionResolver) PageInfo(ctx context.Context) (*gqlutil.PageInfo, error) {
	if len(r.versions) == 0 || !r.hasMore {
		return gqlutil.HasNextPage(false), nil
	}

	next := r.versions[len(r.versions)-1].ID
	cursor := string(relay.MarshalID("PackageRepoReferenceVersion", next))
	return gqlutil.NextPageCursor(cursor), nil
}

type packageRepoReferenceResolver struct {
	db       database.DB
	dep      dependencies.PackageRepoReference
	allRepos func() (map[api.RepoName]*types.Repo, error)
}

func (r *packageRepoReferenceResolver) ID() graphql.ID {
	return relay.MarshalID("PackageRepoReference", r.dep.ID)
}

func (r *packageRepoReferenceResolver) Kind() string {
	return packageSchemeToExternalServiceMap[r.dep.Scheme]
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

func (r *packageRepoReferenceResolver) Blocked() bool {
	return r.dep.Blocked
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
		return NewRepositoryResolver(r.db, gitserver.NewClient("graphql.packagerepo"), repo), nil
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
