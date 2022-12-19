package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"

	"github.com/sourcegraph/src-cli/internal/api"
)

func init() {
	usage := `
This command removes users from a Sourcegraph instance who have been inactive for 60 or more days. Admin accounts are omitted by default.
	
Examples:

	$ src users prune -days 182
	
	$ src users prune -remove-admin -remove-null-users
`

	flagSet := flag.NewFlagSet("prune", flag.ExitOnError)
	usageFunc := func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of 'src users %s':\n", flagSet.Name())
		flagSet.PrintDefaults()
		fmt.Println(usage)
	}
	var (
		daysToDelete       = flagSet.Int("days", 60, "Days threshold on which to remove users, must be 60 days or greater and defaults to this value ")
		removeAdmin        = flagSet.Bool("remove-admin", false, "prune admin accounts")
		removeNoLastActive = flagSet.Bool("remove-null-users", false, "removes users with no last active value")
		skipConfirmation   = flagSet.Bool("force", false, "skips user confirmation step allowing programmatic use")
		apiFlags           = api.NewFlags(flagSet)
	)

	handler := func(args []string) error {
		if err := flagSet.Parse(args); err != nil {
			return err
		}
		if *daysToDelete < 60 {
			fmt.Println("-days flag must be set to 60 or greater")
			return nil
		}

		ctx := context.Background()
		client := cfg.apiClient(apiFlags, flagSet.Output())

		currentUserQuery := `
query getCurrentUser {
	currentUser {
		username
	}
}
`
		var currentUserResult struct {
			Data struct {
				CurrentUser struct {
					Username string
				}
			}
		}
		if ok, err := cfg.apiClient(apiFlags, flagSet.Output()).NewRequest(currentUserQuery, nil).DoRaw(context.Background(), &currentUserResult); err != nil || !ok {
			return err
		}

		getInactiveUsersQuery := `
query getInactiveUsers {
	site {
		users {
			nodes {
				username
				email
				siteAdmin
				lastActiveAt
			}
		}
	}
}
`

		var usersResult struct {
			Site struct {
				Users struct {
					Nodes []SiteUser
				}
			}
		}

		if ok, err := client.NewRequest(getInactiveUsersQuery, nil).Do(ctx, &usersResult); err != nil || !ok {
			return err
		}

		usersToDelete := make([]UserToDelete, 0)
		for _, user := range usersResult.Site.Users.Nodes {
			daysSinceLastUse, hasLastActive, err := computeDaysSinceLastUse(user)
			if err != nil {
				return err
			}
			// never remove user issuing command
			if user.Username == currentUserResult.Data.CurrentUser.Username {
				continue
			}
			if !hasLastActive && !*removeNoLastActive {
				continue
			}
			if !*removeAdmin && user.SiteAdmin {
				continue
			}
			if daysSinceLastUse <= *daysToDelete && hasLastActive {
				continue
			}
			deleteUser := UserToDelete{user, daysSinceLastUse}

			usersToDelete = append(usersToDelete, deleteUser)
		}

		if *skipConfirmation {
			for _, user := range usersToDelete {
				if err := removeUser(user.User, client, ctx); err != nil {
					return err
				}
			}
			return nil
		}

		// confirm and remove users
		if confirmed, _ := confirmUserRemoval(usersToDelete); !confirmed {
			fmt.Println("Aborting removal")
			return nil
		} else {
			fmt.Println("REMOVING USERS")
			for _, user := range usersToDelete {
				if err := removeUser(user.User, client, ctx); err != nil {
					return err
				}
			}
		}

		return nil
	}

	// Register the command.
	usersCommands = append(usersCommands, &command{
		flagSet:   flagSet,
		handler:   handler,
		usageFunc: usageFunc,
	})
}

// computes days since last usage from current day and time and UsageStatistics.LastActiveTime, uses time.Parse
func computeDaysSinceLastUse(user SiteUser) (timeDiff int, hasLastActive bool, _ error) {
	// handle for null LastActiveAt returned from
	if user.LastActiveAt == "" {
		hasLastActive = false
		return 0, hasLastActive, nil
	}
	timeLast, err := time.Parse(time.RFC3339, user.LastActiveAt)
	if err != nil {
		return 0, false, err
	}
	timeDiff = int(time.Since(timeLast).Hours() / 24)

	return timeDiff, true, err
}

// Issue graphQL api request to remove user
func removeUser(user SiteUser, client api.Client, ctx context.Context) error {
	query := `mutation DeleteUser($user: ID!) {
  deleteUser(user: $user) {
    alwaysNil
  }
}`
	vars := map[string]interface{}{
		"user": user.ID,
	}
	if ok, err := client.NewRequest(query, vars).Do(ctx, nil); err != nil || !ok {
		return err
	}
	return nil
}

type UserToDelete struct {
	User             SiteUser
	DaysSinceLastUse int
}

// Verify user wants to remove users with table of users and a command prompt for [y/N]
func confirmUserRemoval(usersToRemove []UserToDelete) (bool, error) {
	fmt.Printf("Users to remove from instance at %s\n", cfg.Endpoint)
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Username", "Email", "Days Since Last Active"})
	for _, user := range usersToRemove {
		if user.User.Email != "" {
			t.AppendRow([]interface{}{user.User.Username, user.User.Email, user.DaysSinceLastUse})
			t.AppendSeparator()
		} else {
			t.AppendRow([]interface{}{user.User.Username, "", user.DaysSinceLastUse})
			t.AppendSeparator()
		}
	}
	t.SetStyle(table.StyleRounded)
	t.Render()
	input := ""
	for strings.ToLower(input) != "y" && strings.ToLower(input) != "n" {
		fmt.Printf("Do you  wish to proceed with user removal [y/N]: ")
		if _, err := fmt.Scanln(&input); err != nil {
			return false, err
		}
	}
	return strings.ToLower(input) == "y", nil
}
