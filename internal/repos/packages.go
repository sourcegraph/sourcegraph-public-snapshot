package repos

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// A PackagesSource yields dependency repositories from a package (dependencies) host connection.
type PackagesSource struct {
	svc        *types.ExternalService
	configDeps []string
	scheme     string
	depsSvc    *dependencies.Service
	src        packagesSource
}

type packagesSource interface {
	// ParseVersionedPackageFromConfiguration parses a package and version from the "dependencies"
	// field from the site-admin interface.
	// For example: "react@1.2.0" or "com.google.guava:guava:30.0-jre".
	ParseVersionedPackageFromConfiguration(dep string) (reposource.VersionedPackage, error)
	// ParsePackageFromRepoName parses a Sourcegraph repository name of the package.
	// For example: "npm/react" or "maven/com.google.guava/guava".
	ParsePackageFromRepoName(repoName api.RepoName) (reposource.Package, error)
	// ParsePackageFromName parses a package from the name of the package, as accepted by the ecosystem's package manager.
	// For example: "react" or "com.google.guava:guava".
	ParsePackageFromName(name reposource.PackageName) (reposource.Package, error)
	// functions in this file that switch against concrete implementations of this interface:
	// getPackage(): to fetch the description of this package, only supported by a few implementations.
	// metadata(): to store gob-encoded structs with implementation-specific metadata.
}

type packagesDownloadSource interface {
	// GetPackage sends a request to the package host to get metadata about this package, like the description.
	GetPackage(ctx context.Context, name reposource.PackageName) (reposource.Package, error)
}

var _ Source = &PackagesSource{}

// CheckConnection at this point assumes availability and relies on errors returned
// from the subsequent calls. This is going to be expanded as part of issue #44683
// to actually only return true if the source can serve requests.
func (s *PackagesSource) CheckConnection(ctx context.Context) error {
	return nil
}

func (s *PackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	deps, err := s.configDependencies(s.configDeps)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	handledPackages := make(map[reposource.PackageName]struct{})

	for _, dep := range deps {
		if err := ctx.Err(); err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		if _, ok := handledPackages[dep.PackageSyntax()]; !ok {
			_, err := getPackage(s.src, dep.PackageSyntax())
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				continue
			}
			repo := s.makeRepo(dep)
			results <- SourceResult{Source: s, Repo: repo}
			handledPackages[dep.PackageSyntax()] = struct{}{}
		}
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := semaphore.NewWeighted(32)
	g, ctx := errgroup.WithContext(ctx)

	defer func() {
		if err := g.Wait(); err != nil && err != context.Canceled {
			results <- SourceResult{Source: s, Err: err}
		}
	}()

	const batchLimit = 100
	var lastID int
	for {
		depRepos, _, err := s.depsSvc.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme: s.scheme,
			After:  lastID,
			Limit:  batchLimit,
		})
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}
		if len(depRepos) == 0 {
			break
		}

		lastID = depRepos[len(depRepos)-1].ID

		// at most batchLimit because of the limit above
		depReposToHandle := make([]dependencies.PackageRepoReference, 0, len(depRepos))
		for _, depRepo := range depRepos {
			if _, ok := handledPackages[depRepo.Name]; !ok {
				// don't need to add to handledPackages here, as the results from
				// depRepos should be unique
				depReposToHandle = append(depReposToHandle, depRepo)
			}
		}

		for _, depRepo := range depReposToHandle {
			if err := sem.Acquire(ctx, 1); err != nil {
				return
			}
			depRepo := depRepo
			g.Go(func() error {
				defer sem.Release(1)
				pkg, err := getPackage(s.src, depRepo.Name)
				if err != nil {
					if !errcode.IsNotFound(err) {
						results <- SourceResult{Source: s, Err: err}
					}
					return nil
				}

				repo := s.makeRepo(pkg)
				results <- SourceResult{Source: s, Repo: repo}

				return nil
			})
		}
	}
}

func (s *PackagesSource) GetRepo(ctx context.Context, repoName string) (*types.Repo, error) {
	parsedPkg, err := s.src.ParsePackageFromRepoName(api.RepoName(repoName))
	if err != nil {
		return nil, err
	}

	pkg, err := getPackage(s.src, parsedPkg.PackageSyntax())
	if err != nil {
		return nil, err
	}
	return s.makeRepo(pkg), nil
}

func (s *PackagesSource) makeRepo(dep reposource.Package) *types.Repo {
	urn := s.svc.URN()
	repoName := dep.RepoName()
	return &types.Repo{
		Name:        repoName,
		Description: dep.Description(),
		URI:         string(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceID:   extsvc.KindToType(s.svc.Kind),
			ServiceType: extsvc.KindToType(s.svc.Kind),
		},
		Private: false,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: string(repoName),
			},
		},
		Metadata: metadata(dep),
	}
}

func getPackage(s packagesSource, name reposource.PackageName) (reposource.Package, error) {
	switch d := s.(type) {
	// Downloading package descriptions is disabled due to performance issues, causing sync times to take >12hr.
	// Don't re-enable the case below without fixing https://github.com/sourcegraph/sourcegraph/issues/39653.
	// case packagesDownloadSource:
	//	return d.GetPackage(ctx, name)
	default:
		return d.ParsePackageFromName(name)
	}
}

func metadata(dep reposource.Package) any {
	switch d := dep.(type) {
	case *reposource.MavenVersionedPackage:
		return &reposource.MavenMetadata{
			Module: d.MavenModule,
		}
	case *reposource.NpmVersionedPackage:
		return &reposource.NpmMetadata{
			Package: d.NpmPackageName,
		}
	default:
		return &struct{}{}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *PackagesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *PackagesSource) SetDependenciesService(depsSvc *dependencies.Service) {
	s.depsSvc = depsSvc
}

func (s *PackagesSource) configDependencies(deps []string) (dependencies []reposource.VersionedPackage, err error) {
	for _, dep := range deps {
		dependency, err := s.src.ParseVersionedPackageFromConfiguration(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}
