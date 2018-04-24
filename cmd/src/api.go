package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
)

func init() {
	usage := `
Examples:

  Run queries (identical behavior):

    	$ echo 'query { currentUser { username } }' | src api
    	$ src api -query='query { currentUser { username } }'

  Specify query variables:

    	$ echo '<query>' | src api 'var1=val1' 'var2=val2'

  Searching for "Router" and getting result count:

    	$ echo 'query($query: String!) { search(query: $query) { results { resultCount } } }' | src api 'query=Router'

  Get the curl command for a query (just add '-get-curl' in the flags section):

    	$ src api -get-curl -query='query { currentUser { username } }'

`

	flagSet := flag.NewFlagSet("api", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	queryFlag := flagSet.String("query", "", "GraphQL query to execute, e.g. 'query { currentUser { username } }' (stdin otherwise)")
	varsFlag := flagSet.String("vars", "", `GraphQL query variables to include as JSON string, e.g. '{"var": "val", "var2": "val2"}'`)
	getCurlFlag := flagSet.Bool("get-curl", false, "Print the curl command for executing this query and exit (WARNING: includes printing your access token!)")

	handler := func(args []string) error {
		flagSet.Parse(args)

		// Build the GraphQL request.
		query := *queryFlag
		if query == "" {
			// Read query from stdin instead.
			if isatty.IsTerminal(os.Stdin.Fd()) {
				return &usageError{errors.New("expected query to be piped into 'src api' or -query flag to be specified")}
			}
			data, err := ioutil.ReadAll(os.Stdin)
			if err != nil {
				return err
			}
			query = string(data)
		}

		// Determine which variables to use in the request.
		vars := map[string]interface{}{}
		if *varsFlag != "" {
			if err := json.Unmarshal([]byte(*varsFlag), &vars); err != nil {
				return err
			}
		}
		for _, arg := range flagSet.Args() {
			idx := strings.Index(arg, "=")
			if idx == -1 {
				return &usageError{fmt.Errorf("parsing argument %q expected 'variable=value' syntax (missing equals)", arg)}
			}
			key := arg[:idx]
			value := arg[idx+1:]
			vars[key] = value
		}

		// Handle the get-curl flag now.
		if *getCurlFlag {
			curl, err := curlCmd(cfg.Endpoint, cfg.AccessToken, query, vars)
			if err != nil {
				return err
			}
			fmt.Println(curl)
			return nil
		}

		var buf bytes.Buffer
		if err := json.NewEncoder(&buf).Encode(map[string]interface{}{
			"query":     query,
			"variables": vars,
		}); err != nil {
			return err
		}

		// Create the HTTP request.
		req, err := http.NewRequest("POST", gqlURL(cfg.Endpoint), nil)
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "token "+cfg.AccessToken)
		req.Body = ioutil.NopCloser(&buf)

		// Perform the request.
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Read the response.
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		// Print the formatted JSON.
		f, err := formatJSON(data)
		if err != nil {
			return err
		}
		fmt.Println(string(f))
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

// gqlURL returns the URL to the GraphQL endpoint for the given Sourcegraph
// instance.
func gqlURL(endpoint string) string {
	if !strings.HasSuffix(endpoint, "/") {
		endpoint = endpoint + "/"
	}
	return endpoint + ".api/graphql"
}

// formatJSON formats the given JSON data with a two-space indent.
func formatJSON(data []byte) ([]byte, error) {
	var v map[string]interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, err
	}
	return json.MarshalIndent(v, "", "  ")
}

// curlCmd returns the curl command to perform the given GraphQL query. Bash-only.
func curlCmd(endpoint, accessToken, query string, vars map[string]interface{}) (string, error) {
	data, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": vars,
	})
	if err != nil {
		return "", err
	}

	s := fmt.Sprintf("curl \\\n")
	s += fmt.Sprintf("   %s \\\n", shellquote.Join("-H", "Authorization: token "+accessToken))
	s += fmt.Sprintf("   %s \\\n", shellquote.Join("-d", string(data)))
	s += fmt.Sprintf("   %s", shellquote.Join(gqlURL(endpoint)))
	return s, nil
}
