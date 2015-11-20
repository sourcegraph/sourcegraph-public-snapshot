package sgx

import (
	"fmt"
	"io/ioutil"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"golang.org/x/crypto/ssh"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	userGroup, err := cli.CLI.AddCommand("user",
		"manage users",
		"Manage registered users.",
		&userCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	userGroup.Aliases = []string{"users", "u"}

	createCmd, err := userGroup.AddCommand("create",
		"create a user account",
		"Create a new user account.",
		&userCreateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	createCmd.Aliases = []string{"add"}

	listCmd, err := userGroup.AddCommand("list",
		"list users",
		"List users.",
		&userListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	listCmd.Aliases = []string{"ls"}

	_, err = userGroup.AddCommand("get",
		"get a user",
		"Show a user's information.",
		&userGetCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	userKeysGroup, err := userGroup.AddCommand("keys",
		"manage user's SSH public keys",
		"Manage user's SSH public keys.",
		&userKeysCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	_, err = userKeysGroup.AddCommand("add",
		"add an SSH public key",
		"Add an SSH public key for a user.",
		&userKeysAddCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	_, err = userKeysGroup.AddCommand("delete",
		"delete the SSH public key",
		"Delete the SSH public key for a user.",
		&userKeysDeleteCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type userCmd struct{}

func (c *userCmd) Execute(args []string) error { return nil }

type userCreateCmd struct {
	Args struct {
		Login string `name:"LOGIN" description:"login of the user to add"`
	} `positional-args:"yes" required:"yes"`
}

func (c *userCreateCmd) Execute(args []string) error {
	cl := Client()

	user, err := cl.Accounts.Create(cliCtx, &sourcegraph.NewAccount{Login: c.Args.Login})
	if err != nil {
		return err
	}

	log.Printf("# Created user %q with UID %d", user.Login, user.UID)
	return nil
}

type userListCmd struct {
	Args struct {
		Query string `name:"QUERY" description:"search query"`
	} `positional-args:"yes"`
}

func (c *userListCmd) Execute(args []string) error {
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

type userGetCmd struct {
	Args struct {
		User string `name:"User" description:"user login (or login@domain)"`
	} `positional-args:"yes"`
}

func (c *userGetCmd) Execute(args []string) error {
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
		if grpc.Code(err) == codes.PermissionDenied {
			fmt.Println("# (permission denied)")
		} else {
			return err
		}
	} else {
		if len(emails.EmailAddrs) == 0 {
			fmt.Println("# (no emails found for user)")
		}
		for _, email := range emails.EmailAddrs {
			fmt.Println(email)
		}
	}

	return nil
}

type userKeysCmd struct{}

func (*userKeysCmd) Execute(args []string) error { return nil }

type userKeysAddCmd struct {
	Args struct {
		PublicKeyPath string `name:"PublicKeyPath" description:"path to SSH public key"`
	} `positional-args:"yes" required:"yes"`
}

func (c *userKeysAddCmd) Execute(args []string) error {
	cl := Client()

	// Get the SSH public key.
	keyBytes, err := ioutil.ReadFile(c.Args.PublicKeyPath)
	if err != nil {
		return fmt.Errorf("failed to read SSH public key: %v", err)
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey(keyBytes)
	if err != nil {
		return fmt.Errorf("failed to parse SSH public key: %v\n\nAre you sure you provided a SSH public key?", err)
	}

	// Get user info for output message.
	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	// Add key.
	_, err = cl.UserKeys.AddKey(cliCtx, &sourcegraph.SSHPublicKey{Key: key.Marshal()})
	if err != nil {
		return err
	}

	log.Printf("# Added SSH public key %v for user %q", c.Args.PublicKeyPath, authInfo.Login)
	return nil
}

type userKeysDeleteCmd struct{}

func (c *userKeysDeleteCmd) Execute(args []string) error {
	cl := Client()

	// Get user info for output message.
	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	// Delete key.
	_, err = cl.UserKeys.DeleteKey(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	log.Printf("# Deleted SSH public key for user %q\n", authInfo.Login)
	return nil
}
