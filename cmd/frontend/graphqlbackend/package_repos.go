package graphqlbackend

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"golang.org/x/exp/slices"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/packagerepos"
	"github.com/sourcegraph/sourcegraph/internal/syncx"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type PackageRepoReferenceConnectionArgs struct {
	graphqlutil.ConnectionArgs
	After  *string
	Scheme *string
	Name   *string
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

	var opts dependencies.ListDependencyReposOpts

	if args.Scheme != nil {
		packageScheme, ok := externalServiceToPackageSchemeMap[*args.Scheme]
		if !ok {
			return nil, errors.Errorf("unknown package scheme %q", *args.Scheme)
		}
		opts.Scheme = packageScheme
	}

	if args.Name != nil {
		opts.Name = reposource.PackageName(*args.Name)
	}

	opts.Limit = int(args.GetFirst())

	if args.After != nil {
		if opts.After, err = graphqlutil.DecodeIntCursor(args.After); err != nil {
			return nil, err
		}
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

type packageMatcher struct {
	NameMatcher *struct {
		PackageGlob string
	}
	VersionMatcher *struct {
		PackageName string
		VersionGlob string
	}
}

func (r *schemaResolver) PackageReposMatches(ctx context.Context, args struct {
	PackageReferenceKind string
	Matcher              packageMatcher
	graphqlutil.ConnectionArgs
	After *string
},
) (*packageRepoReferenceConnectionResolver, error) {
	if args.Matcher.NameMatcher == nil && args.Matcher.VersionMatcher == nil {
		return nil, errors.New("must provide either nameMatcher or versionMatcher")
	}

	if args.Matcher.NameMatcher != nil && args.Matcher.VersionMatcher != nil {
		return nil, errors.New("cannot provide both a name matcher and version matcher")
	}

	kinds := []string{args.PackageReferenceKind}

	extsvcs, err := r.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{Kinds: kinds})
	if err != nil {
		return nil, errors.Wrap(err, "failed to list external services")
	}

	if len(extsvcs) == 0 {
		return nil, errors.Newf("no external service configured of kind %q", args.PackageReferenceKind)
	}

	var (
		matcher     packagerepos.PackageMatcher
		nameToMatch string
	)
	if args.Matcher.NameMatcher != nil {
		matcher, err = packagerepos.NewPackageNameGlob(args.Matcher.NameMatcher.PackageGlob)
	} else {
		matcher, err = packagerepos.NewVersionGlob(args.Matcher.VersionMatcher.PackageName, args.Matcher.VersionMatcher.VersionGlob)
		nameToMatch = args.Matcher.VersionMatcher.PackageName
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to compile glob")
	}

	limit := int(args.GetFirst())

	var after int
	if args.After != nil {
		if after, err = graphqlutil.DecodeIntCursor(args.After); err != nil {
			return nil, err
		}
	}

	packageRepoScheme := externalServiceToPackageSchemeMap[args.PackageReferenceKind]

	depsService := dependencies.NewService(observation.NewContext(r.logger), r.db)

	matchingPkgs := make([]shared.PackageRepoReference, 0, limit)
	if args.Matcher.NameMatcher != nil {
		lastID := after

	gather:
		for limit == 0 || len(matchingPkgs) < limit {
			fmt.Println(limit == 0, len(matchingPkgs) < limit)
			pkgs, _, err := depsService.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
				Scheme: packageRepoScheme,
				After:  lastID,
				Limit:  limit,
			})
			if err != nil {
				return nil, errors.Wrap(err, "failed to list package repo references")
			}

			if len(pkgs) == 0 {
				break
			}

			lastID = pkgs[len(pkgs)-1].ID

			for _, pkg := range pkgs {
				if matcher.Matches(string(pkg.Name), "") {
					pkg.Versions = nil
					matchingPkgs = append(matchingPkgs, pkg)
				}
				if limit != 0 && len(matchingPkgs) == limit {
					break gather
				}
			}
		}
	} else {
		pkgs, _, err := depsService.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme:        packageRepoScheme,
			Name:          reposource.PackageName(nameToMatch),
			ExactNameOnly: true,
			Limit:         1,
		})
		if err != nil {
			return nil, errors.Wrap(err, "failed to list package repo references")
		}

		if len(pkgs) == 0 {
			return nil, errors.Newf("package repo reference not found for name %q", nameToMatch)
		}

		pkg := pkgs[0]
		versions := pkg.Versions[:0]
		for _, version := range pkg.Versions {
			if matcher.Matches(string(pkg.Name), version.Version) {
				versions = append(versions, version)
			}
		}
		pkg.Versions = versions
		matchingPkgs = append(matchingPkgs, pkg)
	}

	return &packageRepoReferenceConnectionResolver{
		db:   r.db,
		deps: matchingPkgs,
		// bit of a lie lol
		total: len(matchingPkgs),
	}, nil
}

func (s *schemaResolver) AddPackageRepoMatcher(ctx context.Context, args struct {
	Behaviour            string
	PackageReferenceKind string
	Matcher              packageMatcher
},
) (*EmptyResponse, error) {
	if args.Matcher.NameMatcher == nil && args.Matcher.VersionMatcher == nil {
		return nil, errors.New("must provide either nameMatcher or versionMatcher")
	}

	if args.Matcher.NameMatcher != nil && args.Matcher.VersionMatcher != nil {
		return nil, errors.New("cannot provide both a name matcher and version matcher")
	}

	extsvcs, err := s.db.ExternalServices().List(ctx, database.ExternalServicesListOptions{Kinds: []string{args.PackageReferenceKind}})
	if err != nil {
		return nil, errors.Wrap(err, "error finding matching external service")
	}

	if len(extsvcs) == 0 {
		return nil, errors.Newf("no matching external service for kind %q", args.PackageReferenceKind)
	}

	config, err := extsvcs[0].Configuration(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to fetch external service configuration")
	}

	switch args.PackageReferenceKind {
	case extsvc.KindJVMPackages:
		config := config.(*schema.JVMPackagesConnection)
		err = addMatcherToConfig(&config.Maven.Allowlist, &config.Maven.Blocklist, args.Behaviour, args.Matcher)
	default:
		return nil, errors.Newf("external service of kind %q does not support allow-/block-lists", args.PackageReferenceKind)
	}

	if err != nil {
		return nil, err
	}

	if err := s.db.ExternalServices().Upsert(ctx, extsvcs[0]); err != nil {
		return nil, errors.Wrap(err, "failed to update external service")
	}

	return &EmptyResponse{}, nil
}

var matcherAlreadyExists = errors.New("matcher already exists in config")

func addMatcherToConfig(allowlist, blocklist *[]any, behaviour string, matcher packageMatcher) error {
	// the format expected by the JSON schema, because 'anyOf'
	m := make(map[string]interface{})
	if matcher.NameMatcher != nil {
		m["packageGlob"] = matcher.NameMatcher.PackageGlob
	} else {
		m["package"] = matcher.VersionMatcher.PackageName
		m["versionGlob"] = matcher.VersionMatcher.VersionGlob
	}

	if behaviour == "BLOCK" {
		if matcherAlreadyInConfig(*blocklist, matcher) {
			return matcherAlreadyExists
		}
		*blocklist = append(*blocklist, matcher)
	} else {
		if matcherAlreadyInConfig(*blocklist, matcher) {
			return matcherAlreadyExists
		}
		*allowlist = append(*allowlist, matcher)
	}

	return nil
}

func matcherAlreadyInConfig(list []any, matcher packageMatcher) bool {
	return slices.ContainsFunc(list, func(m any) bool {
		m1 := m.(map[string]interface{})
		return (matcher.NameMatcher != nil &&
			matcher.NameMatcher.PackageGlob == m1["packageGlob"]) || (matcher.VersionMatcher != nil &&
			matcher.VersionMatcher.PackageName == m1["package"] &&
			matcher.VersionMatcher.VersionGlob == m1["versionGlob"])
	})
}
