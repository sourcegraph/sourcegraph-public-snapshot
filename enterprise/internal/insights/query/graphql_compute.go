package query

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const computeSearchQuery = `
query Run($query: String!) {
  compute(query: $query) {
	__typename
    ... on ComputeMatchContext {
      repository {
        name
        id
      }
      commit
      path
      matches {
        value
        environment {
          variable
          value
        }
      }
    }
  }
}
`

type gqlComputeSearchResponse struct {
	Data struct {
		Compute []json.RawMessage
	}
	Errors []any
}

// ComputeSearch executes the given search query.
func ComputeSearch(ctx context.Context, query string) (results []ComputeResult, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "InsightsComputeSearch")
	defer func() {
		span.LogFields(
			log.Int("resultCount", len(results)),
			log.Error(err),
		)
		span.Finish()
	}()
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(graphQLQuery{
		Query:     computeSearchQuery,
		Variables: gqlSearchVars{Query: query},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Encode")
	}

	url, err := gqlURL("InsightsComputeSearch")
	if err != nil {
		return nil, errors.Wrap(err, "constructing frontend URL")
	}

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}

	req.Header.Set("Content-Type", "application/json")
	if span != nil {
		carrier := opentracing.HTTPHeadersCarrier(req.Header)
		span.Tracer().Inject(
			span.Context(),
			opentracing.HTTPHeaders,
			carrier)
	}

	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	var res *gqlComputeSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "Decode")
	}
	if len(res.Errors) > 0 {
		return nil, errors.Errorf("graphql: errors: %v", res.Errors)
	}

	for _, raw := range res.Data.Compute {
		decoded, err := decodeComputeResult(raw)
		if err != nil {
			return nil, err
		}
		results = append(results, decoded)
	}

	return results, nil
}
