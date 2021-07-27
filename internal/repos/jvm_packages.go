package repos

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/jvmpackages/coursier"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A JVMPackagesSource creates git repositories from `*-sources.jar` files of
// published Maven dependencies from the JVM ecosystem.
type JVMPackagesSource struct {
	svc    *types.ExternalService
	config *schema.JVMPackagesConnection
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

func newJVMPackagesSource(svc *types.ExternalService, c *schema.JVMPackagesConnection) (*JVMPackagesSource, error) {
	return &JVMPackagesSource{
		svc:    svc,
		config: c,
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
