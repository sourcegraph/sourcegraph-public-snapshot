package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/api"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	usage := `
Examples:

  List extensions:

    	$ src extensions list

  List extensions whose names match the query:

    	$ src extensions list -query='myquery'

  List *all* extensions (may be slow!):

    	$ src extensions list -first='-1'

`

	flagSet := flag.NewFlagSet("list", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src extensions %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		firstFlag  = flagSet.Int("first", 1000, "Returns the first n extensions from the list. (use -1 for unlimited)")
		queryFlag  = flagSet.String("query", "", `Returns extensions whose extension IDs match the query. (e.g. "myextension")`)
		formatFlag = flagSet.String("f", "{{.ExtensionID}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ExtensionID}}: {{.Manifest.Description}} ({{.RemoteURL}})" or "{{.|json}}")`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `query RegistryExtensions(
  $first: Int,
  $query: String,
) {
  extensionRegistry {
    extensions(
      first: $first,
      query: $query,
    ) {
      nodes {
        ...RegistryExtensionFields
      }
      error
    }
  }
}` + registryExtensionFragment

		var result struct {
			ExtensionRegistry struct {
				Extensions struct {
					Nodes []Extension
					Error string
				}
			}
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"first": api.NullInt(*firstFlag),
			"query": api.NullString(*queryFlag),
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		if result.ExtensionRegistry.Extensions.Error != "" {
			return errors.Newf("%s", result.ExtensionRegistry.Extensions.Error)
		}

		for _, extension := range result.ExtensionRegistry.Extensions.Nodes {
			if err := execTemplate(tmpl, extension); err != nil {
				return err
			}
		}
		return nil
	}

	// Register the command.
	extensionsCommands = append(extensionsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
