package main

import (
	"flag"
	"fmt"
)

var teamMembersCommands commander

func init() {
	usage := `'src teams members' is a tool that manages team membership in a Sourcegraph instance.

Usage:

	src team members command [command options]

The commands are:

	list	lists team members
	add	add team members
	remove	remove team members

Use "src team members [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("members", flag.ExitOnError)
	handler := func(args []string) error {
		teamMembersCommands.run(flagSet, "src teams members", usage, args)
		return nil
	}

	// Register the command.
	teamsCommands = append(teamsCommands, &command{
		flagSet: flagSet,
		aliases: []string{"member"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

const teamMemberFragment = `
fragment TeamMemberFields on TeamMember {
    ... on User {
		id
		username
	}
}
`

type TeamMember struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}
