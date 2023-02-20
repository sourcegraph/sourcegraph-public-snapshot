package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	contextdetectionbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/contextdetection"
	repobg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background/repo"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func NewResolver(db database.DB, gitserverClient gitserver.Client, repoStore repobg.RepoEmbeddingJobsStore, contextDetectionStore contextdetectionbg.ContextDetectionEmbeddingJobsStore) graphqlbackend.EmbeddingsResolver {
	return &Resolver{db: db, gitserverClient: gitserverClient, repoStore: repoStore, contextDetectionStore: contextDetectionStore}
}

type Resolver struct {
	db                    database.DB
	gitserverClient       gitserver.Client
	repoStore             repobg.RepoEmbeddingJobsStore
	contextDetectionStore contextdetectionbg.ContextDetectionEmbeddingJobsStore
}

func (r *Resolver) EmbeddingsSearch(ctx context.Context, args graphqlbackend.EmbeddingsSearchInputArgs) (graphqlbackend.EmbeddingsSearchResultsResolver, error) {
	repo, err := r.db.Repos().GetByName(ctx, api.RepoName(args.RepoName))
	if err != nil {
		return nil, err
	}

	results, err := embeddings.NewClient().Search(ctx, embeddings.EmbeddingsSearchParameters{
		RepoName:         repo.Name,
		Query:            args.Query,
		CodeResultsCount: int(args.CodeResultsCount),
		TextResultsCount: int(args.TextResultsCount),
	})
	if err != nil {
		return nil, err
	}

	return &embeddingsSearchResultsResolver{results}, nil
}

func (r *Resolver) IsContextRequiredForQuery(ctx context.Context, args graphqlbackend.IsContextRequiredForQueryInputArgs) (bool, error) {
	repo, err := r.db.Repos().GetByName(ctx, api.RepoName(args.RepoName))
	if err != nil {
		return false, err
	}
	return embeddings.NewClient().IsContextRequiredForQuery(ctx, embeddings.IsContextRequiredForQueryParameters{RepoName: repo.Name, Query: args.Query})
}

func (r *Resolver) ScheduleRepositoriesForEmbedding(ctx context.Context, args graphqlbackend.ScheduleRepositoriesForEmbeddingArgs) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// TODO: Check if repos exists, and check if repo + revision embedding job already exists and is not completed
	repoStore := r.db.Repos()
	for _, repoName := range args.RepoNames {
		repo, err := repoStore.GetByName(ctx, api.RepoName(repoName))
		if err != nil {
			return nil, err
		}
		latestRevision, err := r.gitserverClient.ResolveRevision(ctx, repo.Name, "", gitserver.ResolveRevisionOptions{})
		if err != nil {
			return nil, err
		}
		_, err = r.repoStore.CreateRepoEmbeddingJob(ctx, repo.ID, latestRevision)
		if err != nil {
			return nil, err
		}
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (r *Resolver) ScheduleContextDetectionForEmbedding(ctx context.Context) (*graphqlbackend.EmptyResponse, error) {
	// 🚨 SECURITY: Only site admins may schedule embedding jobs.
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	_, err := r.contextDetectionStore.CreateContextDetectionEmbeddingJob(ctx)
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
