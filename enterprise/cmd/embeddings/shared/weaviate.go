package shared

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type weaviateClient struct {
	logger log.Logger

	client    *weaviate.Client
	clientErr error
}

func newWeaviateClient(
	logger log.Logger,
	url *url.URL,
) *weaviateClient {
	if url == nil {
		return &weaviateClient{
			clientErr: errors.New("weaviate client is not configured"),
		}
	}

	client, err := weaviate.NewClient(weaviate.Config{
		Host:   url.Host,
		Scheme: url.Scheme,
	})

	return &weaviateClient{
		logger:    logger.Scoped("weaviate", "client for weaviate embedding index"),
		client:    client,
		clientErr: err,
	}
}

func (w *weaviateClient) Use(ctx context.Context) bool {
	return featureflag.FromContext(ctx).GetBoolOr("search-weaviate", false)
}

func (w *weaviateClient) Search(ctx context.Context, repoName api.RepoName, repoID api.RepoID, query []float32, codeResultsCount, textResultsCount int) (codeResults, textResults []embeddings.EmbeddingSearchResult, _ error) {
	if w.clientErr != nil {
		return nil, nil, w.clientErr
	}

	queryBuilder := func(klass string, limit int) *graphql.GetBuilder {
		return graphql.NewQueryClassBuilder(klass).
			WithNearVector((&graphql.NearVectorArgumentBuilder{}).
				WithVector(query)).
			WithFields([]graphql.Field{
				{Name: "file_name"},
				{Name: "start_line"},
				{Name: "end_line"},
				{Name: "revision"},
				{Name: "_additional", Fields: []graphql.Field{
					{Name: "distance"},
				}},
			}...).
			WithLimit(limit)
	}

	extractResults := func(res *models.GraphQLResponse, typ string) []embeddings.EmbeddingSearchResult {
		get := res.Data["Get"].(map[string]any)
		code := get[typ].([]any)
		if len(code) == 0 {
			return nil
		}

		srs := make([]embeddings.EmbeddingSearchResult, 0, len(code))
		revision := ""
		for _, c := range code {
			cMap := c.(map[string]any)
			fileName := cMap["file_name"].(string)

			if rev := cMap["revision"].(string); revision != rev {
				if revision == "" {
					revision = rev
				} else {
					w.logger.Warn("inconsistent revisions returned for an embedded repository", log.Int("repoid", int(repoID)), log.String("filename", fileName), log.String("revision1", revision), log.String("revision2", rev))
				}
			}

			// multiply by half max int32 since distance will always be between 0 and 2
			similarity := int32(cMap["_additional"].(map[string]any)["distance"].(float64) * (1073741823))

			srs = append(srs, embeddings.EmbeddingSearchResult{
				RepoName:  repoName,
				Revision:  api.CommitID(revision),
				FileName:  fileName,
				StartLine: int(cMap["start_line"].(float64)),
				EndLine:   int(cMap["end_line"].(float64)),
				ScoreDetails: embeddings.SearchScoreDetails{
					Score:           similarity,
					SimilarityScore: similarity,
				},
			})
		}

		return srs
	}

	// We partition the indexes by type and repository. Each class in
	// weaviate is its own index, so we achieve partitioning by a class
	// per repo and type.
	codeClass := fmt.Sprintf("Code_%d", repoID)
	textClass := fmt.Sprintf("Text_%d", repoID)

	res, err := w.client.GraphQL().MultiClassGet().
		AddQueryClass(queryBuilder(codeClass, codeResultsCount)).
		AddQueryClass(queryBuilder(textClass, textResultsCount)).
		Do(ctx)
	if err != nil {
		return nil, nil, errors.Wrap(err, "doing weaviate request")
	}

	if len(res.Errors) > 0 {
		return nil, nil, weaviateGraphQLError(res.Errors)
	}

	return extractResults(res, codeClass), extractResults(res, textClass), nil
}

type weaviateGraphQLError []*models.GraphQLError

func (errs weaviateGraphQLError) Error() string {
	var b strings.Builder
	b.WriteString("failed to query Weaviate:\n")
	for _, err := range errs {
		_, _ = fmt.Fprintf(&b, "- %s %s\n", strings.Join(err.Path, "."), err.Message)
	}
	return b.String()
}
