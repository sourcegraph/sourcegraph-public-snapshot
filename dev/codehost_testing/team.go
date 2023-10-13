package codehost_testing

import (
	"context"
	"fmt"

	"github.com/google/go-github/v55/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Team represents a GitHub team and provdes actions that operate on a GitHub team.
//
// All methods except Get, create actions which are added to the parent GitHubScenario this
// team belongs to.
type Team struct {
	// s is the GithubScenario instance this team was created from
	s *GitHubScenario
	// org is the Org this team belongs to, and is ultimately the one who created this team
	org *Org
	// name is the name of the team
	name string
}

// Get returns the corresponding GitHub Team object that was created by the `CreateTeam`
//
// This method will only return a Team if the Scenario that created it has been applied otherwise
// it will panic.
func (team *Team) Get(ctx context.Context) (*github.Team, error) {
	if team.s.IsApplied() {
		return team.get(ctx)
	}
	return nil, errors.New("cannot retrieve org before scenario is applied")
}

// get retrieves the GitHub team without panicking if not applied. It is meant as an
// internal helper method while actions are getting applied.
func (team *Team) get(ctx context.Context) (*github.Team, error) {
	return team.s.client.GetTeam(ctx, team.org.name, team.name)
}

// AddUser adds an action that will add the given user to this team
func (tm *Team) AddUser(u *User) {
	assignTeamMembership := &Action{
		Name: fmt.Sprintf("team:membership:%s:%s", tm.name, u.name),
		Apply: func(ctx context.Context) error {
			org, err := tm.org.get(ctx)
			if err != nil {
				return err
			}
			team, err := tm.get(ctx)
			if err != nil {
				return err
			}
			user, err := u.get(ctx)
			if err != nil {
				return err
			}
			_, err = tm.s.client.AssignTeamMembership(ctx, org, team, user)
			return err
		},
		Teardown: nil,
	}

	tm.s.Append(assignTeamMembership)
}
