package main

import (
	"flag"
	"fmt"
)

func init() {
	usage := `
Examples:

  Get user with username alice:

    	$ src users get -username=alice

`

	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		usernameFlag = flagSet.String("username", "", `Look up user by username. (e.g. "alice")`)
		formatFlag   = flagSet.String("f", "{{.|json}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Username}} ({{.DisplayName}})")`)
		apiFlags     = newAPIFlags(flagSet)
	)

	handler := func(args []string) error {
		flagSet.Parse(args)

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := `query User(
  $username: String!,
) {
  user(
    username: $username
  ) {
    ...UserFields
  }
}` + userFragment

		var result struct {
			User *User
		}
		return (&apiRequest{
			query: query,
			vars: map[string]interface{}{
				"username": *usernameFlag,
			},
			result: &result,
			done: func() error {
				return execTemplate(tmpl, result.User)
			},
			flags: apiFlags,
		}).do()
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
