package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/mattn/go-isatty"
)

func init() {
	usage := `
Exit codes:

  0: Success
  1: General failures (connection issues, invalid HTTP response, etc.)
  2: GraphQL error response

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
	var (
		queryFlag = flagSet.String("query", "", "GraphQL query to execute, e.g. 'query { currentUser { username } }' (stdin otherwise)")
		varsFlag  = flagSet.String("vars", "", `GraphQL query variables to include as JSON string, e.g. '{"var": "val", "var2": "val2"}'`)
		apiFlags  = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		err := flagSet.Parse(args)
		if err != nil {
			return err
		}

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

		// Perform the request.
		var result interface{}
		return (&apiRequest{
			query:  query,
			vars:   vars,
			result: &result,
			done: func() error {
				// Print the formatted JSON.
				f, err := marshalIndent(result)
				if err != nil {
					return err
				}
				fmt.Println(string(f))
				return nil
			},
			flags:            apiFlags,
			dontUnpackErrors: true,
		}).do()
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
	return endpoint + "/.api/graphql"
}

// curlCmd returns the curl command to perform the given GraphQL query. Bash-only.
func curlCmd(endpoint, accessToken string, additionalHeaders map[string]string, query string, vars map[string]interface{}) (string, error) {
	data, err := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": vars,
	})
	if err != nil {
		return "", err
	}

	s := "curl \\\n"
	if accessToken != "" {
		s += fmt.Sprintf("   %s \\\n", shellquote.Join("-H", "Authorization: token "+accessToken))
	}
	for k, v := range additionalHeaders {
		s += fmt.Sprintf("   %s \\\n", shellquote.Join("-H", k+": "+v))
	}
	s += fmt.Sprintf("   %s \\\n", shellquote.Join("-d", string(data)))
	s += fmt.Sprintf("   %s", shellquote.Join(gqlURL(endpoint)))
	return s, nil
}

// apiRequest represents a GraphQL API request.
type apiRequest struct {
	query  string                 // the GraphQL query
	vars   map[string]interface{} // the GraphQL query variables
	result interface{}            // where to store the result
	done   func() error           // a function to invoke for handling the response. If nil, flags like -get-curl are ignored.
	flags  *apiFlags              // the API flags previously created via newAPIFlags

	// If true, errors will not be unpacked.
	//
	// Consider a GraphQL response like:
	//
	// 	{"data": {...}, "errors": ["something went really wrong"]}
	//
	// 'error unpacking' refers to how we will check if there are any `errors`
	// present in the response (if there are, we will report them on the command
	// line separately AND exit with a proper error code), and if there are no
	// errors `result` will contain only the `{...}` object.
	//
	// When true, the entire response object is stored in `result` -- as if you
	// ran the curl query yourself.
	dontUnpackErrors bool
}

// do performs the API request. If a.flags specify something like -get-curl
// then it is handled immediately and a.done is not invoked. Otherwise, once
// the request is finished a.done is invoked to handle the response (which is
// stored in a.result).
func (a *apiRequest) do() error {
	if a.done != nil {
		// Handle the get-curl flag now.
		if *a.flags.getCurl {
			curl, err := curlCmd(cfg.Endpoint, cfg.AccessToken, cfg.AdditionalHeaders, a.query, a.vars)
			if err != nil {
				return err
			}
			fmt.Println(curl)
			return nil
		}
	} else {
		a.done = func() error { return nil }
	}

	// Create the JSON object.
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(map[string]interface{}{
		"query":     a.query,
		"variables": a.vars,
	}); err != nil {
		return err
	}

	// Create the HTTP request.
	req, err := http.NewRequest("POST", gqlURL(cfg.Endpoint), nil)
	if err != nil {
		return err
	}
	if cfg.AccessToken != "" {
		req.Header.Set("Authorization", "token "+cfg.AccessToken)
	}
	if *a.flags.trace {
		req.Header.Set("X-Sourcegraph-Should-Trace", "true")
	}
	for k, v := range cfg.AdditionalHeaders {
		req.Header.Set(k, v)
	}
	req.Body = ioutil.NopCloser(&buf)

	// Perform the request.
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check trace header before we potentially early exit
	if *a.flags.trace {
		log.Printf("x-trace: %s", resp.Header.Get("x-trace"))
	}

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

	// Handle the case of not unpacking errors.
	if a.dontUnpackErrors {
		if err := jsonCopy(a.result, result); err != nil {
			return err
		}
		if err := a.done(); err != nil {
			return err
		}
		if result.Errors != nil {
			return &exitCodeError{error: nil, exitCode: graphqlErrorsExitCode}
		}
		return nil
	}

	// Handle the case of unpacking errors.
	if result.Errors != nil {
		return &exitCodeError{
			error:    fmt.Errorf("GraphQL errors:\n%s", &graphqlError{result.Errors}),
			exitCode: graphqlErrorsExitCode,
		}
	}
	if err := jsonCopy(a.result, result.Data); err != nil {
		return err
	}
	return a.done()
}

// apiFlags represents generic API flags available in all commands that perform
// API requests. e.g. the ability to turn any CLI command into a curl command.
type apiFlags struct {
	getCurl *bool
	trace   *bool
}

// newAPIFlags creates the API flags. It should be invoked once at flag setup
// time.
func newAPIFlags(flagSet *flag.FlagSet) *apiFlags {
	return &apiFlags{
		getCurl: flagSet.Bool("get-curl", false, "Print the curl command for executing this query and exit (WARNING: includes printing your access token!)"),
		trace:   flagSet.Bool("trace", false, "Log the trace ID for requests. See https://docs.sourcegraph.com/admin/observability/tracing"),
	}
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

func nullInt(n int) *int {
	if n == -1 {
		return nil
	}
	return &n
}

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
