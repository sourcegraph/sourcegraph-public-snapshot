package main

import (
	"flag"
	"fmt"
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
		displayNameFlag = flagSet.String("display-name", "", `The new organization's display name.`)
		apiFlags        = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		query := `mutation CreateOrg(
  $name: String!,
  $displayName: String!,
) {
  createOrg(
    name: $name,
    displayName: $displayName,
  ) {
    id
  }
}`

		var result struct {
			CreateOrg Org
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"name":        *nameFlag,
				"displayName": *displayNameFlag,
			},
			result: &result,
			done: func() error {
				fmt.Printf("Organization %q created.\n", *nameFlag)
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	orgsCommands = append(orgsCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
