package background

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/hashicorp/go-multierror"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/api/internalapi"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

type graphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

const gqlSearchQuery = `query CodeMonitorSearch(
	$query: String!,
) {
	search(query: $query) {
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

type gqlSearchVars struct {
	Query string `json:"query"`
}

type gqlSearchResponse struct {
	Data struct {
		Search struct {
			Results struct {
				ApproximateResultCount string
				Cloning                []api.Repo
				Timedout               []api.Repo
				Results                commitSearchResults
			} `json:"results"`
		} `json:"search"`
	} `json:"data"`
	Errors []gqlerrors.FormattedError
}

type commitSearchResults []commitSearchResult

func (c *commitSearchResults) UnmarshalJSON(b []byte) error {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(b, &rawMessages); err != nil {
		return err
	}

	var results []commitSearchResult
	for _, rawMessage := range rawMessages {
		var t struct {
			Typename string `json:"__typename"`
		}
		if err := json.Unmarshal(rawMessage, &t); err != nil {
			return err
		}
		if t.Typename != "CommitSearchResult" {
			return errors.Errorf("expected result type %q, got %q", "CommitSearchResult", t.Typename)
		}

		var csr commitSearchResult
		if err := json.Unmarshal(rawMessage, &csr); err != nil {
			return err
		}

		results = append(results, csr)
	}
	*c = results
	return nil
}

type commitSearchResult struct {
	Refs           []ref              `json:"refs"`
	SourceRefs     []ref              `json:"sourceRefs"`
	MessagePreview *highlightedString `json:"messagePreview"`
	DiffPreview    *highlightedString `json:"diffPreview"`
	Commit         commit             `json:"commit"`
}

type ref struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Prefix      string `json:"prefix"`
}

type highlightedString struct {
	Value      string `json:"value"`
	Highlights []struct {
		Line      int `json:"line"`
		Character int `json:"character"`
		Length    int `json:"length"`
	} `json:"highlights"`
}

type commit struct {
	Repository struct {
		Name string `json:"name"`
	} `json:"repository"`
	Oid     string `json:"oid"`
	Message string `json:"message"`
	Author  struct {
		Person struct {
			DisplayName string `json:"displayName"`
		} `json:"person"`
		Date string `json:"date"`
	} `json:"author"`
}

func search(ctx context.Context, query string, userID int32) (*gqlSearchResponse, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(graphQLQuery{
		Query:     gqlSearchQuery,
		Variables: gqlSearchVars{Query: query},
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
	resp, err := httpcli.InternalDoer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "Post")
	}
	defer resp.Body.Close()

	var res *gqlSearchResponse
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
	return res, nil
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
func extractTime(result commitSearchResult) (time.Time, error) {
	// This relies on the date format that our API returns. It was previously broken
	// and should be checked first in case date extraction stops working.
	return time.Parse(time.RFC3339, result.Commit.Author.Date)
}
