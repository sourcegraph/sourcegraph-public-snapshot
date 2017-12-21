package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var (
	SearchQueriesEnv = env.Get("LOAD_SEARCH_QUERIES", "", "Search queries to use in load testing")
	QueryPeriodMSEnv = env.Get("LOAD_SEARCH_QUERY_PERIOD", "", "Period of search query issuance (milliseconds). E.g., a value of 200 corresponds to 200ms or 5 QPS")
)

type GQLSearchVars struct {
	Query      string `json:"query"`
	ScopeQuery string `json:"scopeQuery"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	var searchQueries []GQLSearchVars
	if err := json.Unmarshal([]byte(SearchQueriesEnv), &searchQueries); err != nil {
		return err
	}

	qps, err := strconv.Atoi(QueryPeriodMSEnv)
	if err != nil {
		return err
	}

	ticker := time.NewTicker(time.Duration(qps) * time.Millisecond)
	for {
		for _, v := range searchQueries {
			<-ticker.C
			go func(v GQLSearchVars) (err error) {
				defer func() {
					if err != nil {
						log15.Error("Error issuing search query", "query", v.Query, "scopeQuery", v.ScopeQuery, "error", err)
					}
				}()

				gqlQuery := GraphQLQuery{Query: gqlSearch, Variables: v}
				b, err := json.Marshal(gqlQuery)
				if err != nil {
					return err
				}
				resp, err := http.Post("http://localhost:3080/.api/graphql?Search", "application/json", bytes.NewReader(b))
				if err != nil {
					return err
				}
				var res GraphQLResponseSearch
				if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
					return err
				}
				log15.Info("Search results", "query", v.Query, "scopeQuery", v.ScopeQuery, "resultCount", len(res.Data.Search.Results.Results))
				return nil
			}(v)
		}
	}
}

type GraphQLResponseSearch struct {
	Data struct {
		Search struct {
			Results struct {
				Results []interface{} `json:"results"`
			} `json:"results"`
		} `json:"search"`
	} `json:"data"`
}

type GraphQLQuery struct {
	Query     string      `json:"query"`
	Variables interface{} `json:"variables"`
}

const gqlSearch = `query Search(
	$query: String!,
	$scopeQuery: String!,
) {
	search(query: $query, scopeQuery: $scopeQuery) {
		results {
			limitHit
			missing
			cloning
			timedout
			results {
				__typename
				... on FileMatch {
					resource
					limitHit
					lineMatches {
						preview
						lineNumber
						offsetAndLengths
					}
				}
				... on CommitSearchResult {
					refs {
						name
						displayName
						prefix
						repository { uri }
					}
					sourceRefs {
						name
						displayName
						prefix
						repository { uri }
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
							uri
						}
						oid
						abbreviatedOID
						author {
							person {
								displayName
								avatarURL
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
					query {
						query
						scopeQuery
					}
				}
			}
		}
	}
}
`
