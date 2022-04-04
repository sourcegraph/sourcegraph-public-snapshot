package repos

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/sourcegraph/sourcegraph/internal/api"
	dependenciesStore "github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies/store"
	"github.com/sourcegraph/sourcegraph/internal/conf/reposource"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/extsvc/gomodproxy"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

// A GoModulesSource creates git repositories from go module zip files of
// published go dependencies from the Go ecosystem.
type GoModulesSource struct {
	svc       *types.ExternalService
	config    *schema.GoModulesConnection
	depsStore DependenciesStore
	client    *gomodproxy.Client
}

// NewGoModulesSource returns a new GoModulesSource from the given external service.
func NewGoModulesSource(svc *types.ExternalService, cf *httpcli.Factory) (*GoModulesSource, error) {
	var c schema.GoModulesConnection
	if err := jsonc.Unmarshal(svc.Config, &c); err != nil {
		return nil, errors.Errorf("external service id=%d config error: %s", svc.ID, err)
	}

	cli, err := cf.Doer()
	if err != nil {
		return nil, err
	}

	return &GoModulesSource{
		svc:    svc,
		config: &c,
		/*dbStore initialized in SetDB */
		client: gomodproxy.NewClient(&c, cli),
	}, nil
}

var _ Source = &GoModulesSource{}

func (s *GoModulesSource) ListRepos(ctx context.Context, results chan SourceResult) {
	deps, err := goDependencies(s.config)
	if err != nil {
		results <- SourceResult{Source: s, Err: err}
		return
	}

	var mu sync.Mutex
	set := make(map[string]struct{})

	for _, dep := range deps {
		_, err := s.client.GetVersion(ctx, dep.PackageSyntax(), dep.PackageVersion())
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
		depRepos, err := s.depsStore.ListDependencyRepos(ctx, dependenciesStore.ListDependencyReposOpts{
			Scheme:      dependenciesStore.GoModulesScheme,
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

				mod, err := s.client.GetVersion(ctx, depRepo.Name, depRepo.Version)
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
					dep := reposource.NewGoDependency(*mod)
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

func (s *GoModulesSource) GetRepo(ctx context.Context, name string) (*types.Repo, error) {
	dep, err := reposource.ParseGoDependencyFromRepoName(name)
	if err != nil {
		return nil, err
	}

	_, err = s.client.GetVersion(ctx, dep.PackageSyntax(), "")
	if err != nil {
		return nil, err
	}

	return s.makeRepo(dep), nil
}

func (s *GoModulesSource) makeRepo(dep *reposource.GoDependency) *types.Repo {
	urn := s.svc.URN()
	repoName := dep.RepoName()
	return &types.Repo{
		Name: repoName,
		URI:  string(repoName),
		ExternalRepo: api.ExternalRepoSpec{
			ID:          string(repoName),
			ServiceID:   extsvc.TypeGoModules,
			ServiceType: extsvc.TypeGoModules,
		},
		Private: false,
		Sources: map[string]*types.SourceInfo{
			urn: {
				ID:       urn,
				CloneURL: string(repoName),
			},
		},
		Metadata: struct{}{},
	}
}

// ExternalServices returns a singleton slice containing the external service.
func (s *GoModulesSource) ExternalServices() types.ExternalServices {
	return types.ExternalServices{s.svc}
}

func (s *GoModulesSource) SetDB(db dbutil.DB) {
	s.depsStore = dependenciesStore.GetStore(database.NewDB(db))
}

func goDependencies(connection *schema.GoModulesConnection) (dependencies []*reposource.GoDependency, err error) {
	for _, dep := range connection.Dependencies {
		dependency, err := reposource.ParseGoDependency(dep)
		if err != nil {
			return nil, err
		}
		dependencies = append(dependencies, dependency)
	}
	return dependencies, nil
}
