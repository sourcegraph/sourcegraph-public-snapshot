package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	contextdetectionbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	repobg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/cody"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func NewResolver(
	db database.DB,
	gitserverClient gitserver.Client,
	embeddingsClient *embeddings.Client,
	repoStore repobg.RepoEmbeddingJobsStore,
	contextDetectionStore contextdetectionbg.ContextDetectionEmbeddingJobsStore,
) graphqlbackend.EmbeddingsResolver {
	return &Resolver{
		db:                        db,
		gitserverClient:           gitserverClient,
		embeddingsClient:          embeddingsClient,
		repoEmbeddingJobsStore:    repoStore,
		contextDetectionJobsStore: contextDetectionStore,
	}
}

type Resolver struct {
	db                        database.DB
	gitserverClient           gitserver.Client
	embeddingsClient          *embeddings.Client
	repoEmbeddingJobsStore    repobg.RepoEmbeddingJobsStore
	contextDetectionJobsStore contextdetectionbg.ContextDetectionEmbeddingJobsStore
}

func (r *Resolver) EmbeddingsSearch(ctx context.Context, args graphqlbackend.EmbeddingsSearchInputArgs) (graphqlbackend.EmbeddingsSearchResultsResolver, error) {
	if !conf.EmbeddingsEnabled() {
		return nil, errors.New("embeddings are not configured or disabled")
	}

	if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
		return nil, errors.New("cody experimental feature flag is not enabled for current user")
	}

	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repo)
	if err != nil {
		return nil, err
	}

	repo, err := r.db.Repos().Get(ctx, repoID)
	if err != nil {
		return nil, err
	}

	results, err := r.embeddingsClient.Search(ctx, embeddings.EmbeddingsSearchParameters{
		RepoName:         repo.Name,
		RepoID:           repoID,
		Query:            args.Query,
		CodeResultsCount: int(args.CodeResultsCount),
		TextResultsCount: int(args.TextResultsCount),
	})
	if err != nil {
		return nil, err
	}

	return &embeddingsSearchResultsResolver{results}, nil
}

func (r *Resolver) IsContextRequiredForChatQuery(ctx context.Context, args graphqlbackend.IsContextRequiredForChatQueryInputArgs) (bool, error) {
	if !conf.EmbeddingsEnabled() {
		return false, errors.New("embeddings are not configured or disabled")
	}
	if isEnabled := cody.IsCodyEnabled(ctx); !isEnabled {
		return false, errors.New("cody experimental feature flag is not enabled for current user")
	}
	return r.embeddingsClient.IsContextRequiredForChatQuery(ctx, embeddings.IsContextRequiredForChatQueryParameters{Query: args.Query})
}

func (r *Resolver) RepoEmbeddingJobs(ctx context.Context, args graphqlbackend.ListRepoEmbeddingJobsArgs) (*graphqlutil.ConnectionResolver[graphqlbackend.RepoEmbeddingJobResolver], error) {
	if !conf.EmbeddingsEnabled() {
		return nil, errors.New("embeddings are not configured or disabled")
	}
	// 🚨 SECURITY: Only site admins may list repo embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	return NewRepoEmbeddingJobsResolver(r.db, r.gitserverClient, r.repoEmbeddingJobsStore, args)
}

func isRepoEmbeddingJobScheduledOrCompleted(job *repobg.RepoEmbeddingJob) bool {
	return job != nil && (job.State == "completed" || job.State == "processing" || job.State == "queued")
}

func (r *Resolver) ScheduleRepositoriesForEmbedding(ctx context.Context, args graphqlbackend.ScheduleRepositoriesForEmbeddingArgs) (_ *graphqlbackend.EmptyResponse, err error) {
	if !conf.EmbeddingsEnabled() {
		return nil, errors.New("embeddings are not configured or disabled")
	}

	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err = auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	tx, err := r.repoEmbeddingJobsStore.Transact(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { err = tx.Done(err) }()

	repoStore := r.db.Repos()
	for _, repoName := range args.RepoNames {
		// Scope the iteration to an anonymous function so we can capture all errors and properly rollback tx in defer above.
		err = func() error {
			repo, err := repoStore.GetByName(ctx, api.RepoName(repoName))
			if err != nil {
				return err
			}

			refName, latestRevision, err := r.gitserverClient.GetDefaultBranch(ctx, repo.Name, false)
			if err != nil {
				return err
			}
			if refName == "" {
				return errors.Newf("could not get latest commit for repo %s", repo.Name)
			}

			job, _ := tx.GetLastRepoEmbeddingJobForRevision(ctx, repo.ID, latestRevision)
			// Skip creating a repo embedding job for a repo at revision, if there already exists
			// an identical job that has been completed or is scheduled to run (processing or queued).
			if isRepoEmbeddingJobScheduledOrCompleted(job) {
				return nil
			}

			_, err = tx.CreateRepoEmbeddingJob(ctx, repo.ID, latestRevision)
			return err
		}()
		if err != nil {
			return nil, err
		}
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleContextDetectionForEmbedding(ctx context.Context) (*graphqlbackend.EmptyResponse, error) {
	if !conf.EmbeddingsEnabled() {
		return nil, errors.New("embeddings are not configured or disabled")
	}

	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	_, err := r.contextDetectionJobsStore.CreateContextDetectionEmbeddingJob(ctx)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

type embeddingsSearchResultsResolver struct {
	results *embeddings.EmbeddingSearchResults
}

func (r *embeddingsSearchResultsResolver) CodeResults(ctx context.Context) []graphqlbackend.EmbeddingsSearchResultResolver {
	codeResults := make([]graphqlbackend.EmbeddingsSearchResultResolver, len(r.results.CodeResults))
	for idx, result := range r.results.CodeResults {
		codeResults[idx] = &embeddingsSearchResultResolver{result}
	}
	return codeResults
}

func (r *embeddingsSearchResultsResolver) TextResults(ctx context.Context) []graphqlbackend.EmbeddingsSearchResultResolver {
	textResults := make([]graphqlbackend.EmbeddingsSearchResultResolver, len(r.results.TextResults))
	for idx, result := range r.results.TextResults {
		textResults[idx] = &embeddingsSearchResultResolver{result}
	}
	return textResults
}

type embeddingsSearchResultResolver struct {
	result embeddings.EmbeddingSearchResult
}

func (r *embeddingsSearchResultResolver) FileName(ctx context.Context) string {
	return r.result.FileName
}

func (r *embeddingsSearchResultResolver) StartLine(ctx context.Context) int32 {
	return int32(r.result.StartLine)
}

func (r *embeddingsSearchResultResolver) EndLine(ctx context.Context) int32 {
	return int32(r.result.EndLine)
}

func (r *embeddingsSearchResultResolver) Content(ctx context.Context) string {
	return r.result.Content
}
