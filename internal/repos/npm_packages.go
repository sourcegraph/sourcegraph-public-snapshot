package repos

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npmpackages"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/npmpackages/npm"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A NPMPackagesSource creates git repositories from `*-sources.tar.gz` files of
// published NPM dependencies from the JS ecosystem.
type NPMPackagesSource struct {
	svc     *types.ExternalService
	config  *schema.NPMPackagesConnection
	dbStore NPMPackagesRepoStore
}

// NewNPMPackagesSource returns a new NPMSource from the given external
// service.
func NewNPMPackagesSource(svc *types.ExternalService) (*NPMPackagesSource, error) {
	var c schema.NPMPackagesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return &NPMPackagesSource{svc: svc, config: &c /*dbStore initialized in SetDB */}, nil
}

var _ Source = &NPMPackagesSource{}

// ListRepos returns all NPM artifacts accessible to all connections
// configured in Sourcegraph via the external services configuration.
//
// [FIXME: deduplicate-listed-repos] The current implementation will return
// multiple repos with the same URL if there are different versions of it.
func (s *NPMPackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	npmPackages, err := npmPackages(*s.config)
	if err != nil {
		results <- SourceResult{Err: err}
		return
	}
	for _, npmPackage := range npmPackages {
		repo := s.makeRepo(npmPackage)
		results <- SourceResult{
			Source: s,
			Repo:   repo,
		}
	}
	if err != nil {
		results <- SourceResult{Err: err}
		return
	}
	totalDBFetched, totalDBResolved, lastID := 0, 0, 0
	for {
		dbDeps, err := s.dbStore.GetNPMDependencyRepos(ctx, dbstore.GetNPMDependencyReposOpts{
			After: lastID,
			Limit: 100,
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
			parsedDbPackage, err := reposource.ParseNPMPackageFromPackageSyntax(dbDep.Package)
			if err != nil {
				log15.Error("failed to parse npm package name retrieved from database", "package", dbDep.Package)
				continue
			}
			npmDependency := reposource.NPMDependency{NPMPackage: *parsedDbPackage, Version: dbDep.Version}
			if err = npm.Exists(ctx, s.config, npmDependency); err != nil {
				log15.Warn("failed to resolve npm dependency", "package", npmDependency.PackageManagerSyntax(), "message", err)
				continue
			}
			repo := s.makeRepo(npmDependency.NPMPackage)
			totalDBResolved++
			results <- SourceResult{Source: s, Repo: repo}
		}
	}
	log15.Info("finish resolving npm artifacts", "totalDB", totalDBFetched, "totalDBResolved", totalDBResolved, "totalConfig", len(npmPackages))
}

func (s *NPMPackagesSource) makeRepo(npmPackage reposource.NPMPackage) *types.Repo {
	urn := s.svc.URN()
	cloneURL := npmPackage.CloneURL()
	repoName := npmPackage.RepoName()
	return &types.Repo{
		Name: repoName,
		URI:  string(repoName),
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
	once.Do(func() {
		observationContext = &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}
		operationMetrics = dbstore.NewREDMetrics(observationContext)
	})
	s.dbStore = dbstore.NewWithDB(db, observationContext, operationMetrics)
}

// npmPackages gets the list of applicable packages by de-duplicating dependencies
// present in the configuration.
func npmPackages(connection schema.NPMPackagesConnection) ([]reposource.NPMPackage, error) {
	dependencies, err := npmDependencies(connection)
	if err != nil {
		return nil, err
	}
	npmPackages := []reposource.NPMPackage{}
	isAdded := make(map[reposource.NPMPackage]bool)
	for _, dep := range dependencies {
		npmPackage := dep.NPMPackage
		if !isAdded[npmPackage] {
			npmPackages = append(npmPackages, npmPackage)
		}
		isAdded[npmPackage] = true
	}
	return npmPackages, nil
}

func npmDependencies(connection schema.NPMPackagesConnection) (dependencies []reposource.NPMDependency, err error) {
	for _, dep := range connection.Dependencies {
		dependency, err := reposource.ParseNPMDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, *dependency)
	}
	return dependencies, nil
}

type NPMPackagesRepoStore interface {
	GetNPMDependencyRepos(ctx context.Context, filter dbstore.GetNPMDependencyReposOpts) ([]dbstore.NPMDependencyRepo, error)
}
