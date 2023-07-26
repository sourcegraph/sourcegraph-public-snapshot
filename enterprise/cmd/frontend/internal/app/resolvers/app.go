package resolvers

import (
	"context"
	"path/filepath"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/jsonc"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/schema"
)

type appResolver struct {
	logger    log.Logger
	db        database.DB
	gitClient gitserver.Client
}

var _ graphqlbackend.AppResolver = &appResolver{}

func NewAppResolver(logger log.Logger, db database.DB, gitClient gitserver.Client) *appResolver {
	return &appResolver{
		logger:    logger,
		db:        db,
		gitClient: gitClient,
	}
}

func (r *appResolver) checkLocalDirectoryAccess(ctx context.Context) error {
	return auth.CheckCurrentUserIsSiteAdmin(ctx, r.db)
}

func (r *appResolver) LocalDirectories(ctx context.Context, args *graphqlbackend.LocalDirectoryArgs) (graphqlbackend.LocalDirectoryResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := r.checkLocalDirectoryAccess(ctx); err != nil {
		return nil, err
	}

	// Make sure all paths are absolute
	absPaths := make([]string, 0, len(args.Paths))
	for _, path := range args.Paths {
		if path == "" {
			return nil, errors.New("Path must be non-empty string")
		}

		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, err
		}
		absPaths = append(absPaths, absPath)
	}

	return &localDirectoryResolver{paths: absPaths}, nil
}

func (r *appResolver) SetupNewAppRepositoriesForEmbedding(ctx context.Context, args graphqlbackend.SetupNewAppRepositoriesForEmbeddingArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	repoEmbeddingsStore := repo.NewRepoEmbeddingJobsStore(r.db)
	jobContext, cancel := context.WithDeadline(ctx, time.Now().Add(20*time.Second))
	defer cancel()
	p := pool.New().WithMaxGoroutines(10).WithContext(jobContext)
	for _, repo := range args.RepoNames {
		repoName := api.RepoName(repo)
		p.Go(func(ctx context.Context) error {
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-jobContext.Done():
					return errors.New("time limit exceeded unable to schedule repo")
				case <-ticker.C:
					r.logger.Debug("Checking repo")
					branch, _, err := r.gitClient.GetDefaultBranch(ctx, repoName, true)
					if err == nil && branch != "" {
						if err := embeddings.ScheduleRepositoriesForEmbedding(
							ctx,
							[]api.RepoName{repoName},
							false,
							r.db,
							repoEmbeddingsStore,
							r.gitClient,
						); err == nil {
							r.logger.Debug("Repo scheduled")
							return nil
						}
					}
					r.logger.Debug("Repo not cloned")
				}
			}
		})
	}
	err := p.Wait()
	if err != nil {
		r.logger.Warn("error scheduling repos for embedding", log.Error(err))
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

type localDirectoryResolver struct {
	paths []string
}

func (r *localDirectoryResolver) Paths() []string {
	return r.paths
}

func (r *localDirectoryResolver) Repositories(ctx context.Context) ([]graphqlbackend.LocalRepositoryResolver, error) {
	var allRepos []graphqlbackend.LocalRepositoryResolver

	for _, path := range r.paths {
		repos, err := servegit.Service.Repos(ctx, path)
		if err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}

		for _, repo := range repos {
			allRepos = append(allRepos, localRepositoryResolver{
				name: repo.Name,
				path: repo.AbsFilePath,
			})
		}
	}

	return allRepos, nil
}

type localRepositoryResolver struct {
	name string
	path string
}

func (r localRepositoryResolver) Name() string {
	return r.name
}

func (r localRepositoryResolver) Path() string {
	return r.path
}

func (r *appResolver) LocalExternalServices(ctx context.Context) ([]graphqlbackend.LocalExternalServiceResolver, error) {
	// 🚨 SECURITY: Only site admins on app may use API which accesses local filesystem.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	externalServices, err := backend.NewAppExternalServices(r.db).LocalExternalServices(ctx)
	if err != nil {
		return nil, err
	}

	localExternalServices := make([]graphqlbackend.LocalExternalServiceResolver, 0)
	for _, externalService := range externalServices {
		serviceConfig, err := externalService.Config.Decrypt(ctx)
		if err != nil {
			return nil, err
		}

		var otherConfig schema.OtherExternalServiceConnection
		if err = jsonc.Unmarshal(serviceConfig, &otherConfig); err != nil {
			return nil, errors.Wrap(err, "failed to unmarshal service config JSON")
		}

		// Sourcegraph App Upserts() an external service with ID of ExtSVCID to the database and we
		// distinguish this in our returned results to discern which external services should not be modified
		// by users
		isAppAutogenerated := externalService.ID == servegit.ExtSVCID

		localExtSvc := localExternalServiceResolver{
			id:            graphqlbackend.MarshalExternalServiceID(externalService.ID),
			path:          otherConfig.Root,
			autogenerated: isAppAutogenerated,
		}
		localExternalServices = append(localExternalServices, localExtSvc)
	}

	return localExternalServices, nil
}

type localExternalServiceResolver struct {
	id            graphql.ID
	path          string
	autogenerated bool
}

func (r localExternalServiceResolver) ID() graphql.ID {
	return r.id
}

func (r localExternalServiceResolver) Path() string {
	return r.path
}

func (r localExternalServiceResolver) Autogenerated() bool {
	return r.autogenerated
}
