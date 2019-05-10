package main

import (
	"errors"
	"flag"
	"fmt"
)

var extsvcCommands commander

func init() {
	usage := `'src extsvc' is a tool that manages external services on a Sourcegraph instance.

Usage:

	src extsvc command [command options]

The commands are:

	list      lists the external services on the Sourcegraph instance
	edit      edits external services on the Sourcegraph instance

Use "src extsvc [command] -h" for more information about a command.
`

	flagSet := flag.NewFlagSet("extsvc", flag.ExitOnError)
	handler := func(args []string) error {
		extsvcCommands.run(flagSet, "src extsvc", usage, args)
		return nil
	}

	// Register the command.
	commands = append(commands, &command{
		flagSet: flagSet,
		aliases: []string{"extsvc", "external-service"},
		handler: handler,
		usageFunc: func() {
			fmt.Println(usage)
		},
	})
}

type externalService struct {
	ID                   string
	Kind                 string
	DisplayName          string
	Config               string
	CreatedAt, UpdatedAt string
}

func lookupExternalService(byID, byName string) (*externalService, error) {
	var result struct {
		ExternalServices struct {
			Nodes []*externalService
		}
	}
	err := (&apiRequest{
		query: externalServicesListQuery,
		vars: map[string]interface{}{
			"first": 99999,
		},
		result: &result,
	}).do()
	if err != nil {
		return nil, err
	}
	for _, svc := range result.ExternalServices.Nodes {
		if byID != "" && svc.ID == byID {
			return svc, nil
		}
		if byName != "" && svc.DisplayName == byName {
			return svc, nil
		}
	}
	return nil, errors.New("no such external service")
}
