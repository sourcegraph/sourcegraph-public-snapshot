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

  Create a user account:

    	$ src users create -username=alice -email=alice@example.com

`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		usernameFlag         = flagSet.String("username", "", `The new user's username. (required)`)
		emailFlag            = flagSet.String("email", "", `The new user's email address. (required)`)
		resetPasswordURLFlag = flagSet.Bool("reset-password-url", false, `Print the reset password URL to manually send to the new user.`)
		apiFlags             = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation CreateUser(
  $username: String!,
  $email: String!,
) {
  createUser(
    username: $username,
    email: $email,
  ) {
    resetPasswordURL
  }
}`

		var result struct {
			CreateUser struct {
				ResetPasswordURL string
			}
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"username": *usernameFlag,
			"email":    *emailFlag,
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		fmt.Printf("User %q created.\n", *usernameFlag)
		if *resetPasswordURLFlag && result.CreateUser.ResetPasswordURL != "" {
			fmt.Println()
			fmt.Printf("\tReset pasword URL: %s\n", result.CreateUser.ResetPasswordURL)
		}
		return nil
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
