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

  Add a tag "foo" to a user:

    	$ src users tag -user-id=$(src users get -f '{{.ID}}' -username=alice) -tag=foo

  Remove a tag "foo" to a user:

    	$ src users tag -user-id=$(src users get -f '{{.ID}}' -username=alice) -remove -tag=foo

Related examples:

  List all users with the "foo" tag:

    	$ src users list -tag=foo

`

	flagSet := flag.NewFlagSet("tag", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		userIDFlag = flagSet.String("user-id", "", `The ID of the user to tag. (required)`)
		tagFlag    = flagSet.String("tag", "", `The tag to set on the user. (required)`)
		removeFlag = flagSet.Bool("remove", false, `Remove the tag. (default: add the tag`)
		apiFlags   = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation SetUserTag(
  $user: ID!,
  $tag: String!,
  $present: Boolean!
) {
  setTag(
    node: $user,
    tag: $tag,
    present: $present
  ) {
    alwaysNil
  }
}`

		_, err := client.NewRequest(query, map[string]interface{}{
			"user":    *userIDFlag,
			"tag":     *tagFlag,
			"present": !*removeFlag,
		}).Do(context.Background(), &struct{}{})
		return err
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
