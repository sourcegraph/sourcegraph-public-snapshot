package sgx

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
)

func init() {
	usersGroup, err := cli.CLI.AddCommand("user",
		"manage users",
		"The user subcommands manage registered users.",
		&usersCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	usersGroup.Aliases = []string{"users", "u"}

	createC, err := usersGroup.AddCommand("create",
		"create a user account",
		"The `sgx users create` command creates a new user account.",
		&usersCreateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	createC.Aliases = []string{"add"}

	listC, err := usersGroup.AddCommand("list",
		"list users",
		"The `sgx user list` command lists users.",
		&usersListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	listC.Aliases = []string{"ls"}

	_, err = usersGroup.AddCommand("get",
		"get a user",
		"The `sgx user get` command shows a user's information.",
		&usersGetCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type usersCmd struct{}

func (c *usersCmd) Execute(args []string) error { return nil }

type usersCreateCmd struct {
	Args struct {
		Login string `name:"LOGIN" description:"login of the user to add"`
	} `positional-args:"yes" required:"yes"`
}

func (c *usersCreateCmd) Execute(args []string) error {
	cl := Client()

	user, err := cl.Accounts.Create(cliCtx, &sourcegraph.NewAccount{Login: c.Args.Login})
	if err != nil {
		return err
	}

	log.Printf("# Created user %q with UID %d", user.Login, user.UID)
	return nil
}

type usersListCmd struct {
	Args struct {
		Query string `name:"QUERY" description:"search query"`
	} `positional-args:"yes"`
}

func (c *usersListCmd) Execute(args []string) error {
	cl := Client()

	for page := 1; ; page++ {
		users, err := cl.Users.List(cliCtx, &sourcegraph.UsersListOptions{
			Query:       c.Args.Query,
			ListOptions: sourcegraph.ListOptions{Page: int32(page)},
		})

		if err != nil {
			return err
		}
		if len(users.Users) == 0 {
			break
		}
		for _, user := range users.Users {
			fmt.Println(user)
		}
	}
	return nil
}

type usersGetCmd struct {
	Args struct {
		User string `name:"User" description:"user login (or login@domain)"`
	} `positional-args:"yes"`
}

func (c *usersGetCmd) Execute(args []string) error {
	cl := Client()

	userSpec, err := sourcegraph.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(cliCtx, &userSpec)
	if err != nil {
		return err
	}
	fmt.Println(user)
	fmt.Println()

	fmt.Println("# Emails")
	userSpec2 := user.Spec()
	emails, err := cl.Users.ListEmails(cliCtx, &userSpec2)
	if err != nil {
		return err
	}
	if len(emails.EmailAddrs) == 0 {
		fmt.Println("# (no emails found for user)")
	}
	for _, email := range emails.EmailAddrs {
		fmt.Println(email)
	}

	return nil
}
