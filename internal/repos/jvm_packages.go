package repos

import (
	"context"
	"fmt"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/metrics"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

var (
	observationContext *observation.Context
	operationMetrics   *metrics.OperationMetrics
	once               sync.Once
)

// A JVMPackagesSource creates git repositories from `*-sources.jar` files of
// published Maven dependencies from the JVM ecosystem.
type JVMPackagesSource struct {
	svc     *types.ExternalService
	config  *schema.JVMPackagesConnection
	dbStore JVMPackagesRepoStore
}

type JVMPackagesRepoStore interface {
	GetJVMDependencyRepos(ctx context.Context, filter dbstore.GetJVMDependencyReposOpts) ([]dbstore.JVMDependencyRepo, error)
}

// NewJVMPackagesSource returns a new MavenSource from the given external
// service.
func NewJVMPackagesSource(svc *types.ExternalService) (*JVMPackagesSource, error) {
	var c schema.JVMPackagesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, fmt.Errorf("external service id=%d config error: %s", svc.ID, err)
	}
	return newJVMPackagesSource(svc, &c)
}

func (s *JVMPackagesSource) SetDB(db dbutil.DB) {
	once.Do(func() {
		observationContext = &observation.Context{
			Logger:     log15.Root(),
			Tracer:     &trace.Tracer{Tracer: opentracing.GlobalTracer()},
			Registerer: prometheus.DefaultRegisterer,
		}
		operationMetrics = dbstore.NewOperationsMetrics(observationContext)
	})
	s.dbStore = dbstore.NewWithDB(db, observationContext, operationMetrics)
}

func newJVMPackagesSource(svc *types.ExternalService, c *schema.JVMPackagesConnection) (*JVMPackagesSource, error) {
	return &JVMPackagesSource{
		svc:     svc,
		config:  c,
		dbStore: nil, // set via SetDB decorator
	}, nil
}

// ListRepos returns all Maven artifacts accessible to all connections
// configured in Sourcegraph via the external services configuration.
func (s *JVMPackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	s.listDependentRepos(ctx, results)
}

func (s *JVMPackagesSource) listDependentRepos(ctx context.Context, results chan SourceResult) {
	modules, err := MavenModules(*s.config)
	if err != nil {
		results <- SourceResult{Err: err}
		return
	}
	for _, module := range modules {
		repo := s.makeRepo(module)
		results <- SourceResult{
			Source: s,
			Repo:   repo,
		}
	}
	if err != nil {
		results <- SourceResult{Err: err}
		return
	}

	lastID := 0
	for {
		dbDeps, err := s.dbStore.GetJVMDependencyRepos(ctx, dbstore.GetJVMDependencyReposOpts{
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

		lastID = dbDeps[len(dbDeps)-1].ID

		for _, dep := range dbDeps {
			parsedModule, err := reposource.ParseMavenModule(dep.Module)
			if err != nil {
				log15.Warn("error parsing maven module", "error", err, "module", dep.Module)
				continue
			}
			mavenDependency := reposource.MavenDependency{MavenModule: parsedModule, Version: dep.Version}

			// We dont return anything that isnt resolvable here, to reduce logspam from gitserver. This codepath
			// should be hit much less frequently than gitservers attempts to get packages, so there should be less
			// logspam. This may no longer hold true if the extsvc syncs more often than gitserver would, but I
			// don't foresee that happening (not soon at least).
			if !coursier.Exists(ctx, s.config, mavenDependency) {
				log15.Warn("jvm package not resolvable from coursier", "package", mavenDependency.CoursierSyntax())
				continue
			}

			repo := s.makeRepo(mavenDependency.MavenModule)
			results <- SourceResult{
				Source: s,
				Repo:   repo,
			}
		}
	}
}

func (s *JVMPackagesSource) GetRepo(ctx context.Context, artifactPath string) (*types.Repo, error) {
	module, err := reposource.ParseMavenModule(artifactPath)
	if err != nil {
		return nil, err
	}

	dependencies, err := MavenDependencies(*s.config)
	if err != nil {
		return nil, err
	}

	dbDeps, err := s.dbStore.GetJVMDependencyRepos(ctx, dbstore.GetJVMDependencyReposOpts{
		ArtifactName: artifactPath,
	})
	if err != nil {
		return nil, errors.Wrap(err, "dbstore.GetJVMDependencyRepos")
	}

	for _, dep := range dbDeps {
		parsedModule, err := reposource.ParseMavenModule(dep.Module)
		if err != nil {
			log15.Warn("error parsing maven module", "error", err, "module", dep.Module)
			continue
		}
		dependency := reposource.MavenDependency{
			MavenModule: parsedModule,
			Version:     dep.Version,
		}
		dependencies = append(dependencies, dependency)
	}

	nonExistentDependencies := make([]reposource.MavenDependency, 0)
	hasAtLeastOneValidDependency := false
	for _, dep := range dependencies {
		if dep.MavenModule == module {
			if coursier.Exists(ctx, s.config, dep) {
				hasAtLeastOneValidDependency = true
			} else {
				nonExistentDependencies = append(nonExistentDependencies, dep)
			}
		}
	}

	if !hasAtLeastOneValidDependency {
		return nil, &jvmDependencyNotFound{
			dependencies: nonExistentDependencies,
		}
	}

	for _, nonExistentDependency := range nonExistentDependencies {
		// Don't reject all versions if a single version fails to
		// resolve. Instead, we just log a warning about the unresolved
		// dependency. A dependency can fail to resolve if it gets
		// removed from the package host for some reason.
		log15.Warn("Skipping non-existing JVM package", "nonExistentDependency", nonExistentDependency.CoursierSyntax())
	}

	return s.makeRepo(module), nil
}

type jvmDependencyNotFound struct {
	dependencies []reposource.MavenDependency
}

func (e *jvmDependencyNotFound) Error() string {
	return fmt.Sprintf("not found: jvm dependency '%v'", e.dependencies)
}

func (s *JVMPackagesSource) makeRepo(module reposource.MavenModule) *types.Repo {
	urn := s.svc.URN()
	cloneURL := module.CloneURL()
	return &types.Repo{
		Name: module.RepoName(),
		URI:  string(module.RepoName()),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(module.RepoName()),
			ServiceID:   extsvc.TypeJVMPackages,
			ServiceType: extsvc.TypeJVMPackages,
		},
		Private: false,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: cloneURL,
			},
		},
		Metadata: &jvmpackages.Metadata{
			Module: module,
		},
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *JVMPackagesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func MavenDependencies(connection schema.JVMPackagesConnection) (dependencies []reposource.MavenDependency, err error) {
	for _, dep := range connection.Maven.Dependencies {
		dependency, err := reposource.ParseMavenDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}

func MavenModules(connection schema.JVMPackagesConnection) ([]reposource.MavenModule, error) {
	isAdded := make(map[reposource.MavenModule]bool)
	modules := []reposource.MavenModule{}
	dependencies, err := MavenDependencies(connection)
	if err != nil {
		return nil, err
	}
	for _, dep := range dependencies {
		module := dep.MavenModule
		if _, added := isAdded[module]; !added {
			modules = append(modules, module)
		}
		isAdded[module] = true
	}
	return modules, nil
}
