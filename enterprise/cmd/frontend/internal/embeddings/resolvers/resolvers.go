package resolvers

import (
	"context"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

func NewResolver(db database.DB) graphqlbackend.EmbeddingsResolver {
	return &Resolver{db: db}
}

type Resolver struct {
	db database.DB
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
