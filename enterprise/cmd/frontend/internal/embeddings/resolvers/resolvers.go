package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	embeddingsbg "github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings/background"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
)

func NewResolver(db database.DB, store embeddingsbg.RepoEmbeddingJobsStore, gitserverClient gitserver.Client) graphqlbackend.EmbeddingsResolver {
	return &Resolver{db: db, gitserverClient: gitserverClient, store: store}
}

type Resolver struct {
	db              database.DB
	gitserverClient gitserver.Client
	store           embeddingsbg.RepoEmbeddingJobsStore
}

func (r *Resolver) EmbeddingsSearch(ctx context.Context, args graphqlbackend.EmbeddingsSearchInputArgs) (graphqlbackend.EmbeddingsSearchResultsResolver, error) {
	repoID, err := graphqlbackend.UnmarshalRepositoryID(args.Repository)
	if err != nil {
		return nil, err
	}

	repo, err := r.db.Repos().Get(ctx, repoID)
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

func (r *Resolver) ScheduleRepositoriesForEmbedding(ctx context.Context, args graphqlbackend.ScheduleRepositoriesForEmbeddingArgs) (*graphqlbackend.EmptyResponse, error) {
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
		_, err = r.store.CreateRepoEmbeddingJob(ctx, repo.ID, latestRevision)
		if err != nil {
			return nil, err
		}
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

func (r *embeddingsSearchResultResolver) FilePath(ctx context.Context) string {
	return r.result.FilePath
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
