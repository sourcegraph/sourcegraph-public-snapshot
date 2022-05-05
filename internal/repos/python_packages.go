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
	"github.com/sourcegraph/sourcegraph/internal/extsvc/pypi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A PythonPackagesSource creates git repositories from python files of
// published python dependencies.
type PythonPackagesSource struct {
	svc     *types.ExternalService
	config  *schema.PythonPackagesConnection
	depsSvc *dependencies.Service
	client  *pypi.Client
}

// NewPythonPackagesSource returns a new PythonPackagesSource from the given external service.
func NewPythonPackagesSource(svc *types.ExternalService, cf *httpcli.Factory) (*PythonPackagesSource, error) {
	var c schema.PythonPackagesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	return &PythonPackagesSource{
		svc:    svc,
		config: &c,
		/* depsSvc initialized in SetDependenciesService */
		client: pypi.NewClient(svc.URN(), c.Urls, cli),
	}, nil
}

var _ Source = &PythonPackagesSource{}

func (s *PythonPackagesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	deps, err := pythonDependencies(s.config)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	var mu sync.Mutex
	set := make(map[string]struct{})

	for _, dep := range deps {
		if _, ok := set[dep.PackageSyntax()]; ok {
			continue
		}

		_, err := s.client.Version(ctx, dep.PackageSyntax(), dep.PackageVersion())
		if err != nil {
			results <- SourceResult{Source: s, Err: err}
			continue
		}

		repo := s.makeRepo(dep)
		results <- SourceResult{Source: s, Repo: repo}
		set[dep.PackageSyntax()] = struct{}{}
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
			Scheme:      dependencies.PythonPackagesScheme,
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

				_, err := s.client.Version(ctx, depRepo.Name, depRepo.Version)
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
					dep := reposource.NewPythonDependency(depRepo.Name, depRepo.Version)

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

func (s *PythonPackagesSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	dep, err := reposource.ParsePythonDependencyFromRepoName(name)
	if err != nil {
		return nil, err
	}

	_, err = s.client.Project(ctx, dep.PackageSyntax())
	if err != nil {
		return nil, err
	}

	return s.makeRepo(dep), nil
}

func (s *PythonPackagesSource) makeRepo(dep *reposource.PythonDependency) *types.Repo {
	urn := s.svc.URN()
	repoName := dep.RepoName()
	return &types.Repo{
		Name: repoName,
		URI:  string(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceID:   extsvc.TypePythonPackages,
			ServiceType: extsvc.TypePythonPackages,
		},
		Private: false,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: dep.Name,
			},
		},
		Metadata: &struct{}{},
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *PythonPackagesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *PythonPackagesSource) SetDependenciesService(depsSvc *dependencies.Service) {
	s.depsSvc = depsSvc
}

func pythonDependencies(connection *schema.PythonPackagesConnection) (dependencies []*reposource.PythonDependency, err error) {
	for _, dep := range connection.Dependencies {
		dependency, err := reposource.ParsePythonDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}
