package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"

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
		apiFlags  = api.NewFlags(flagSet)
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
				return cmderrors.Usage("expected query to be piped into 'src api' or -query flag to be specified")
			}
			data, err := io.ReadAll(os.Stdin)
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
				return cmderrors.Usagef("parsing argument %q expected 'variable=value' syntax (missing equals)", arg)
			}
			key := arg[:idx]
			value := arg[idx+1:]
			vars[key] = value
		}

		// Perform the request.
		var result interface{}
		if ok, err := cfg.apiClient(apiFlags, flagSet.Output()).NewRequest(query, vars).DoRaw(context.Background(), &result); err != nil || !ok {
			return err
		}

		// Print the formatted JSON.
		f, err := marshalIndent(result)
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
