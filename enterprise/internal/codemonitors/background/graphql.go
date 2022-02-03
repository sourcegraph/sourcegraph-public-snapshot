package background

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/hashicorp/go-multierror"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/log"

	cmtypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codemonitors/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type graphQLQuery struct {
	Query     string        `json:"query"`
	Variables gqlSearchVars `json:"variables"`
}

type gqlSearchVars struct {
	Query     string      `json:"query"`
	MonitorID *graphql.ID `json:"monitorID"`
}

type gqlSearchResponse struct {
	Data struct {
		CodeMonitorSearch struct {
			Results searchResults `json:"results"`
		} `json:"codeMonitorSearch"`
	} `json:"data"`
	Errors []gqlerrors.FormattedError
}

type searchResults struct {
	ApproximateResultCount string                      `json:"approximateResultCount"`
	Cloning                []api.Repo                  `json:"cloning,omitempty"`
	Timedout               []api.Repo                  `json:"timedout,omitempty"`
	Results                cmtypes.CommitSearchResults `json:"results"`
}

const gqlSearchQuery = `query CodeMonitorSearch(
	$query: String!,
	$monitorID: ID,
) {
	codeMonitorSearch(query: $query, codeMonitorID: $monitorID) {
		results {
			approximateResultCount
			limitHit
			cloning { name }
			timedout { name }
			results {
				__typename
				... on CommitSearchResult {
					refs {
						name
						displayName
						prefix
					}
					sourceRefs {
						name
						displayName
						prefix
					}
					messagePreview {
						value
						highlights {
							line
							character
							length
						}
					}
					diffPreview {
						value
						highlights {
							line
							character
							length
						}
					}
					commit {
						repository {
							name
						}
						oid
						author {
							person {
								displayName
							}
							date
						}
						committer {
							person {
								displayName
							}
							date
						}
						message
					}
				}
			}
			alert {
				title
				description
				proposedQueries {
					description
					query
				}
			}
		}
	}
}`

func search(ctx context.Context, query string, monitorID *int64) (_ *searchResults, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "CodeMonitorSearch")
	defer func() {
		span.LogFields(log.Error(err))
		span.Finish()
	}()

	vars := gqlSearchVars{Query: query}
	if monitorID != nil {
		id := relay.MarshalID("CodeMonitorID", *monitorID)
		vars.MonitorID = &id
	}

	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(graphQLQuery{
		Query:     gqlSearchQuery,
		Variables: vars,
	})
	if err != nil {
		return nil, errors.Wrap(err, "Encode")
	}

	url, err := gqlURL("CodeMonitorSearch")
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
			carrier,
		)
	}
	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	var res gqlSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, errors.Wrap(err, "Decode")
	}
	if len(res.Errors) > 0 {
		var combined error
		for _, err := range res.Errors {
			combined = multierror.Append(combined, err)
		}
		return nil, combined
	}
	return &res.Data.CodeMonitorSearch.Results, nil
}

func gqlURL(queryName string) (string, error) {
	u, err := url.Parse(internalapi.Client.URL)
	if err != nil {
		return "", err
	}
	u.Path = "/.internal/graphql"
	u.RawQuery = queryName
	return u.String(), nil
}

// extractTime extracts the time from the given search result.
func extractTime(result cmtypes.CommitSearchResult) (time.Time, error) {
	// This relies on the date format that our API returns. It was previously broken
	// and should be checked first in case date extraction stops working.
	return time.Parse(time.RFC3339, result.Commit.Committer.Date)
}
