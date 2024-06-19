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
		apiFlags     = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

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
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"organization": *orgIDFlag,
			"username":     *usernameFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		fmt.Printf("User %q added as member to organization with ID %q.\n", *usernameFlag, *orgIDFlag)
		return nil
	}

	// Register the command.
	orgsMembersCommands = append(orgsMembersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
