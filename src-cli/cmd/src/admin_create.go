package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/sourcegraph/src-cli/internal/users"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func init() {
	usage := `
Examples:

	Create an initial admin user on a new Sourcegraph deployment:

		$ src admin create -url https://your-sourcegraph-url -username admin -email admin@yourcompany.com -with-token

	Create an initial admin user on a new Sourcegraph deployment using '-password' flag. 
	WARNING: for security purposes we strongly recommend using the SRC_ADMIN_PASS environment variable when possible.

		$ src admin create -url https://your-sourcegraph-url -username admin -email admin@yourcompany.com -password p@55w0rd -with-token

Environmental variables

	SRC_ADMIN_PASS		The new admin user's password
`

	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}

	var (
		urlFlag      = flagSet.String("url", "", "The base URL for the Sourcegraph instance.")
		usernameFlag = flagSet.String("username", "", "The new admin user's username.")
		emailFlag    = flagSet.String("email", "", "The new admin user's email address.")
		passwordFlag = flagSet.String("password", "", "The new admin user's password.")
		tokenFlag    = flagSet.Bool("with-token", false, "Optionally create and output an admin access token.")
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		ok, _, err := users.NeedsSiteInit(*urlFlag)
		if err != nil {
			return err
		}
		if !ok {
			return errors.New("failed to create admin, site already initialized")
		}

		envAdminPass := os.Getenv("SRC_ADMIN_PASS")

		var client *users.Client

		switch {
		case envAdminPass != "" && *passwordFlag == "":
			client, err = users.SiteAdminInit(*urlFlag, *emailFlag, *usernameFlag, envAdminPass)
			if err != nil {
				return err
			}
		case envAdminPass == "" && *passwordFlag != "":
			client, err = users.SiteAdminInit(*urlFlag, *emailFlag, *usernameFlag, *passwordFlag)
			if err != nil {
				return err
			}
		case envAdminPass != "" && *passwordFlag != "":
			return errors.New("failed to read admin password: environment variable and -password flag both set")
		case envAdminPass == "" && *passwordFlag == "":
			return errors.New("failed to read admin password from 'SRC_ADMIN_PASS' environment variable or -password flag")
		}

		if *tokenFlag {
			token, err := client.CreateAccessToken("", []string{"user:all", "site-admin:sudo"}, "src-cli")
			if err != nil {
				return err
			}

			fmt.Println(token)
		}

		return nil
	}

	adminCommands = append(adminCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
