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
	"github.com/sourcegraph/sourcegraph/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

var _ Source = &PackagesSource{}

// CheckConnection at this point assumes availability and relies on errors returned
// from the subsequent calls. This is going to be expanded as part of issue #44683
// to actually only return true if the source can serve requests.
func (s *PackagesSource) CheckConnection(ctx context.Context) error {
	return nil
}

func (s *PackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	staticConfigDeps, err := s.configDependencies()
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	handledPackages := make(map[reposource.PackageName]struct{})

	for _, dep := range staticConfigDeps {
		if err := ctx.Err(); err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}

		if _, ok := handledPackages[dep.PackageSyntax()]; !ok {
			_, err := getPackageFromName(s.src, dep.PackageSyntax())
			if err != nil {
				results <- SourceResult{Source: s, Err: err}
				continue
			}
			repo := s.packageToRepoType(dep)
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
		depRepos, _, _, err := s.depsSvc.ListPackageRepoRefs(ctx, dependencies.ListDependencyReposOpts{
			Scheme: s.scheme,
			After:  lastID,
			Limit:  batchLimit,
			// deliberate for clarity
			IncludeBlocked: false,
		})
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			return
		}
		if len(depRepos) == 0 {
			break
		}

		lastID = depRepos[len(depRepos)-1].ID

		for _, depRepo := range depRepos {
			if _, ok := handledPackages[depRepo.Name]; ok {
				continue
			}
			if err := sem.Acquire(ctx, 1); err != nil {
				return
			}
			depRepo := depRepo
			g.Go(func() error {
				defer sem.Release(1)
				pkg, err := getPackageFromName(s.src, depRepo.Name)
				if err != nil {
					if !errcode.IsNotFound(err) {
						results <- SourceResult{Source: s, Err: err}
					}
					return nil
				}

				repo := s.packageToRepoType(pkg)
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

	if allowed, err := s.depsSvc.IsPackageRepoAllowed(ctx, s.scheme, parsedPkg.PackageSyntax()); err != nil {
		return nil, errors.Wrapf(err, "error checking if package repo (%s, %s) is allowed", s.scheme, parsedPkg.PackageSyntax())
	} else if !allowed {
		return nil, &repoupdater.ErrNotFound{
			Repo:       api.RepoName(repoName),
			IsNotFound: true,
		}
	}

	pkg, err := getPackageFromName(s.src, parsedPkg.PackageSyntax())
	if err != nil {
		return nil, err
	}
	return s.packageToRepoType(pkg), nil
}

func (s *PackagesSource) packageToRepoType(dep reposource.Package) *types.Repo {
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
		Metadata: packageMetadata(dep),
	}
}

func getPackageFromName(s packagesSource, name reposource.PackageName) (reposource.Package, error) {
	switch d := s.(type) {
	// Downloading package descriptions is disabled due to performance issues, causing sync times to take >12hr.
	// Don't re-enable the case below without fixing https://github.com/sourcegraph/sourcegraph/issues/39653.
	// case packagesDownloadSource:
	//	return d.GetPackage(ctx, name)
	default:
		return d.ParsePackageFromName(name)
	}
}

func packageMetadata(dep reposource.Package) any {
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

func (s *PackagesSource) configDependencies() (dependencies []reposource.VersionedPackage, err error) {
	for _, dep := range s.configDeps {
		dependency, err := s.src.ParseVersionedPackageFromConfiguration(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}
