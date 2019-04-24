package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

func main() {
	usage := `srcsearch runs a search against a Sourcegraph instance.

Usage:

	srcsearch [options] query

The options are:

	-config=$HOME/src-config.json    specifies a file containing {"accessToken": "<secret>", "endpoint": "https://sourcegraph.com"}
	-endpoint=                       specifies the endpoint to use e.g. "https://sourcegraph.com" (overrides -config, if any)

Examples:

  Perform a search and get results in JSON format:

        $ srcsearch 'repogroup:sample error'

Other tips:

  Query syntax: https://about.sourcegraph.com/docs/search/query-syntax/
`

	// Configure logging.
	log.SetFlags(0)
	log.SetPrefix("")
	configPath := flag.String("config", "", "")
	endpoint := flag.String("endpoint", "", "")
	flag.Parse()
	if flag.NArg() != 1 {
		log.Println("expected exactly one argument: the search query")
		log.Println(usage)
		os.Exit(1)
	}
	searchQuery := flag.Arg(0)

	if err := srcsearch(*configPath, *endpoint, searchQuery); err != nil {
		log.Fatalf("srcsearch: %v", err)
	}
}

func srcsearch(configPath, endpoint, searchQuery string) error {
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

	cfg, err := readConfig(configPath, endpoint)
	if err != nil {
		return errors.Wrap(err, "reading config")
	}

	return (&apiRequest{
		query: query,
		vars: map[string]interface{}{
			"query": nullString(searchQuery),
		},
		result: &result,
		done: func() error {
			improved := searchResultsImproved{
				SourcegraphEndpoint: cfg.Endpoint,
				Query:               searchQuery,
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

// gqlURL returns the URL to the GraphQL endpoint for the given Sourcegraph
// instance.
func gqlURL(endpoint string) string {
	return endpoint + "/.api/graphql"
}

// apiRequest represents a GraphQL API request.
type apiRequest struct {
	query       string                 // the GraphQL query
	vars        map[string]interface{} // the GraphQL query variables
	result      interface{}            // where to store the result
	done        func() error           // a function to invoke for handling the response. If nil, flags like -get-curl are ignored.
	endpoint    string
	accessToken string
}

// do performs the API request. Once the request is finished a.done is invoked to
// handle the response (which is stored in a.result).
func (a *apiRequest) do() error {
	// Create the JSON object.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"query":     a.query,
		"variables": a.vars,
	}); err != nil {
		return err
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", gqlURL(a.endpoint), nil)
	if err != nil {
		return err
	}
	if a.accessToken != "" {
		req.Header.Set("Authorization", "token "+a.accessToken)
	}
	req.Body = ioutil.NopCloser(&buf)

	// Perform the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Our request may have failed before the reaching GraphQL endpoint, so
	// confirm the status code. You can test this easily with e.g. an invalid
	// endpoint like -endpoint=https://google.com
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusUnauthorized && isatty.IsCygwinTerminal(os.Stdout.Fd()) {
			fmt.Println("You may need to specify or update your access token to use this endpoint.")
			fmt.Println("See https://github.com/sourcegraph/src-cli#authentication")
			fmt.Println("")
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return fmt.Errorf("error: %s\n\n%s", resp.Status, body)
	}

	// Decode the response.
	var result struct {
		Data   interface{} `json:"data,omitempty"`
		Errors interface{} `json:"errors,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result.Errors != nil {
		return fmt.Errorf("GraphQL errors:\n%s", &graphqlError{result.Errors})
	}
	if err := jsonCopy(a.result, result.Data); err != nil {
		return err
	}
	return a.done()
}

// jsonCopy is a cheaty method of copying an already-decoded JSON (src)
// response into its destination (dst) that would usually be passed to e.g.
// json.Unmarshal.
//
// We could do this with reflection, obviously, but it would be much more
// complex and JSON re-marshaling should be cheap enough anyway. Can improve in
// the future.
func jsonCopy(dst, src interface{}) error {
	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.NewDecoder(bytes.NewReader(data)).Decode(dst)
}

type graphqlError struct {
	Errors interface{}
}

func (g *graphqlError) Error() string {
	j, _ := marshalIndent(g.Errors)
	return string(j)
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// json.MarshalIndent, but with defaults.
func marshalIndent(v interface{}) ([]byte, error) {
	return json.MarshalIndent(v, "", "  ")
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

// config represents the config format.
type config struct {
	Endpoint    string `json:"endpoint"`
	AccessToken string `json:"accessToken"`
}

// readConfig reads the config file from the given path.
func readConfig(configPath, endpoint string) (*config, error) {
	cfgPath := configPath
	userSpecified := configPath != ""
	if !userSpecified {
		u, err := user.Current()
		if err != nil {
			return nil, err
		}
		cfgPath = filepath.Join(u.HomeDir, "src-config.json")
	}
	data, err := ioutil.ReadFile(os.ExpandEnv(cfgPath))
	if err != nil && (!os.IsNotExist(err) || userSpecified) {
		return nil, err
	}
	var cfg config
	if err == nil {
		if err := json.Unmarshal(data, &cfg); err != nil {
			return nil, err
		}
	}

	// Apply config overrides.
	if envToken := os.Getenv("SRC_ACCESS_TOKEN"); envToken != "" {
		cfg.AccessToken = envToken
	}
	if endpoint != "" {
		cfg.Endpoint = strings.TrimSuffix(endpoint, "/")
	}
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://sourcegraph.com"
	}
	return &cfg, nil
}

type User struct {
	ID            string
	Username      string
	DisplayName   string
	SiteAdmin     bool
	Organizations struct {
		Nodes []Org
	}
	Emails []UserEmail
	URL    string
}

type UserEmail struct {
	Email    string
	Verified bool
}

type Org struct {
	ID          string
	Name        string
	DisplayName string
	Members     struct {
		Nodes []User
	}
}
