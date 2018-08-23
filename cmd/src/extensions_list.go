package main

import (
	"flag"
	"fmt"
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
		formatFlag = flagSet.String("f", "{{.ExtensionID}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ExtensionID}}: {{.Manifest.Title}} ({{.RemoteURL}})" or "{{.|json}}")`)
		apiFlags   = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

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
    }
  }
}` + registryExtensionFragment

		var result struct {
			ExtensionRegistry struct {
				Extensions struct {
					Nodes []Extension
				}
			}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"first": nullInt(*firstFlag),
				"query": nullString(*queryFlag),
			},
			result: &result,
			done: func() error {
				for _, extension := range result.ExtensionRegistry.Extensions.Nodes {
					if err := execTemplate(tmpl, extension); err != nil {
						return err
					}
				}
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	extensionsCommands = append(extensionsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
