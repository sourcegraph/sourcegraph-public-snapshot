package main

import (
	"flag"
	"fmt"
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
		apiFlags   = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

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
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"name": *nameFlag,
			},
			result: &result,
			done: func() error {
				if err := execTemplate(tmpl, result.Repository); err != nil {
					return err
				}
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	reposCommands = append(reposCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
