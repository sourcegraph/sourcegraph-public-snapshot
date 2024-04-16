package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
Examples:

  Create an organization:

    	$ src orgs create -name=abc-org -display-name='ABC Organization'

`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src orgs %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		nameFlag        = flagSet.String("name", "", `The new organization's name. (required)`)
		displayNameFlag = flagSet.String("display-name", "", `The new organization's display name. Defaults to organization name if unspecified.`)
		apiFlags        = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation CreateOrg(
  $name: String!,
  $displayName: String!,
) {
  createOrganization(
    name: $name,
    displayName: $displayName,
  ) {
    id
  }
}`

		var result struct {
			CreateOrg Org
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"name":        *nameFlag,
			"displayName": *displayNameFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		fmt.Printf("Organization %q created.\n", *nameFlag)
		return nil
	}

	// Register the command.
	orgsCommands = append(orgsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
