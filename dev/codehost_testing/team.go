package codehost_scenario

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type Team struct {
	s    *GithubScenario
	org  *Org
	name string
}

func (team *Team) Get(ctx context.Context) (*github.Team, error) {
	if team.s.IsApplied() {
		return team.get(ctx)
	}
	panic("cannot retrieve org before scenario is applied")
}

func (team *Team) get(ctx context.Context) (*github.Team, error) {
	return team.s.client.GetTeam(ctx, team.org.name, team.name)
}

func (tm *Team) AddUser(u *User) {
	assignTeamMembership := &action{
		name: fmt.Sprintf("team:membership:%s:%s", tm.name, u.name),
		apply: func(ctx context.Context) error {
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
		teardown: nil,
	}

	tm.s.append(assignTeamMembership)
}
