package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/sourcegraph/sourcegraph/lib/group"
)

func delete(ctx context.Context, cfg config) {
	localOrgs, err := store.loadOrgs()
	if err != nil {
		log.Fatalf("Failed to load orgs from state: %s", err)
	}

	if len(localOrgs) == 0 {
		// Fetch orgs currently on instance due to lost state
		remoteOrgs := getGitHubOrgs(ctx)

		writeInfo(out, "Storing %d orgs in state", len(remoteOrgs))
		for _, o := range remoteOrgs {
			if strings.HasPrefix(*o.Name, "org-") {
				o := &org{
					Login:   *o.Login,
					Admin:   cfg.orgAdmin,
					Failed:  "",
					Created: true,
				}
				if err := store.saveOrg(o); err != nil {
					log.Fatalf("Failed to store orgs in state: %s", err)
				}
				localOrgs = append(localOrgs, o)
			}
		}
	}

	localUsers, err := store.loadUsers()
	if err != nil {
		log.Fatalf("Failed to load users from state: %s", err)
	}

	if len(localUsers) == 0 {
		remoteUsers := getGitHubUsers(ctx)

		writeInfo(out, "Storing %d users in state", len(remoteUsers))
		for _, u := range remoteUsers {
			if strings.HasPrefix(*u.Login, "user-") {
				u := &user{
					// Fetch users currently on instance due to lost state
					Login:   *u.Login,
					Email:   fmt.Sprintf("%s@%s", *u.Login, emailDomain),
					Failed:  "",
					Created: true,
				}
				if err := store.saveUser(u); err != nil {
					log.Fatalf("Failed to store users in state: %s", err)
				}
				localUsers = append(localUsers, u)
			}
		}
	}

	localTeams, err := store.loadTeams()
	if err != nil {
		log.Fatalf("Failed to load teams from state: %s", err)
	}

	if len(localTeams) == 0 {
		// Fetch teams currently on instance due to lost state
		remoteTeams := getGitHubTeams(ctx, localOrgs)

		writeInfo(out, "Storing %d teams in state", len(remoteTeams))
		for _, t := range remoteTeams {
			if strings.HasPrefix(*t.Name, "team-") {
				t := &team{
					Name:         *t.Name,
					Org:          *t.Organization.Login,
					Failed:       "",
					Created:      true,
					TotalMembers: 0, //not important for deleting but subsequent use of state will be problematic
				}
				if err := store.saveTeam(t); err != nil {
					log.Fatalf("Failed to store teams in state: %s", err)
				}
				localTeams = append(localTeams, t)
			}
		}
	}

	g := group.New().WithMaxConcurrency(1000)

	// delete users from instance
	usersToDelete := len(localUsers) - cfg.userCount
	for i := 0; i < usersToDelete; i++ {
		currentUser := localUsers[i]
		if i%100 == 0 {
			writeInfo(out, "Deleted %d out of %d users", i, usersToDelete)
		}
		g.Go(func() {
			executeDeleteUser(ctx, currentUser)
		})
	}

	teamsToDelete := len(localTeams) - cfg.teamCount
	for i := 0; i < teamsToDelete; i++ {
		currentTeam := localTeams[i]
		if i%100 == 0 {
			writeInfo(out, "Deleted %d out of %d teams", i, teamsToDelete)
		}
		g.Go(func() {
			executeDeleteTeam(ctx, currentTeam)
		})
	}
	g.Wait()

	for _, t := range localTeams {
		currentTeam := t
		g.Go(func() {
			executeDeleteTeamMembershipsForTeam(ctx, currentTeam.Org, currentTeam.Name)
		})
	}
	g.Wait()
}
