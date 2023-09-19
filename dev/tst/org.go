package tst

import (
	"context"
	"fmt"

	"github.com/google/go-github/v53/github"
)

type GitHubScenarioOrg struct {
	ScenarioResource
}

func NewGitHubScenarioOrg(name string) *GitHubScenarioOrg {
	return &GitHubScenarioOrg{
		ScenarioResource: *NewScenarioResource(name),
	}
}

func (o *GitHubScenarioOrg) ID() string {
	return o.id
}

func (o *GitHubScenarioOrg) Name() string {
	return o.name
}

func (o *GitHubScenarioOrg) Key() string {
	return o.key
}

// CreateOrgAction creates an Action which will create a GitHub Organization
// with a unique ID. The created organization is added to the store
func (g GitHubScenarioOrg) CreateOrgAction(client *GitHubClient) Action {
	return &action{
		id:   g.Key(),
		name: "create-org",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := client.CreateOrg(ctx, g.Key())
			if err != nil {
				return nil, err
			}
			store.SetOrg(org)
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}
}

// UpdateOrgPermissionsAction creates an Action that will update an organization
// to allow members to:
// - create private repos
// - fork private repos
//
// This action requires that the store contains a Github Organization which
// can be loaded with the `CreateOrgAction`.
func (g GitHubScenarioOrg) UpdateOrgPermissionsAction(client *GitHubClient) Action {
	return &action{
		id:   g.Key(),
		name: "update-org-permissions",
		fn: func(ctx context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}

			org.MembersCanCreatePrivateRepos = boolp(true)
			org.MembersCanForkPrivateRepos = boolp(true)

			org, err = client.UpdateOrg(ctx, org)
			if err != nil {
				return nil, err
			}
			store.SetOrg(org)
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}

}

// DeleteOrgAction creates an Action that will delete the organization. At
// the moment it just prints a message, since our current GHE instance API
// does not support deleting organizations.
func (g GitHubScenarioOrg) DeleteOrgAction(client *GitHubClient) Action {
	return &action{
		id:   g.Key(),
		name: fmt.Sprintf("delete-org(%s)", g.Key()),
		fn: func(_ context.Context, store *ScenarioStore) (ActionResult, error) {
			org, err := store.GetOrg()
			if err != nil {
				return nil, err
			}
			// Our GHE instance needs to be updated to support the delete API
			// so for now we just print out that we need to delete it ...
			fmt.Printf("NEED TO DELETE ORG: %s\n", org.GetLogin())
			return &actionResult[*github.Organization]{item: org}, nil
		},
	}
}
