package repos

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// A DependenciesSource yields dependency repositories from a package (dependencies) host connection.
type DependenciesSource struct {
	svc        *types.ExternalService
	configDeps []string
	scheme     string
	depsSvc    *dependencies.Service
	src        dependenciesSource
}

type dependenciesSource interface {
	Get(ctx context.Context, name, version string) (reposource.PackageDependency, error)
	ParseDependency(dep string) (reposource.PackageDependency, error)
	ParseDependencyFromRepoName(repoName string) (reposource.PackageDependency, error)
}

var _ Source = &DependenciesSource{}

func (s *DependenciesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	deps, err := s.configDependencies(s.configDeps)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	var mu sync.Mutex
	set := make(map[string]struct{})

	for _, dep := range deps {
		_, err := s.src.Get(ctx, dep.PackageSyntax(), dep.PackageVersion())
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			continue
		}

		if _, ok := set[dep.PackageSyntax()]; !ok {
			repo := s.makeRepo(dep)
			results <- SourceResult{Source: s, Repo: repo}
			set[dep.PackageSyntax()] = struct{}{}
		}
	}

	lastID := 0

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sem := semaphore.NewWeighted(32)
	g, ctx := errgroup.WithContext(ctx)

	defer func() {
		if err := g.Wait(); err != nil && err != context.Canceled {
			results <- SourceResult{Source: s, Err: err}
		}
	}()

	for {
		depRepos, err := s.depsSvc.ListDependencyRepos(ctx, dependencies.ListDependencyReposOpts{
			Scheme:      s.scheme,
			After:       lastID,
			Limit:       100,
			NewestFirst: true,
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
			if err := sem.Acquire(ctx, 1); err != nil {
				return
			}

			depRepo := depRepo
			g.Go(func() error {
				defer sem.Release(1)

				dep, err := s.src.Get(ctx, depRepo.Name, depRepo.Version)
				if err != nil {
					if errcode.IsNotFound(err) {
						return nil
					}
					return err
				}

				mu.Lock()
				if _, ok := set[depRepo.Name]; !ok {
					set[depRepo.Name] = struct{}{}
					mu.Unlock()
					repo := s.makeRepo(dep)
					results <- SourceResult{Source: s, Repo: repo}
				} else {
					mu.Unlock()
				}

				return nil
			})
		}
	}
}

func (s *DependenciesSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	dep, err := s.src.ParseDependencyFromRepoName(name)
	if err != nil {
		return nil, err
	}

	dep, err = s.src.Get(ctx, dep.PackageSyntax(), "")
	if err != nil {
		return nil, err
	}

	return s.makeRepo(dep), nil
}

func (s *DependenciesSource) makeRepo(dep reposource.PackageDependency) *types.Repo {
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

func metadata(dep reposource.PackageDependency) any {
	switch d := dep.(type) {
	case *reposource.MavenDependency:
		return &reposource.MavenMetadata{
			Module: d.MavenModule,
		}
	case *reposource.NpmDependency:
		return &reposource.NpmMetadata{
			Package: d.NpmPackage,
		}
	default:
		return &struct{}{}
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *DependenciesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *DependenciesSource) SetDependenciesService(depsSvc *dependencies.Service) {
	s.depsSvc = depsSvc
}

func (s *DependenciesSource) configDependencies(deps []string) (dependencies []reposource.PackageDependency, err error) {
	for _, dep := range deps {
		dependency, err := s.src.ParseDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}
