package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Delete an organization by ID:

    	$ src orgs delete -id=VXNlcjox

  Delete an organization by name:

    	$ src orgs delete -id=$(src orgs get -f='{{.ID}}' -name=abc-org)

  Delete all organizations that match the query

    	$ src orgs list -f='{{.ID}}' -query=abc-org | xargs -n 1 -I ORGID src orgs delete -id=ORGID

`

	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src orgs %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		orgIDFlag = flagSet.String("id", "", `The ID of the organization to delete.`)
		apiFlags  = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		query := `mutation DeleteOrganization(
  $organization: ID!
) {
  deleteOrganization(
    organization: $organization
  ) {
    alwaysNil
  }
}`

		var result struct {
			DeleteOrganization struct{}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"organization": *orgIDFlag,
			},
			result: &result,
			done: func() error {
				fmt.Printf("Organization with ID %q deleted.\n", *orgIDFlag)
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
