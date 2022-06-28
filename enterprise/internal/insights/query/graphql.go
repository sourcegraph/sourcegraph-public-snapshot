package query

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// This file contains all the methods required to execute Sourcegraph searches using our GraphQL
// API and get results back.

// graphQLQuery describes a general GraphQL query and its variables.
type graphQLQuery struct {
	Query     string `json:"query"`
	Variables any    `json:"variables"`
}

const gqlSearchQuery = `query Search(
	$query: String!,
) {
	search(query: $query, version: V2, patternType:literal) {
		results {
			limitHit
			cloning { name }
			missing { name }
			timedout { name }
			matchCount
			results {
				__typename
				... on FileMatch {
					repository {
						id
						name
					}
					lineMatches {
						offsetAndLengths
					}
					symbols {
						name
					}
				}
				... on CommitSearchResult {
					matches {
						highlights {
							line
						}
					}
					commit {
						repository {
							id
							name
						}
					}
				}
				... on Repository {
					id
					name
				}
			}
			alert {
				title
				description
			}
		}
	}
}`

type gqlSearchVars struct {
	Query string `json:"query"`
}

type GqlSearchResponse struct {
	Data struct {
		Search struct {
			Results struct {
				LimitHit   bool
				Cloning    []*api.Repo
				Missing    []*api.Repo
				Timedout   []*api.Repo
				MatchCount int
				Results    []json.RawMessage
				Alert      *struct {
					Title       string
					Description string
				}
			}
		}
	}
	Errors []any
}

// search executes the given search query.
func Search(ctx context.Context, query string) (_ *GqlSearchResponse, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "InsightsSearch")
	defer func() {
		span.LogFields(
			log.Error(err),
		)
		span.Finish()
	}()
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(graphQLQuery{
		Query:     gqlSearchQuery,
		Variables: gqlSearchVars{Query: query},
	})
	if err != nil {
		return nil, errors.Wrap(err, "Encode")
	}

	url, err := gqlURL("InsightsSearch")
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

	var res *GqlSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "Decode")
	}
	if len(res.Errors) > 0 {
		return res, errors.Errorf("graphql: errors: %v", res.Errors)
	}
	return res, nil
}

// gqlURL returns the frontend's internal GraphQL API URL, with the given ?queryName parameter
// which is used to keep track of the source and type of GraphQL queries.
func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}
