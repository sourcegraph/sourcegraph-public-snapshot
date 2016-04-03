package sgx

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/client"
	"sourcegraph.com/sqs/pbtypes"
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
}

type userCmd struct{}

func (c *userCmd) Execute(args []string) error { return nil }

type userCreateCmd struct {
	Args struct {
		Login    string `name:"LOGIN" description:"login of the user to add"`
		Password string `name:"PASSWORD" description:"password of the user to add"`
		Email    string `name:"EMAIL" description:"email of the user to add"`
	} `positional-args:"yes" required:"yes"`
}

func (c *userCreateCmd) Execute(args []string) error {
	cl := client.Client()

	user, err := cl.Accounts.Create(client.Ctx, &sourcegraph.NewAccount{Login: c.Args.Login, Password: c.Args.Password, Email: c.Args.Email})
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
	cl := client.Client()

	for page := 1; ; page++ {
		users, err := cl.Users.List(client.Ctx, &sourcegraph.UsersListOptions{
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
	cl := client.Client()

	userSpec, err := sourcegraph.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(client.Ctx, &userSpec)
	if err != nil {
		return err
	}
	fmt.Println(user)
	fmt.Println()

	fmt.Println("# Emails")
	userSpec2 := user.Spec()
	emails, err := cl.Users.ListEmails(client.Ctx, &userSpec2)
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
	} `positional-args:"yes" required:"yes"`
}

func (c *userUpdateCmd) Execute(args []string) error {
	cl := client.Client()

	userSpec, err := sourcegraph.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(client.Ctx, &userSpec)
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

		if _, err := cl.Accounts.Update(client.Ctx, user); err != nil {
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
	cl := client.Client()

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

	pendingReset, err := cl.Accounts.RequestPasswordReset(client.Ctx, person)
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
	cl := client.Client()

	authInfo, err := cl.Auth.Identify(client.Ctx, &pbtypes.Void{})
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

	_, err = cl.Accounts.Delete(client.Ctx, person)
	if err != nil {
		return err
	}

	fmt.Printf("# User %q deleted.\n", identifier)

	return nil
}
