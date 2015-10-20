package sgx

import (
	"fmt"
	"io/ioutil"
	"log"

	"golang.org/x/crypto/ssh"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
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
		"Create a new user account.",
		&usersCreateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	createC.Aliases = []string{"add"}

	listC, err := usersGroup.AddCommand("list",
		"list users",
		"List users.",
		&usersListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	listC.Aliases = []string{"ls"}

	_, err = usersGroup.AddCommand("get",
		"get a user",
		"Show a user's information.",
		&usersGetCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = usersGroup.AddCommand("add-key",
		"add an ssh public key",
		"Add an ssh public key for a user.",
		&usersAddKeyCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	_, err = usersGroup.AddCommand("delete-key",
		"delete the ssh public key",
		"Delete the ssh public key for a user.",
		&usersDeleteKeyCmd{},
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

type usersAddKeyCmd struct {
	Args struct {
		PublicKeyPath string `name:"PublicKeyPath" description:"path to ssh public key"`
	} `positional-args:"yes" required:"yes"`
}

func (c *usersAddKeyCmd) Execute(args []string) error {
	cl := Client()

	// Get the SSH public key.
	keyBytes, err := ioutil.ReadFile(c.Args.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read public SSH key: %v", err)
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse public SSH key: %v\n\nAre you sure you provided a public SSH key?", err)
	}

	// Get user info for output message.
	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		return fmt.Errorf("Error verifying auth credentials: %s.", err)
	}
	user, err := cl.Users.Get(cliCtx, &sourcegraph.UserSpec{UID: authInfo.UID})
	if err != nil {
		return fmt.Errorf("Error getting user with UID %d: %s.", authInfo.UID, err)
	}

	// Add key.
	_, err = cl.UserKeys.AddKey(cliCtx, &sourcegraph.SSHPublicKey{Key: key.Marshal()})
	if err != nil {
		return err
	}

	log.Printf("# Added ssh public key %v for user %q", c.Args.PublicKeyPath, user.Login)
	return nil
}

type usersDeleteKeyCmd struct{}

func (c *usersDeleteKeyCmd) Execute(args []string) error {
	cl := Client()

	// Get user info for output message.
	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		return fmt.Errorf("Error verifying auth credentials: %s.", err)
	}
	user, err := cl.Users.Get(cliCtx, &sourcegraph.UserSpec{UID: authInfo.UID})
	if err != nil {
		return fmt.Errorf("Error getting user with UID %d: %s.", authInfo.UID, err)
	}

	// Delete key.
	_, err = cl.UserKeys.DeleteKey(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	log.Printf("# Deleted ssh public key for user %q\n", user.Login)
	return nil
}
