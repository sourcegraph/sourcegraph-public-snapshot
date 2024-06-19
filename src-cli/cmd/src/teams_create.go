package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
	"github.com/sourcegraph/src-cli/internal/cmderrors"
)

func init() {
	usage := `
Examples:

  Create a team "engineering":

    	$ src teams create -name='engineering' [-display-name='Engineering Team'] [-parent-team='engineering-leadership'] [-read-only]

`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src teams %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag        = flagSet.String("name", "", "The team name")
		displayNameFlag = flagSet.String("display-name", "", "Optional additional display name for a more human-readable UI")
		parentTeamFlag  = flagSet.String("parent-team", "", "Optional name or ID of the parent team")
		readonlyFlag    = flagSet.Bool("read-only", false, "Optionally create the team as read-only marking it as externally managed in this UI")
		apiFlags        = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *nameFlag == "" {
			return errors.New("provide a name")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation CreateTeam(
	$name: String!,
	$displayName: String,
	$parentTeam: String,
	$readonly: Boolean
) {
	createTeam(
		name: $name,
		displayName: $displayName,
		parentTeamName: $parentTeam,
		readonly: $readonly,
	) {
		...TeamFields		
	}
}
` + teamFragment

		var result struct {
			CreateTeam Team
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"name":        *nameFlag,
			"displayName": api.NullString(*displayNameFlag),
			"parentTeam":  api.NullString(*parentTeamFlag),
			"readonly":    *readonlyFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			var gqlErr api.GraphQlErrors
			if errors.As(err, &gqlErr) {
				for _, e := range gqlErr {
					if strings.Contains(e.Error(), "team name is already taken") {
						return cmderrors.ExitCode(3, err)
					}
				}
			}
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
