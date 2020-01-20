package main

import (
	"flag"
	"fmt"
	"time"
)

var reposCommands commander

func init() {
	usage := `'src repos' is a tool that manages repositories on a Sourcegraph instance.

Usage:

	src repos command [command options]

The commands are:

	get        gets a repository
	list       lists repositories
	enable     enables repositories
	disable    disables repositories
	delete 	   deletes repositories

Use "src repos [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("repos", flag.ExitOnError)
	handler := func(args []string) error {
		reposCommands.run(flagSet, "src repos", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"repo"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

const repositoryFragment = `
fragment RepositoryFields on Repository {
	id
	name
	url
	description
	language
	createdAt
	updatedAt
	externalRepository {
		id
		serviceType
		serviceID
	}
	defaultBranch {
		name
		displayName
	}
	viewerCanAdminister
}
`

type Repository struct {
	ID                  string             `json:"id"`
	Name                string             `json:"name"`
	URL                 string             `json:"url"`
	Description         string             `json:"description"`
	Language            string             `json:"language"`
	CreatedAt           time.Time          `json:"createdAt"`
	UpdatedAt           *time.Time         `json:"updatedAt"`
	ExternalRepository  ExternalRepository `json:"externalRepository"`
	DefaultBranch       GitRef             `json:"defaultBranch"`
	ViewerCanAdminister bool               `json:"viewerCanAdminister"`
}

type ExternalRepository struct {
	ID          string `json:"id"`
	ServiceType string `json:"serviceType"`
	ServiceID   string `json:"serviceID"`
}

type GitRef struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}
