package main

import (
	"flag"
	"fmt"
)

var teamsCommands commander

func init() {
	usage := `'src teams' is a tool that manages teams in a Sourcegraph instance.

Usage:

	src teams command [command options]

The commands are:

	list	lists teams
	create	create a team
	update	update a team
	delete	delete a team
	members	manage team members, use "src teams members [command] -h" for more information.

Use "src teams [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("teams", flag.ExitOnError)
	handler := func(args []string) error {
		teamsCommands.run(flagSet, "src teams", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"team"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

const teamFragment = `
fragment TeamFields on Team {
    id
    name
    displayName
	readonly
}
`

type Team struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Readonly    bool   `json:"readonly"`
}
