package sgx

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	accessGroup, err := cli.CLI.AddCommand("access",
		"manage access",
		"The access subcommands manage permissions to access the Sourcegraph instance.",
		&accessCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = accessGroup.AddCommand("grant",
		"grant access to a user",
		"The `src access grant` command grants read/write access to a user.",
		&accessGrantCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = accessGroup.AddCommand("revoke",
		"revoke access from a user",
		"The `src access revoke` command revokes access from a user.",
		&accessRevokeCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = accessGroup.AddCommand("list",
		"list all users with access",
		"The `src access list` command lists all users with access to this server.",
		&accessListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type accessCmd struct{}

func (c *accessCmd) Execute(args []string) error { return nil }

type accessGrantCmd struct {
	Args struct {
		Users []string `value-name:"USERS" description:"user logins"`
	} `positional-args:"yes"`
	Write bool `long:"write" description:"set write permissions on all specified users"`
	Admin bool `long:"admin" description:"set admin permissions on all specified users"`
}

func (c *accessGrantCmd) Execute(args []string) error {
	cl := Client()
	endpointURL := getEndpointURL().String()

	if len(c.Args.Users) == 0 {
		return fmt.Errorf("Must specify at least one user to grant access to (e.g. \"src access grant bill\")")
	}

	for _, login := range c.Args.Users {
		userSpec, err := sourcegraph.ParseUserSpec(login)
		if err != nil {
			return err
		}
		user, err := cl.Users.Get(cliCtx, &userSpec)
		if err != nil {
			fmt.Printf("# fetching user info for login %s failed: %s\n", login, err)
			continue
		}
		if c.Admin {
			fmt.Printf("# granting admin access to user %s (UID %d) on server running at %s... ", user.Login, user.UID, endpointURL)
		} else {
			fmt.Printf("# granting read/write access to user %s (UID %d) on server running at %s... ", user.Login, user.UID, endpointURL)
		}

		permsOpt := &sourcegraph.UserPermissions{
			UID:   user.UID,
			Read:  true,
			Write: (c.Write || c.Admin),
			Admin: c.Admin,
		}
		if _, err := cl.RegisteredClients.SetUserPermissions(cliCtx, permsOpt); err != nil {
			fmt.Println("FAILED")
			fmt.Printf("   ERROR: %v\n", err)
			continue
		} else {
			fmt.Println("SUCCESS")
		}
	}

	return nil
}

type accessRevokeCmd struct {
	Args struct {
		Users []string `value-name:"USERS" description:"user logins"`
	} `positional-args:"yes"`
}

func (c *accessRevokeCmd) Execute(args []string) error {
	cl := Client()
	endpointURL := getEndpointURL().String()

	if len(c.Args.Users) == 0 {
		return fmt.Errorf("Must specify at least one user to revoke access from (e.g. \"src access revoke bill\")")
	}

	for _, login := range c.Args.Users {
		userSpec, err := sourcegraph.ParseUserSpec(login)
		if err != nil {
			return err
		}
		user, err := cl.Users.Get(cliCtx, &userSpec)
		if err != nil {
			fmt.Printf("# fetching user info for login %s failed: %s\n", login, err)
			continue
		}
		fmt.Printf("# revoking all access from user %s (UID %d) on server running at %s... ", user.Login, user.UID, endpointURL)

		permsOpt := &sourcegraph.UserPermissions{
			UID:   user.UID,
			Read:  false,
			Write: false,
			Admin: false,
		}
		if _, err := cl.RegisteredClients.SetUserPermissions(cliCtx, permsOpt); err != nil {
			fmt.Println("FAILED")
			fmt.Printf("   ERROR: %v\n", err)
			continue
		} else {
			fmt.Println("SUCCESS")
		}
	}

	return nil
}

type accessListCmd struct{}

func (c *accessListCmd) Execute(args []string) error {
	cl := Client()
	endpointURL := getEndpointURL().String()

	fmt.Printf("# fetching list of users with access to server running at %s... ", endpointURL)
	userList, err := cl.RegisteredClients.ListUserPermissions(cliCtx, &sourcegraph.RegisteredClientSpec{})
	if err != nil {
		fmt.Println("FAILED")
		fmt.Printf("   ERROR: %v\n", err)
		return nil
	} else {
		fmt.Println("SUCCESS")
	}

	for _, userPerms := range userList.UserPermissions {
		var login string
		user, err := cl.Users.Get(cliCtx, &sourcegraph.UserSpec{UID: userPerms.UID})
		if err != nil {
			fmt.Printf("# fetching login info for UID %v failed: %s\n", userPerms.UID, err)
		} else {
			login = user.Login
		}
		fmt.Printf("# User %s (UID %d): read=%v, write=%v, admin=%v\n", login, userPerms.UID, userPerms.Read, userPerms.Write, userPerms.Admin)
	}

	return nil
}

func accessString(up *sourcegraph.UserPermissions) string {
	switch {
	case up.Admin:
		return "admin"
	case up.Write:
		return "write"
	case up.Read:
		return "read"
	default:
		return "none"
	}
}
