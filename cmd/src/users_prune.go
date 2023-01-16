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
		daysToDelete         = flagSet.Int("days", 60, "Days threshold on which to remove users, must be 60 days or greater and defaults to this value ")
		removeAdmin          = flagSet.Bool("remove-admin", false, "prune admin accounts")
		removeNoLastActive   = flagSet.Bool("remove-null-users", false, "removes users with no last active value")
		skipConfirmation     = flagSet.Bool("force", false, "skips user confirmation step allowing programmatic use")
		displayUsersToDelete = flagSet.Bool("display-users", false, "display table of users to be deleted by prune")
		apiFlags             = api.NewFlags(flagSet)
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

		// get current user so as not to delete issuer of the prune request
		currentUserQuery := `query getCurrentUser { currentUser { username }}`
		var currentUserResult struct {
			CurrentUser struct {
				Username string
			}
		}
		if ok, err := cfg.apiClient(apiFlags, flagSet.Output()).NewRequest(currentUserQuery, nil).Do(context.Background(), &currentUserResult); err != nil || !ok {
			return err
		}

		// get total users to paginate over
		totalUsersQuery := `query getTotalUsers { site { users { totalCount }}}`
		var totalUsers struct {
			Site struct {
				Users struct {
					TotalCount float64
				}
			}
		}
		if ok, err := cfg.apiClient(apiFlags, flagSet.Output()).NewRequest(totalUsersQuery, nil).Do(context.Background(), &totalUsers); err != nil || !ok {
			return err
		}

		// get 100 site users
		getInactiveUsersQuery := `
		query getInactiveUsers($limit: Int $offset: Int) {
	site {
		users {
			nodes (limit: $limit offset: $offset) {
				id
				username
				email
				siteAdmin
				lastActiveAt
				deletedAt
			}
		}
	}
}
`

		// paginate through users
		var aggregatedUsers []SiteUser
		// pagination variables, limit set to maximum possible users returned per request
		offset := 0
		const limit int = 100

		// paginate requests until all site users have been checked -- this includes soft deleted users
		for len(aggregatedUsers) < int(totalUsers.Site.Users.TotalCount) {
			pagVars := map[string]interface{}{
				"offset": offset,
				"limit":  limit,
			}

			var usersResult struct {
				Site struct {
					Users struct {
						Nodes []SiteUser
					}
					TotalCount float64
				}
			}
			if ok, err := client.NewRequest(getInactiveUsersQuery, pagVars).Do(ctx, &usersResult); err != nil || !ok {
				return err
			}
			// increment graphql request offset by the length of the last user set returned
			offset = offset + len(usersResult.Site.Users.Nodes)
			// append graphql user results to aggregated users to be processed against user removal conditions
			aggregatedUsers = append(aggregatedUsers, usersResult.Site.Users.Nodes...)
		}

		// filter users for deletion
		usersToDelete := make([]UserToDelete, 0)
		for _, user := range aggregatedUsers {
			// never remove user issuing command
			if user.Username == currentUserResult.CurrentUser.Username {
				continue
			}
			// filter out soft deleted users returned by site graphql endpoint
			if user.DeletedAt != "" {
				continue
			}
			//compute days since last use
			daysSinceLastUse, hasLastActive, err := computeDaysSinceLastUse(user)
			if err != nil {
				return err
			}
			// don't remove users with no last active value unless option flag is set
			if !hasLastActive && !*removeNoLastActive {
				continue
			}
			// don't remove admins unless option flag is set
			if !*removeAdmin && user.SiteAdmin {
				continue
			}
			// remove users who have been inactive for longer than the threshold set by the -days flag
			if daysSinceLastUse <= *daysToDelete && hasLastActive {
				continue
			}
			// serialize user to print in table as part of confirmUserRemoval, add to delete slice
			userToDelete := UserToDelete{user, daysSinceLastUse}
			usersToDelete = append(usersToDelete, userToDelete)
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
		if confirmed, _ := confirmUserRemoval(usersToDelete, int(totalUsers.Site.Users.TotalCount), *daysToDelete, *displayUsersToDelete); !confirmed {
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

// computes days since last usage from current day and time and aggregated_user_statistics.lastActiveAt, uses time.Parse
func computeDaysSinceLastUse(user SiteUser) (timeDiff int, hasLastActive bool, _ error) {
	// handle for null LastActiveAt, users who have never been active
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
	query := `mutation DeleteUser($user: ID!) { deleteUser(user: $user) { alwaysNil }}`
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
func confirmUserRemoval(usersToDelete []UserToDelete, totalUsers int, daysThreshold int, displayUsers bool) (bool, error) {
	if displayUsers {
		fmt.Printf("Users to remove from %s\n", cfg.Endpoint)
		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Username", "Email", "Days Since Last Active"})
		for _, user := range usersToDelete {
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
	}
	input := ""
	for strings.ToLower(input) != "y" && strings.ToLower(input) != "n" {
		fmt.Printf("%v users were inactive for more than %v days on %v.\nDo you  wish to proceed with user removal [y/N]: ", len(usersToDelete), daysThreshold, cfg.Endpoint)
		if _, err := fmt.Scanln(&input); err != nil {
			return false, err
		}
	}
	return strings.ToLower(input) == "y", nil
}
