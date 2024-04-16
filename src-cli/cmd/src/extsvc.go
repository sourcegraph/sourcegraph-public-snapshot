package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
)

var extsvcCommands commander

func init() {
	usage := `'src extsvc' is a tool that manages external services on a Sourcegraph instance.

Usage:

	src extsvc command [command options]

The commands are:

	list      lists the external services on the Sourcegraph instance
	edit      edits external services on the Sourcegraph instance
	add       add an external service on the Sourcegraph instance

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

var errServiceNotFound = errors.New("no such external service")

func lookupExternalService(ctx context.Context, client api.Client, byID, byName string) (*externalService, error) {
	var result struct {
		ExternalServices struct {
			Nodes []*externalService
		}
	}
	if ok, err := client.NewRequest(externalServicesListQuery, map[string]interface{}{
		"first": 99999,
	}).Do(ctx, &result); err != nil || !ok {
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
	return nil, errServiceNotFound
}
