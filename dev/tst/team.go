package tst

import (
	"context"

	"github.com/google/go-github/v53/github"
)

type GitHubScenarioTeam struct {
	name  string
	id    string
	key   string
	users []GitHubScenarioUser
}

func NewGitHubScenarioTeam(name string, u ...GitHubScenarioUser) *GitHubScenarioTeam {
	id := id()
	key := joinID(name, "-", id, 39)
	return &GitHubScenarioTeam{
		name:  name,
		id:    id,
		key:   key,
		users: u,
	}
}

func (t *GitHubScenarioTeam) ID() string {
	return t.id
}

func (t *GitHubScenarioTeam) Name() string {
	return t.name
}

func (t *GitHubScenarioTeam) Key() string {
	return t.key
}

func (gt *GitHubScenarioTeam) CreateTeamAction(client *GitHubClient) Action {
	return &action{
		id:   gt.Key(),
		name: "get-or-create-team" + gt.name,
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}
			newTeam, err := client.GetOrCreateTeam(ctx, org, gt.name)
			if err != nil {
				return nil, err
			}
			store.SetTeam(gt, newTeam)
			return &actionResult[*github.Team]{item: newTeam}, nil
		},
	}
}

func (gt *GitHubScenarioTeam) DeleteTeamAction(client *GitHubClient) Action {
	return &action{
		id:   gt.Key(),
		name: "delete-team(%s)",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}
			err = client.DeleteTeam(ctx, org, gt.name)
			if err != nil {
				return nil, err
			}
			return &actionResult[bool]{item: true}, nil
		},
	}
}

func (gt *GitHubScenarioTeam) AssignTeamAction(client *GitHubClient) Action {
	return &action{
		id:   gt.Key(),
		name: "assign-team-membership",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}
			team, err := store.GetTeam(gt)
			if err != nil {
				return nil, err
			}
			teamUsers := make([]*github.User, 0)
			for _, u := range gt.users {
				if ghUser, err := store.GetScenarioUser(u); err == nil {
					teamUsers = append(teamUsers, ghUser)
				} else {
					return nil, err
				}
				client.AssignTeamMembership(ctx, org, team, teamUsers...)
			}

			return &actionResult[*github.Team]{item: team}, nil
		},
	}
}
