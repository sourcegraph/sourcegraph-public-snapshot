package sgx

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"

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

	_, err = userGroup.AddCommand("invite",
		"send an invite to access this server",
		"Send a user invite.",
		&userInviteCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = userGroup.AddCommand("rm-invite",
		"remove an existing, not-yet-accepted invite",
		"Remove an existing, not-yet-accepted invite.",
		&userRmInviteCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

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

	_, err = userGroup.AddCommand("update",
		"update a user",
		"Update a user's information.",
		&userUpdateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = userGroup.AddCommand("reset-password",
		"generate a password reset link for user",
		"Generate a password reset link for user.",
		&userResetPasswordCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = userGroup.AddCommand("delete",
		"delete a user account",
		"Delete a user account.",
		&userDeleteCmd{},
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
	_, err = userKeysGroup.AddCommand("list",
		"list SSH public keys",
		"List the SSH public keys for a user.",
		&userKeysListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type userCmd struct{}

func (c *userCmd) Execute(args []string) error { return nil }

type userCreateCmd struct {
	Args struct {
		Login    string `name:"LOGIN" description:"login of the user to add"`
		Password string `name:"PASSWORD" description:"password of the user to add"`
	} `positional-args:"yes" required:"yes"`
}

func (c *userCreateCmd) Execute(args []string) error {
	cl := cli.Client()

	user, err := cl.Accounts.Create(cli.Ctx, &sourcegraph.NewAccount{Login: c.Args.Login, Password: c.Args.Password})
	if err != nil {
		return err
	}

	log.Printf("# Created user %q with UID %d", user.Login, user.UID)
	return nil
}

type userInviteCmd struct {
	Args struct {
		Emails []string `value-name:"EMAILS" description:"user emails"`
	} `positional-args:"yes"`
	Write bool `long:"write" description:"set write permissions on all specified users"`
	Admin bool `long:"admin" description:"set admin permissions on all specified users"`
}

func (c *userInviteCmd) Execute(args []string) error {
	cl := cli.Client()

	if len(c.Args.Emails) == 0 {
		return fmt.Errorf(`Must specify at least one email to invite (e.g. "src user invite EMAIL")`)
	}

	var success bool
	for _, email := range c.Args.Emails {
		pendingInvite, err := cl.Accounts.Invite(cli.Ctx, &sourcegraph.AccountInvite{
			Email: email,
			Write: c.Write || c.Admin,
			Admin: c.Admin,
		})
		if err != nil {
			fmt.Printf("FAIL %s: %v\n", email, err)
			continue
		}
		status := fmt.Sprintf("  OK %s: %s", email, pendingInvite.Link)
		if pendingInvite.EmailSent {
			status += " (email sent)"
		}
		fmt.Println(status)
		success = true
	}

	if success {
		fmt.Println("# Share the above link with the user(s) to accept the invite")
	}

	return nil
}

type userRmInviteCmd struct {
	Args struct {
		Emails []string `value-name:"EMAILS" description:"user emails"`
	} `positional-args:"yes"`
}

func (c *userRmInviteCmd) Execute(args []string) error {
	cl := cli.Client()

	if len(c.Args.Emails) == 0 {
		return fmt.Errorf(`Must specify at least one email (e.g. "src user rm-invite EMAIL")`)
	}

	for _, email := range c.Args.Emails {
		_, err := cl.Accounts.DeleteInvite(cli.Ctx, &sourcegraph.InviteSpec{Email: email})
		if err != nil {
			return fmt.Errorf("deleting invite for %s: %s", email, err)
		}
		log.Printf("%s: deleted invite", email)
	}

	return nil
}

type userListCmd struct {
	Args struct {
		Query string `name:"QUERY" description:"search query"`
	} `positional-args:"yes"`
}

func (c *userListCmd) Execute(args []string) error {
	cl := cli.Client()

	for page := 1; ; page++ {
		users, err := cl.Users.List(cli.Ctx, &sourcegraph.UsersListOptions{
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
	cl := cli.Client()

	userSpec, err := sourcegraph.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(cli.Ctx, &userSpec)
	if err != nil {
		return err
	}
	fmt.Println(user)
	fmt.Println()

	fmt.Println("# Emails")
	userSpec2 := user.Spec()
	emails, err := cl.Users.ListEmails(cli.Ctx, &userSpec2)
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

type userUpdateCmd struct {
	Access string `long:"access" description:"set access level for user (read|write|admin)"`
	Args   struct {
		User string `name:"User" description:"user login (or login@domain)"`
	} `positional-args:"yes"`
}

func (c *userUpdateCmd) Execute(args []string) error {
	cl := cli.Client()

	userSpec, err := sourcegraph.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(cli.Ctx, &userSpec)
	if err != nil {
		return err
	}

	if c.Access != "" {
		switch c.Access {
		case "read":
			user.Write = false
			user.Admin = false
		case "write":
			user.Write = true
			user.Admin = false
		case "admin":
			user.Write = true
			user.Admin = true
		default:
			return fmt.Errorf("access level not recognized (should be one of read/write/admin): %s", c.Access)
		}

		if _, err := cl.Accounts.Update(cli.Ctx, user); err != nil {
			return err
		}
		fmt.Printf("# updated access level for user %s to %s\n", user.Login, c.Access)
	}

	return nil
}

type userResetPasswordCmd struct {
	Email string `long:"email" short:"e" description:"email address associated with account"`
	Login string `long:"login" short:"l" description:"login name of the user account"`
}

func (c *userResetPasswordCmd) Execute(args []string) error {
	cl := cli.Client()

	person := &sourcegraph.PersonSpec{}
	var identifier string
	if c.Email != "" {
		person.Email = c.Email
		identifier = c.Email
	} else if c.Login != "" {
		person.Login = c.Login
		identifier = c.Login
	} else {
		return fmt.Errorf("need to specify either email or login of the user account")
	}

	pendingReset, err := cl.Accounts.RequestPasswordReset(cli.Ctx, person)
	if err != nil {
		return err
	}

	var status string
	if pendingReset.EmailSent {
		status = "email sent"
	} else {
		status = "email not sent"
	}
	fmt.Printf("# Password reset link generated for %v (%s)\n", identifier, status)

	if pendingReset.Link != "" {
		fmt.Println("# Share the link below with the user to set a new password.")
		fmt.Printf("login: %s, reset link: %s\n", pendingReset.Login, pendingReset.Link)
	} else {
		fmt.Println("# Link not available: need to be authenticated as an admin user.")
	}

	return nil
}

type userDeleteCmd struct {
	Email string `long:"email" short:"e" description:"email address associated with user account"`
	Login string `long:"login" short:"l" description:"login name of the user account"`
	UID   int32  `long:"uid" short:"i" description:"UID of the user account"`
}

func (c *userDeleteCmd) Execute(args []string) error {
	cl := cli.Client()

	authInfo, err := cl.Auth.Identify(cli.Ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	if !authInfo.Admin {
		return fmt.Errorf("# Permission denied: need admin access to complete this operation.")
	}

	person := &sourcegraph.PersonSpec{}
	var identifier string
	if c.Email != "" {
		person.Email = c.Email
		identifier = c.Email
	} else if c.Login != "" {
		person.Login = c.Login
		identifier = c.Login
	} else if c.UID != 0 {
		person.UID = c.UID
		identifier = fmt.Sprintf("UID %d", c.UID)
	} else {
		return fmt.Errorf("need to specify email, login or UID of the user account")
	}

	_, err = cl.Accounts.Delete(cli.Ctx, person)
	if err != nil {
		return err
	}

	fmt.Printf("# User %q deleted.\n", identifier)

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
	cl := cli.Client()

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
	authInfo, err := cl.Auth.Identify(cli.Ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	// Add key.
	_, err = cl.UserKeys.AddKey(cli.Ctx, &sourcegraph.SSHPublicKey{Key: key.Marshal()})
	if err != nil {
		return err
	}

	log.Printf("# Added SSH public key %v for user %q", c.Args.PublicKeyPath, authInfo.Login)
	return nil
}

type userKeysDeleteCmd struct {
	ID   string `long:"id" description:"ID of the key to delete"`
	Name string `long:"name" description:"name of the key to delete"`
}

func (c *userKeysDeleteCmd) Execute(args []string) error {
	cl := cli.Client()

	if c.ID == "" && c.Name == "" {
		log.Fatal("Must specify either --id or --name of key to delete.")
	}
	id, _ := strconv.ParseUint(c.ID, 10, 64)

	// Get user info for output message.
	authInfo, err := cl.Auth.Identify(cli.Ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	// Delete key.
	_, err = cl.UserKeys.DeleteKey(cli.Ctx, &sourcegraph.SSHPublicKey{
		ID:   id,
		Name: c.Name,
	})
	if err != nil {
		return err
	}

	log.Printf("# Deleted SSH public key for user %q\n", authInfo.Login)
	return nil
}

type userKeysListCmd struct{}

func (c *userKeysListCmd) Execute(args []string) error {
	cl := cli.Client()

	// Get user info for output message.
	authInfo, err := cl.Auth.Identify(cli.Ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	// List keys.
	keys, err := cl.UserKeys.ListKeys(cli.Ctx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	if len(keys.SSHKeys) == 0 {
		log.Printf("User %q has no SSH public keys.\n", authInfo.Login)
	} else {
		log.Printf("SSH public keys for user %q:\n", authInfo.Login)
		for _, k := range keys.SSHKeys {
			log.Printf("%d. %q\n", k.ID, k.Name)
		}
	}
	return nil
}
