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

  Delete a user account by ID:

    	$ src users delete -id=VXNlcjox

  Delete a user account by username:

    	$ src users delete -id=$(src users get -f='{{.ID}}' -username=alice)

  Delete all user accounts that match the query:

    	$ src users list -f='{{.ID}}' -query=alice | xargs -n 1 -I USERID src users delete -id=USERID

`

	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		userIDFlag = flagSet.String("id", "", `The ID of the user to delete.`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation DeleteUser(
  $user: ID!
) {
  deleteUser(
    user: $user
  ) {
    alwaysNil
  }
}`

		var result struct {
			DeleteUser struct{}
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"user": *userIDFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		fmt.Printf("User with ID %q deleted.\n", *userIDFlag)
		return nil
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
