package cli

import (
	"fmt"
	"log"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/routevar"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
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

	_, err = userGroup.AddCommand("revoke-external-auth",
		"revoke an external authorization",
		"Revoke and delete an external authorization (e.g., a GitHub OAuth2 authorization).",
		&userRevokeExternalAuthCmd{},
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
	cl := cliClient

	user, err := cl.Accounts.Create(cliContext, &sourcegraph.NewAccount{Login: c.Args.Login, Password: c.Args.Password, Email: c.Args.Email})
	if err != nil {
		return err
	}

	log.Printf("# Created user %q with UID %d", c.Args.Login, user.UID)
	return nil
}

type userListCmd struct {
	AllBetas       string `long:"all-betas" description:"only users participating in all the given betas"`
	RegisteredBeta bool   `long:"registered-beta" description:"filter by users who have registered for beta access"`
	HaveBeta       bool   `long:"have-beta" description:"filter by users who have access to at least one beta"`
	Args           struct {
		Query string `name:"QUERY" description:"search query"`
	} `positional-args:"yes"`
}

func (c *userListCmd) Execute(args []string) error {
	cl := cliClient

	for page := 1; ; page++ {
		users, err := cl.Users.List(cliContext, &sourcegraph.UsersListOptions{
			Query:          c.Args.Query,
			AllBetas:       commaSplit(c.AllBetas),
			RegisteredBeta: c.RegisteredBeta,
			HaveBeta:       c.HaveBeta,
			ListOptions:    sourcegraph.ListOptions{Page: int32(page)},
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
	cl := cliClient

	userSpec, err := routevar.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(cliContext, &userSpec)
	if err != nil {
		return err
	}
	fmt.Println(user)
	fmt.Println()

	fmt.Println("# Emails")
	userSpec2 := user.Spec()
	emails, err := cl.Users.ListEmails(cliContext, &userSpec2)
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

func quotedComma(strs []string) string {
	for i, s := range strs {
		strs[i] = fmt.Sprintf("%q", s)
	}
	return strings.Join(strs, ",")
}

type userUpdateCmd struct {
	Access   string `long:"access" description:"set access level for user (read|write|admin)"`
	SetBetas string `long:"set-betas" description:"set given betas for user"`
	AddBetas string `long:"add-betas" description:"add given betas to user"`
	RmBetas  string `long:"rm-betas" description:"remove given betas from user"`
	Args     struct {
		User string `name:"User" description:"user login (or login@domain)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *userUpdateCmd) Execute(args []string) error {
	cl := cliClient

	userSpec, err := routevar.ParseUserSpec(c.Args.User)
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(cliContext, &userSpec)
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

		if _, err := cl.Accounts.Update(cliContext, user); err != nil {
			return err
		}
		fmt.Printf("# updated access level for user %s to %s\n", user.Login, c.Access)
	}

	updateBetas := func(betas []string) error {
		oldBetas := user.Betas
		user.Betas = betas
		if _, err := cl.Accounts.Update(cliContext, user); err != nil {
			return err
		}
		// Get the user again, because Update implicitly manages the pending
		// and accepted beta statuses.
		user, err = cl.Users.Get(cliContext, &userSpec)
		if err != nil {
			return err
		}
		fmt.Printf("# updated betas for user %s from %s to %s\n", user.Login, quotedComma(oldBetas), quotedComma(user.Betas))
		return nil
	}

	if c.SetBetas != "" {
		if err := updateBetas(commaSplit(c.SetBetas)); err != nil {
			return err
		}
	}
	if c.AddBetas != "" {
		newBetas := make([]string, len(user.Betas))
		copy(newBetas, user.Betas)
		for _, beta := range commaSplit(c.AddBetas) {
			if !user.InBeta(beta) {
				newBetas = append(newBetas, beta)
			}
		}
		if err := updateBetas(newBetas); err != nil {
			return err
		}
	}
	if c.RmBetas != "" {
		var (
			newBetas []string
			rmBetas  = commaSplit(c.RmBetas)
		)
		for _, beta := range user.Betas {
			remove := false
			for _, rmBeta := range rmBetas {
				if rmBeta == beta {
					remove = true
					break
				}
			}
			if !remove {
				newBetas = append(newBetas, beta)
			}
		}
		if err := updateBetas(newBetas); err != nil {
			return err
		}
	}

	return nil
}

// commaSplit splits the given string by "," and returns any non-whitespace strings.
func commaSplit(s string) []string {
	var out []string
	for _, s := range strings.Split(s, ",") {
		s = strings.TrimSpace(s)
		if len(s) > 0 {
			out = append(out, s)
		}
	}
	return out
}

type userResetPasswordCmd struct {
	Email string `long:"email" short:"e" description:"email address associated with account"`
	Login string `long:"login" short:"l" description:"login name of the user account"`
}

func (c *userResetPasswordCmd) Execute(args []string) error {
	cl := cliClient

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

	pendingReset, err := cl.Accounts.RequestPasswordReset(cliContext, person)
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
	cl := cliClient

	authInfo, err := cl.Auth.Identify(cliContext, &pbtypes.Void{})
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

	_, err = cl.Accounts.Delete(cliContext, person)
	if err != nil {
		return err
	}

	fmt.Printf("# User %q deleted.\n", identifier)

	return nil
}

type userRevokeExternalAuthCmd struct {
	Host     string `long:"host" description:"host of external authorization provider" default-mask:"github.com"`
	ClientID string `long:"client-id" description:"external OAuth2 client ID" default-mask:"(github.com value)"`

	Args struct {
		User string `name:"User" description:"user login"`
	} `positional-args:"yes" required:"yes"`
}

func (c *userRevokeExternalAuthCmd) Execute(args []string) error {
	cl := cliClient

	user, err := cl.Users.Get(cliContext, &sourcegraph.UserSpec{Login: c.Args.User})
	if err != nil {
		return err
	}

	_, err = cl.Auth.DeleteAndRevokeExternalToken(cliContext, &sourcegraph.ExternalTokenSpec{
		UID:      user.UID,
		Host:     c.Host,
		ClientID: c.ClientID,
	})
	if err != nil {
		return err
	}
	fmt.Printf("# External authorization for user %q deleted.\n", user.Login)
	return nil
}
