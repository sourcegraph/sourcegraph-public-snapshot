package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  List external service configurations on the Sourcegraph instance:

    	$ src extsvc list

  List external service configurations and choose output format:

    	$ src extsvc list -f '{{.ID}}'

`

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extsvc %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		firstFlag  = flagSet.Int("first", -1, "Return only the first n external services. (use -1 for unlimited)")
		formatFlag = flagSet.String("f", "", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.|json}}")`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		first := *firstFlag
		if first == -1 {
			first = 9999999 // GraphQL API doesn't support negative for unlimited query
		}

		var formatStr string
		if *formatFlag != "" {
			formatStr = *formatFlag
		} else {
			// Set default here instead of in flagSet.String because it is very long and makes the usage message ugly.
			formatStr = `{{range .Nodes}}ID: {{.id}} | {{padRight .kind 15 " "}} | {{.displayName}}{{"\n"}}{{end}}`
		}
		tmpl, err := parseTemplate(formatStr)
		if err != nil {
			return err
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		queryVars := map[string]interface{}{
			"first": first,
		}
		var result externalServicesListResult
		if ok, err := client.NewRequest(externalServicesListQuery, queryVars).Do(ctx, &result); err != nil || !ok {
			return err
		}
		return execTemplate(tmpl, result.ExternalServices)
	}

	// Register the command.
	extsvcCommands = append(extsvcCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

const externalServicesListQuery = `
	query ($first: Int!) {
		externalServices(first: $first) {
			nodes {
				id
				kind
				displayName
				config
				createdAt
				updatedAt
			}
			totalCount
			pageInfo {
				hasNextPage
			}
		}
	}
`

// Typing here is primarily for ordering of results as we format them (maps are unordered).
type externalServicesListResult struct {
	ExternalServices struct {
		Nodes      []map[string]interface{}
		TotalCount int
		PageInfo   struct {
			HasNextPage bool
		}
	}
}
