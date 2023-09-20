package tst

import (
	"context"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubScenarioUser struct {
	ScenarioResource
}

func NewGitHubScenarioUser(name string) *GitHubScenarioUser {
	return &GitHubScenarioUser{
		ScenarioResource: *NewScenarioResource(name),
	}
}

func (u *GitHubScenarioUser) ID() string {
	return u.id
}

func (u *GitHubScenarioUser) Name() string {
	return u.name
}

func (u *GitHubScenarioUser) Key() string {
	if u == &Admin {
		return "admin"
	}
	return u.key
}

var User1 GitHubScenarioUser = *NewGitHubScenarioUser("user1")
var User2 GitHubScenarioUser = *NewGitHubScenarioUser("user2")
var User3 GitHubScenarioUser = *NewGitHubScenarioUser("user3")
var User4 GitHubScenarioUser = *NewGitHubScenarioUser("user4")
var User5 GitHubScenarioUser = *NewGitHubScenarioUser("user5")
var User6 GitHubScenarioUser = *NewGitHubScenarioUser("user6")
var User7 GitHubScenarioUser = *NewGitHubScenarioUser("user7")
var User8 GitHubScenarioUser = *NewGitHubScenarioUser("user8")
var User9 GitHubScenarioUser = *NewGitHubScenarioUser("user9")
var User10 GitHubScenarioUser = *NewGitHubScenarioUser("user10")
var Admin GitHubScenarioUser = *NewGitHubScenarioUser("admin")

func preloadUsersAction(client *GitHubClient) Action {
	return &action{
		id:   "",
		name: "internal.preload-users",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}
			users, err := client.orgUsers(ctx, org)
			if err != nil {
				return nil, err
			}
			store.SetUsers(users)

			return &actionResult[[]*github.User]{item: users}, nil
		},
	}
}

func mapUsersAction(_ *GitHubClient, scenarioUsers []*GitHubScenarioUser) Action {
	return &action{
		id:   "",
		name: "internal.map-scenario-users()",
		fn: func(_ context.Context, store *ScenarioStore) (ActionResult, error) {
			users, err := store.GetUsers()
			if err != nil {
				return nil, err
			}
			if len(scenarioUsers) > len(users) {
				return nil, errors.Newf("not enough users to use for scenario - required %d, available %d", len(scenarioUsers), len(users))
			}

			for i, user := range scenarioUsers {
				store.SetScenarioUserMapping(user, users[i])
			}
			return &actionResult[bool]{item: true}, nil
		},
	}
}

func userEmail(u *GitHubScenarioUser) string {
	return "tst-pkg-user@sourcegraph.com"
}

func (u *GitHubScenarioUser) GetUserAction(client *GitHubClient) Action {
	// we have to deref here otherwise inside the action the user changes to
	// not what the user is AT THIS POINT
	scenarioUser := *u
	return &action{
		id:   u.Key(),
		name: "get-github-user-" + u.Name(),
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			var username = scenarioUser.Name()
			if username == Admin.Name() {
				username = client.cfg.User
			}
			user, err := client.GetUser(ctx, username)
			if err != nil {
				return nil, err
			}

			store.SetScenarioUserMapping(&scenarioUser, user)
			return &actionResult[*github.User]{item: user}, nil
		},
	}
}

func (u *GitHubScenarioUser) CreateUserAction(client *GitHubClient) Action {
	return &action{
		id:   u.Key(),
		name: "create-user",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			user, err := client.CreateUser(ctx, "test", userEmail(u))
			if err != nil {
				return nil, err
			}

			store.SetScenarioUserMapping(u, user)
			return &actionResult[*github.User]{item: user}, nil
		},
	}
}

func (u GitHubScenarioUser) DeleteUserAction(client *GitHubClient) Action {
	return &action{
		id:   u.Key(),
		name: "delete-user",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			user, err := store.GetScenarioUser(u)
			if err != nil {
				return nil, err
			}
			err = client.DeleteUser(ctx, user.GetLogin())
			if err != nil {
				return nil, err
			}

			return &actionResult[*github.User]{item: user}, nil
		},
	}
}
