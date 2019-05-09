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

func lookupExternalServiceByName(name string) (id string, err error) {
	var result struct {
		ExternalServices struct {
			Nodes []struct {
				DisplayName string
				ID          string
			}
		}
	}
	err = (&apiRequest{
		query: externalServicesListQuery,
		vars: map[string]interface{}{
			"first": 99999,
		},
		result: &result,
	}).do()
	for _, svc := range result.ExternalServices.Nodes {
		if svc.DisplayName == name {
			id = svc.ID
			break
		}
	}
	if id == "" {
		return "", errors.New("no such external service")
	}
	return
}
