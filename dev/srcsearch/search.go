package main

import (
	"flag"
	"fmt"
	"github.com/pkg/errors"
)

func search() error {
	flag.Parse()
	if flag.NArg() != 1 {
		return &usageError{errors.New("expected exactly one argument: the search query")}
	}
	queryString := flag.Arg(0)

	query := `fragment FileMatchFields on FileMatch {
				repository {
					name
					url
				}
				file {
					name
					path
					url
					commit {
						oid
					}
				}
				lineMatches {
					preview
					lineNumber
					offsetAndLengths
					limitHit
				}
			}

			fragment CommitSearchResultFields on CommitSearchResult {
				messagePreview {
					value
					highlights{
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
				label {
					html
				}
				url
				matches {
					url
					body {
						html
						text
					}
					highlights {
						character
						line
						length
					}
				}
				commit {
					repository {
						name
					}
					oid
					url
					subject
					author {
						date
						person {
							displayName
						}
					}
				}
			}

		  fragment RepositoryFields on Repository {
			name
			url
			externalURLs {
			  serviceType
			  url
			}
			label {
				html
			}
		  }

		  query ($query: String!) {
			site {
				buildVersion
			}
			search(query: $query) {
			  results {
				results{
				  __typename
				  ... on FileMatch {
					...FileMatchFields
				  }
				  ... on CommitSearchResult {
					...CommitSearchResultFields
				  }
				  ... on Repository {
					...RepositoryFields
				  }
				}
				limitHit
				cloning {
				  name
				}
				missing {
				  name
				}
				timedout {
				  name
				}
				resultCount
				elapsedMilliseconds
			  }
			}
		  }
		`

	var result struct {
		Site struct {
			BuildVersion string
		}
		Search struct {
			Results searchResults
		}
	}

	// Parse config.
	cfg, err := readConfig()
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	return (&apiRequest{
		query: query,
		vars: map[string]interface{}{
			"query": nullString(queryString),
		},
		result: &result,
		done: func() error {
			improved := searchResultsImproved{
				SourcegraphEndpoint: cfg.Endpoint,
				Query:               queryString,
				Site:                result.Site,
				searchResults:       result.Search.Results,
			}

			// Print the formatted JSON.
			f, err := marshalIndent(improved)
			if err != nil {
				return err
			}
			fmt.Println(string(f))
			return nil
		},
		endpoint:    cfg.Endpoint,
		accessToken: cfg.AccessToken,
	}).do()
}

// searchResults represents the data we get back from the GraphQL search request.
type searchResults struct {
	Results                    []map[string]interface{}
	LimitHit                   bool
	Cloning, Missing, Timedout []map[string]interface{}
	ResultCount                int
	ElapsedMilliseconds        int
}

// searchResultsImproved is a superset of what the GraphQL API returns. It
// contains the query responsible for producing the results, which is nice for
// most consumers.
type searchResultsImproved struct {
	SourcegraphEndpoint string
	Query               string
	Site                struct{ BuildVersion string }
	searchResults
}

