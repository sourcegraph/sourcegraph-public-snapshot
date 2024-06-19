package main

import (
	"flag"
	"fmt"
)

var usersCommands commander

func init() {
	usage := `'src users' is a tool that manages users on a Sourcegraph instance.

Usage:

	src users command [command options]

The commands are:

	list       lists users
	get        gets a user
	create     creates a user account
	delete     deletes a user account
	prune      deletes inactive users
	tag        add/remove a tag on a user

Use "src users [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("users", flag.ExitOnError)
	handler := func(args []string) error {
		usersCommands.run(flagSet, "src users", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"user"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

const userFragment = `
fragment UserFields on User {
    id
    username
    displayName
    siteAdmin
    organizations {
		nodes {
        	id
        	name
        	displayName
		}
    }
    emails {
        email
		verified
    }
    usageStatistics {
        lastActiveTime
        lastActiveCodeHostIntegrationTime
    }
    url
}
`

type User struct {
	ID            string
	Username      string
	DisplayName   string
	SiteAdmin     bool
	Organizations struct {
		Nodes []Org
	}
	Emails          []UserEmail
	UsageStatistics UserUsageStatistics
	URL             string
}

type UserEmail struct {
	Email    string
	Verified bool
}

type UserUsageStatistics struct {
	LastActiveTime                    string
	LastActiveCodeHostIntegrationTime string
}

type SiteUser struct {
	ID           string
	Username     string
	Email        string
	SiteAdmin    bool
	LastActiveAt string
	DeletedAt    string
}
