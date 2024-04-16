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

  Look up a repository by name:

    	$ src repos get -name=github.com/sourcegraph/src-cli

`

	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src repos %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag   = flagSet.String("name", "", "The name of the repository. (required)")
		formatFlag = flagSet.String("f", "{{.ID}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Name}}") or "{{.|json}}")`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := `query Repository(
  $name: String!,
) {
  repository(
    name: $name
  ) {
    ...RepositoryFields
  }
}
` + repositoryFragment

		var result struct {
			Repository Repository
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"name": *nameFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		return execTemplate(tmpl, result.Repository)
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
