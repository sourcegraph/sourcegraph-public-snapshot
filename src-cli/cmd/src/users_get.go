package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/sourcegraph/sourcegraph/lib/errors"

	"github.com/sourcegraph/src-cli/internal/api"
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
		emailFlag    = flagSet.String("email", "", `Look up user by email. (e.g. "alice@sourcegraph.com")`)
		formatFlag   = flagSet.String("f", "{{.|json}}", `Format for the output, using the syntax of Go package text/template. (e.g. "{{.ID}}: {{.Username}} ({{.DisplayName}})")`)
		apiFlags     = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		if usernameFlag != nil && *usernameFlag != "" && emailFlag != nil && *emailFlag != "" {
			return errors.New("cannot specify both email and username")
		}

		tmpl, err := parseTemplate(*formatFlag)
		if err != nil {
			return err
		}

		query := `query User(
  $username: String,
  $email: String,
) {
  user(
    username: $username,
    email: $email,
  ) {
    ...UserFields
  }
}` + userFragment

		var result struct {
			User *User
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"username": api.NullString(*usernameFlag),
			"email":    api.NullString(*emailFlag),
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		return execTemplate(tmpl, result.User)
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
