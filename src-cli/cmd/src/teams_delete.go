package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  Delete the team "engineering":

    	$ src teams delete -name='engineering'

`

	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src teams %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag = flagSet.String("name", "", "The team name")
		apiFlags = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *nameFlag == "" {
			return errors.New("provide a name")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation DeleteTeam(
	$name: String!,
) {
	deleteTeam(
		name: $name,
	) {
		alwaysNil		
	}
}
`

		var result struct {
			DeleteTeam any
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"name": *nameFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		return nil
	}

	// Register the command.
	teamsCommands = append(teamsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
