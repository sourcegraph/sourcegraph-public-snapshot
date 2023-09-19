package tst

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/go-github/v53/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubScenarioRepo struct {
	ScenarioResource
	teamName string
	fork     bool
	private  bool
}

func NewGitHubScenarioRepo(name, teamKey string, fork, private bool) *GitHubScenarioRepo {
	return &GitHubScenarioRepo{
		ScenarioResource: *NewScenarioResource(name),
		teamName:         teamKey,
		fork:             fork,
		private:          private,
	}
}

func (gr *GitHubScenarioRepo) ForkRepoAction(client *GitHubClient) Action {
	return &action{
		id:   gr.Key(),
		name: "fork-repo",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			var owner, repoName string
			parts := strings.Split(gr.name, "/")
			if len(parts) >= 2 {
				owner = parts[0]
				repoName = parts[1]
			} else {
				return nil, errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			err = client.ForkRepo(ctx, org, owner, repoName)
			if err != nil {
				return nil, err
			}
			return &actionResult[bool]{item: true}, nil
		},
	}
}

func (gr *GitHubScenarioRepo) GetRepoAction(client *GitHubClient) Action {
	return &action{
		name: fmt.Sprintf("get-repo(%s)", gr.Key()),
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			// Wait till fork has synced
			time.Sleep(1 * time.Second)
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			var repoName string
			parts := strings.Split(gr.name, "/")
			if len(parts) >= 2 {
				repoName = parts[1]
			} else {
				repoName = parts[0]
			}
			if gr.fork && repoName == "" {
				return nil, errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			repo, err := client.GetRepo(ctx, org.GetLogin(), repoName)
			if err != nil {
				return nil, err
			}
			// Since this is a forked repo we need to update the GitHubScenarioRepo
			// We only edit the name but the id stays the same because the initial name
			// is "someorg/repo" and the name should reflect the name with the current org
			// "currentOrg/repo"
			// TODO: this is nasty - find a better way
			gr.name = repo.GetFullName()
			store.SetRepo(gr, repo)
			return &actionResult[bool]{item: true}, nil
		},
	}

}

func (gr *GitHubScenarioRepo) InitLocalRepoAction(client *GitHubClient) Action {
	return &action{
		id:   gr.Key(),
		name: "init-new-repo",
		// this should ideally be two actions but we need a nice way to share the directory location between the two actions
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			githubRepo, err := store.GetRepo(gr)
			if err != nil {
				return nil, err
			}

			localRepo, err := NewLocalRepo(githubRepo.GetName(), client.cfg.User, client.cfg.Password)
			if err != nil {
				return nil, err
			}

			err = localRepo.Init(ctx)
			if err != nil {
				return nil, err
			}

			err = localRepo.AddRemote(ctx, githubRepo.GetGitURL())
			if err != nil {
				return nil, err
			}

			err = localRepo.PushRemote(ctx, 5)
			if err != nil {
				return nil, err
			}

			localRepo.Cleanup()

			return &actionResult[LocalRepo]{item: *localRepo}, nil
		},
	}
}

func (gr *GitHubScenarioRepo) NewRepoAction(client *GitHubClient) Action {
	return &action{
		id:   gr.Key(),
		name: "create-repo",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			var repoName string
			parts := strings.Split(gr.name, "/")
			if len(parts) >= 2 {
				repoName = parts[1]
			} else {
				repoName = parts[0]
			}

			repo, err := client.NewRepo(ctx, org, repoName, gr.private)
			if err != nil {
				return nil, err
			}

			gr.name = repo.GetFullName()
			store.SetRepo(gr, repo)

			return &actionResult[bool]{item: true}, nil
		},
	}
}

func (gr *GitHubScenarioRepo) SetPermissionsAction(client *GitHubClient) Action {
	return &action{
		id:   gr.Key(),
		name: "repo-permissions",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			repo, err := store.GetRepo(gr)
			if err != nil {
				return nil, err
			}

			repo.Private = &gr.private

			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			repo, err = client.UpdateRepo(ctx, org, repo)
			if err != nil {
				return nil, err
			}
			store.SetRepo(gr, repo)
			return &actionResult[*github.Repository]{item: repo}, nil
		},
	}
}

func (gr *GitHubScenarioRepo) DeleteRepoAction(client *GitHubClient) Action {
	return &action{
		id:   gr.Key(),
		name: "delete-repo",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			repo, err := store.GetRepo(gr)
			if err != nil {
				return nil, err
			}

			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			err = client.DeleteRepo(ctx, org, repo)
			if err != nil {
				return nil, err
			}
			store.SetRepo(gr, repo)
			return &actionResult[bool]{item: true}, nil
		},
	}
}

func (gr *GitHubScenarioRepo) AssignTeamAction(client *GitHubClient) Action {
	return &action{
		id:   gr.Key(),
		name: "assign-team-" + gr.teamName,
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			repo, err := store.GetRepo(gr)
			if err != nil {
				return nil, err
			}

			team, err := store.GetTeamByName(gr.teamName)
			if err != nil {
				return nil, err
			}

			err = client.UpdateTeamRepoPermissions(ctx, org, team, repo)
			if err != nil {
				return nil, err
			}
			store.SetRepo(gr, repo)
			return &actionResult[bool]{item: true}, nil
		},
	}
}
