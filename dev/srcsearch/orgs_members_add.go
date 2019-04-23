package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Add a member (alice) to an organization (abc-org):

    	$ src orgs members add -org-id=$(src org get -f '{{.ID}}' -name=abc-org) -username=alice

`

	flagSet := flag.NewFlagSet("add", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src orgs members %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		orgIDFlag    = flagSet.String("org-id", "", "ID of organization to which to add member. (required)")
		usernameFlag = flagSet.String("username", "", "Username of user to add as member. (required)")
		apiFlags     = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		query := `mutation AddUserToOrganization(
  $organization: ID!,
  $username: String!,
) {
  addUserToOrganization(
    organization: $organization,
    username: $username,
  ) {
    alwaysNil
  }
}`

		var result struct {
			AddUserToOrganization struct{}
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"organization": *orgIDFlag,
				"username":     *usernameFlag,
			},
			result: &result,
			done: func() error {
				fmt.Printf("User %q added as member to organization with ID %q.\n", *usernameFlag, *orgIDFlag)
				return nil
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	orgsMembersCommands = append(orgsMembersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
