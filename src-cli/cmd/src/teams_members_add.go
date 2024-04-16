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

  Add a team member:

    	$ src teams members add -team-name='engineering' [-email='alice@sourcegraph.com'] [-username='alice'] [-id='VXNlcjox'] [-external-account-service-id='https://github.com/' -external-account-service-type='github' [-external-account-account-id='123123123'] [-external-account-login='alice']]

`

	flagSet := flag.NewFlagSet("add", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src teams %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		teamNameFlag                   = flagSet.String("team-name", "", "The team name")
		skipUnmatchedMembersFlag       = flagSet.Bool("skip-unmatched-members", false, "If true, members that don't match a Sourcegraph user or team will be silently skipped")
		emailFlag                      = flagSet.String("email", "", "Email to match the user by")
		usernameFlag                   = flagSet.String("username", "", "Username to match the user by")
		idFlag                         = flagSet.String("id", "", "Sourcegraph user ID to match the user by")
		externalAccountServiceIDFlag   = flagSet.String("external-account-service-id", "", "External account service ID to match the user by, must specify all of externalAccount*")
		externalAccountServiceTypeFlag = flagSet.String("external-account-service-type", "", "External account service type to match the user by, must specify all of externalAccount*")
		externalAccountAccountIDFlag   = flagSet.String("external-account-account-id", "", "External account account ID to match the user by, must specify all of externalAccount*")
		externalAccountLoginFlag       = flagSet.String("external-account-login", "", "External account login ID to match the user by, must specify all of externalAccount*")
		apiFlags                       = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}

		if *teamNameFlag == "" {
			return errors.New("provide a team name")
		}

		client := cfg.apiClient(apiFlags, flagSet.Output())

		query := `mutation AddTeamMember(
	$teamName: String!
	$id: ID,
	$email: String,
	$username: String,
	$externalAccountServiceID: String,
	$externalAccountServiceType: String,
	$externalAccountAccountID: String,
	$externalAccountLogin: String,
	$skipUnmatchedMembers: Boolean,
) {
	addTeamMembers(
		teamName: $teamName,
		members: [{
			userID: $id,
			email: $email,
			username: $username,
			externalAccountServiceID: $externalAccountServiceID,
			externalAccountServiceType: $externalAccountServiceType,
			externalAccountAccountID: $externalAccountAccountID,
			externalAccountLogin: $externalAccountLogin,
		}],
		skipUnmatchedMembers: $skipUnmatchedMembers,
	) {
		...TeamFields
	}
}
` + teamFragment

		var result struct {
			AddTeamMembers Team
		}
		if ok, err := client.NewRequest(query, map[string]interface{}{
			"teamName":                   *teamNameFlag,
			"skipUnmatchedMembers":       *skipUnmatchedMembersFlag,
			"id":                         api.NullString(*idFlag),
			"email":                      api.NullString(*emailFlag),
			"username":                   api.NullString(*usernameFlag),
			"externalAccountServiceID":   api.NullString(*externalAccountServiceIDFlag),
			"externalAccountServiceType": api.NullString(*externalAccountServiceTypeFlag),
			"externalAccountAccountID":   api.NullString(*externalAccountAccountIDFlag),
			"externalAccountLogin":       api.NullString(*externalAccountLoginFlag),
		}).Do(context.Background(), &result); err != nil || !ok {
			return err
		}

		return nil
	}

	// Register the command.
	teamMembersCommands = append(teamMembersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}
