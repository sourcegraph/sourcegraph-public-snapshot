package repos

import (
	"context"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	dependenciesStore "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npm/npmpackages"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A NpmPackagesSource creates git repositories from `*-sources.tar.gz` files of
// published npm dependencies from the JS ecosystem.
type NpmPackagesSource struct {
	svc        *types.ExternalService
	connection schema.NpmPackagesConnection
	depsStore  DependenciesStore
	client     npm.Client
}

// NewNpmPackagesSource returns a new NpmSource from the given external
// service.
func NewNpmPackagesSource(svc *types.ExternalService) (*NpmPackagesSource, error) {
	var c schema.NpmPackagesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return &NpmPackagesSource{
		svc:        svc,
		connection: c,
		/*dbStore initialized in SetDB */
		client: npm.NewHTTPClient(c.Registry, c.RateLimit, c.Credentials),
	}, nil
}

var _ Source = &NpmPackagesSource{}

// ListRepos returns all npm artifacts accessible to all connections
// configured in Sourcegraph via the external services configuration.
//
// [FIXME: deduplicate-listed-repos] The current implementation will return
// multiple repos with the same URL if there are different versions of it.
func (s *NpmPackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	npmPackages, err := npmPackages(s.connection)
	if err != nil {
		results <- SourceResult{Err: err}
		return
	}

	for _, npmPackage := range npmPackages {
		info, err := s.client.GetPackageInfo(ctx, npmPackage)
		if err != nil {
			results <- SourceResult{Err: err}
			continue
		}

		repo := s.makeRepo(npmPackage, info.Description)
		results <- SourceResult{
			Source: s,
			Repo:   repo,
		}
	}

	totalDBFetched, totalDBResolved, lastID := 0, 0, 0
	pkgVersions := map[string]*npm.PackageInfo{}
	for {
		dbDeps, err := s.depsStore.ListDependencyRepos(ctx, dependenciesStore.ListDependencyReposOpts{
			Scheme:      dependenciesStore.NpmPackagesScheme,
			After:       lastID,
			Limit:       100,
			NewestFirst: true,
		})
		if err != nil {
			results <- SourceResult{Err: err}
			return
		}
		if len(dbDeps) == 0 {
			break
		}
		totalDBFetched += len(dbDeps)
		lastID = dbDeps[len(dbDeps)-1].ID
		for _, dbDep := range dbDeps {
			parsedDbPackage, err := reposource.ParseNpmPackageFromPackageSyntax(dbDep.Name)
			if err != nil {
				log15.Error("failed to parse npm package name retrieved from database", "package", dbDep.Name, "error", err)
				continue
			}

			npmDependency := reposource.NpmDependency{NpmPackage: parsedDbPackage, Version: dbDep.Version}
			pkgKey := npmDependency.PackageSyntax()
			info := pkgVersions[pkgKey]

			if info == nil {
				info, err = s.client.GetPackageInfo(ctx, npmDependency.NpmPackage)
				if err != nil {
					pkgVersions[pkgKey] = &npm.PackageInfo{Versions: map[string]*npm.DependencyInfo{}}
					continue
				}

				pkgVersions[pkgKey] = info
			}

			if _, hasVersion := info.Versions[npmDependency.Version]; !hasVersion {
				continue
			}

			repo := s.makeRepo(npmDependency.NpmPackage, info.Description)
			totalDBResolved++
			results <- SourceResult{Source: s, Repo: repo}
		}
	}
	log15.Info("finish resolving npm artifacts", "totalDB", totalDBFetched, "totalDBResolved", totalDBResolved, "totalConfig", len(npmPackages))
}

func (s *NpmPackagesSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	pkg, err := reposource.ParseNpmPackageFromRepoURL(name)
	if err != nil {
		return nil, err
	}

	info, err := s.client.GetPackageInfo(ctx, pkg)
	if err != nil {
		return nil, err
	}

	return s.makeRepo(pkg, info.Description), nil
}

func (s *NpmPackagesSource) makeRepo(npmPackage *reposource.NpmPackage, description string) *types.Repo {
	urn := s.svc.URN()
	cloneURL := npmPackage.CloneURL()
	repoName := npmPackage.RepoName()
	return &types.Repo{
		Name:        repoName,
		Description: description,
		URI:         string(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceID:   extsvc.TypeNpmPackages,
			ServiceType: extsvc.TypeNpmPackages,
		},
		Private: false,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metadata: &npmpackages.Metadata{
			Package: npmPackage,
		},
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *NpmPackagesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *NpmPackagesSource) SetDB(db dbutil.DB) {
	s.depsStore = dependenciesStore.GetStore(database.NewDB(db))
}

// npmPackages gets the list of applicable packages by de-duplicating dependencies
// present in the configuration.
func npmPackages(connection schema.NpmPackagesConnection) ([]*reposource.NpmPackage, error) {
	dependencies, err := npmDependencies(connection)
	if err != nil {
		return nil, err
	}
	npmPackages := []*reposource.NpmPackage{}
	isAdded := make(map[string]bool)
	for _, dep := range dependencies {
		if key := dep.PackageSyntax(); !isAdded[key] {
			npmPackages = append(npmPackages, dep.NpmPackage)
			isAdded[key] = true
		}
	}
	return npmPackages, nil
}

func npmDependencies(connection schema.NpmPackagesConnection) (dependencies []*reposource.NpmDependency, err error) {
	for _, dep := range connection.Dependencies {
		dependency, err := reposource.ParseNpmDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}

type DependenciesStore interface {
	ListDependencyRepos(ctx context.Context, opts dependenciesStore.ListDependencyReposOpts) ([]dependenciesStore.DependencyRepo, error)
}
