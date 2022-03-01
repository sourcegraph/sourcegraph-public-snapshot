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

// A NPMPackagesSource creates git repositories from `*-sources.tar.gz` files of
// published NPM dependencies from the JS ecosystem.
type NPMPackagesSource struct {
	svc        *types.ExternalService
	connection schema.NPMPackagesConnection
	depsStore  DependenciesStore
	client     npm.Client
}

// NewNPMPackagesSource returns a new NPMSource from the given external
// service.
func NewNPMPackagesSource(svc *types.ExternalService) (*NPMPackagesSource, error) {
	var c schema.NPMPackagesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return &NPMPackagesSource{
		svc:        svc,
		connection: c,
		/*dbStore initialized in SetDB */
		client: npm.NewHTTPClient(c.Registry, c.RateLimit, c.Credentials),
	}, nil
}

var _ Source = &NPMPackagesSource{}

// ListRepos returns all NPM artifacts accessible to all connections
// configured in Sourcegraph via the external services configuration.
//
// [FIXME: deduplicate-listed-repos] The current implementation will return
// multiple repos with the same URL if there are different versions of it.
func (s *NPMPackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
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
			Scheme: dependenciesStore.NPMPackagesScheme,
			After:  lastID,
			Limit:  100,
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
			parsedDbPackage, err := reposource.ParseNPMPackageFromPackageSyntax(dbDep.Name)
			if err != nil {
				log15.Error("failed to parse npm package name retrieved from database", "package", dbDep.Name, "error", err)
				continue
			}

			npmDependency := reposource.NPMDependency{NPMPackage: parsedDbPackage, Version: dbDep.Version}
			pkgKey := npmDependency.PackageSyntax()
			info := pkgVersions[pkgKey]
			if info == nil {
				info, err = s.client.GetPackageInfo(ctx, npmDependency.NPMPackage)
				if err != nil {
					pkgVersions[pkgKey] = &npm.PackageInfo{Versions: map[string]*npm.DependencyInfo{}}
					log15.Warn("npm package not found in registry", "package", pkgKey, "err", err)
					continue
				}
				pkgVersions[pkgKey] = info
			}
			if _, hasVersion := info.Versions[npmDependency.Version]; !hasVersion {
				if len(info.Versions) != 0 { // We must've already logged a package not found earlier if len is 0.
					log15.Warn("npm dependency does not exist",
						"dependency", npmDependency.PackageManagerSyntax())
				}
				continue
			}
			repo := s.makeRepo(npmDependency.NPMPackage, info.Description)
			totalDBResolved++
			results <- SourceResult{Source: s, Repo: repo}
		}
	}
	log15.Info("finish resolving npm artifacts", "totalDB", totalDBFetched, "totalDBResolved", totalDBResolved, "totalConfig", len(npmPackages))
}

func (s *NPMPackagesSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	pkg, err := reposource.ParseNPMPackageFromRepoURL(name)
	if err != nil {
		return nil, err
	}

	info, err := s.client.GetPackageInfo(ctx, pkg)
	if err != nil {
		return nil, err
	}

	return s.makeRepo(pkg, info.Description), nil
}

func (s *NPMPackagesSource) makeRepo(npmPackage *reposource.NPMPackage, description string) *types.Repo {
	urn := s.svc.URN()
	cloneURL := npmPackage.CloneURL()
	repoName := npmPackage.RepoName()
	return &types.Repo{
		Name:        repoName,
		Description: description,
		URI:         string(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceID:   extsvc.TypeNPMPackages,
			ServiceType: extsvc.TypeNPMPackages,
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
func (s *NPMPackagesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *NPMPackagesSource) SetDB(db dbutil.DB) {
	s.depsStore = dependenciesStore.GetStore(database.NewDB(db))
}

// npmPackages gets the list of applicable packages by de-duplicating dependencies
// present in the configuration.
func npmPackages(connection schema.NPMPackagesConnection) ([]*reposource.NPMPackage, error) {
	dependencies, err := npmDependencies(connection)
	if err != nil {
		return nil, err
	}
	npmPackages := []*reposource.NPMPackage{}
	isAdded := make(map[string]bool)
	for _, dep := range dependencies {
		if key := dep.PackageSyntax(); !isAdded[key] {
			npmPackages = append(npmPackages, dep.NPMPackage)
			isAdded[key] = true
		}
	}
	return npmPackages, nil
}

func npmDependencies(connection schema.NPMPackagesConnection) (dependencies []*reposource.NPMDependency, err error) {
	for _, dep := range connection.Dependencies {
		dependency, err := reposource.ParseNPMDependency(dep)
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
