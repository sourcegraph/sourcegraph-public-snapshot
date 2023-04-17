package shared

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/sourcegraph/log"
	"github.com/weaviate/weaviate-go-client/v4/weaviate"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/filters"
	"github.com/weaviate/weaviate-go-client/v4/weaviate/graphql"
	"github.com/weaviate/weaviate/entities/models"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/embeddings"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type weaviateClient struct {
	logger            log.Logger
	readFile          readFileFn
	getQueryEmbedding getQueryEmbeddingFn

	client    *weaviate.Client
	clientErr error
}

func newWeaviateClient(
	logger log.Logger,
	readFile readFileFn,
	getQueryEmbedding getQueryEmbeddingFn,
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
		logger:            logger.Scoped("weaviate", "client for weaviate embedding index"),
		readFile:          readFile,
		getQueryEmbedding: getQueryEmbedding,
		client:            client,
		clientErr:         err,
	}
}

func (w *weaviateClient) Use(ctx context.Context) bool {
	return featureflag.FromContext(ctx).GetBoolOr("search-weaviate", false)
}

func (w *weaviateClient) Search(ctx context.Context, params embeddings.EmbeddingsSearchParameters) (*embeddings.EmbeddingSearchResults, error) {
	if w.clientErr != nil {
		return nil, w.clientErr
	}

	embeddedQuery, err := w.getQueryEmbedding(params.Query)
	if err != nil {
		return nil, errors.Wrap(err, "getting query embedding")
	}

	queryBuilder := func(typ string, limit int) *graphql.GetBuilder {
		return graphql.NewQueryClassBuilder(typ).
			WithNearVector((&graphql.NearVectorArgumentBuilder{}).
				WithVector(embeddedQuery)).
			WithWhere(filters.Where().
				WithOperator(filters.Equal).
				WithPath([]string{"repo"}).
				WithValueString(string(params.RepoName))).
			WithFields([]graphql.Field{
				{Name: "file_name"},
				{Name: "start_line"},
				{Name: "end_line"},
			}...).
			WithLimit(limit)
	}

	extractResults := func(res *models.GraphQLResponse, typ string) []embeddings.EmbeddingSearchResult {
		get := res.Data["Get"].(map[string]any)
		code := get[typ].([]any)
		srs := make([]embeddings.EmbeddingSearchResult, 0, len(code))
		for _, c := range code {
			cMap := c.(map[string]any)
			srs = append(srs, embeddings.EmbeddingSearchResult{
				RepoEmbeddingRowMetadata: embeddings.RepoEmbeddingRowMetadata{
					FileName:  cMap["file_name"].(string),
					StartLine: int(cMap["start_line"].(float64)),
					EndLine:   int(cMap["end_line"].(float64)),
				},
			})
		}

		// TODO store in index the commit. Future we more likely is that we
		// will store contents in weaviate for keyword search.
		commit := api.CommitID("HEAD")

		return filterAndHydrateContent(ctx, w.logger, params.RepoName, commit, w.readFile, srs)
	}

	res, err := w.client.GraphQL().MultiClassGet().
		AddQueryClass(queryBuilder("Code", params.CodeResultsCount)).
		AddQueryClass(queryBuilder("Text", params.TextResultsCount)).
		Do(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "doing weaviate request")
	}

	if len(res.Errors) > 0 {
		return nil, weaviateGraphQLError(res.Errors)
	}

	return &embeddings.EmbeddingSearchResults{
		CodeResults: extractResults(res, "Code"),
		TextResults: extractResults(res, "Text"),
	}, nil
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
