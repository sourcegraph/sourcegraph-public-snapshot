package codehost_testing

import (
	"context"
	"fmt"

	"github.com/google/go-github/v55/github"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Org represents a GitHub organization and provides actions that operate on the org.
//
// All methods except Get, create actions which are added to the GitHubScenario this
// org was was created from.
type Org struct {
	// s is the GithubScenario instance this org was created from
	s *GitHubScenario
	// name is the name of the GitHub organization
	name string
}

// Get returns the corresponding GitHub Organization object that was created by the `CreateOrg`
//
// This method will only return a Org if the Scenario that created it has been applied otherwise
// it will panic.
func (o *Org) Get(ctx context.Context) (*github.Organization, error) {
	if o.s.IsApplied() {
		return o.s.client.GetOrg(ctx, o.name)
	}
	return nil, errors.New("cannot retrieve org before scenario is applied")
}

// get retrieves the GitHub organization without panicking if not applied. It is meant as an
// internal helper method while actions are getting applied.
func (o *Org) get(ctx context.Context) (*github.Organization, error) {
	return o.s.client.GetOrg(ctx, o.name)
}

// AllowPrivateForks adds an action to the scenario to enable private forks and repos for the org
func (o *Org) AllowPrivateForks() {
	updateOrgPermissions := &Action{
		Name: "org:permissions:update:" + o.name,
		Apply: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			org.MembersCanCreatePrivateRepos = github.Bool(true)
			org.MembersCanForkPrivateRepos = github.Bool(true)

			_, err = o.s.client.UpdateOrg(ctx, org)
			if err != nil {
				return err
			}
			return nil
		},
		Teardown: nil,
	}
	o.s.Append(updateOrgPermissions)
}

// CreateTeam adds an action to the scenario to create a team with the given name for the org.
// The Scenario ID will be added as a suffix to the given name.
func (o *Org) CreateTeam(name string) any {
	createTeam := &Action{
		Name: "org:team:create:" + name,
		Apply: func(ctx context.Context) error {
			return nil
		},
		Teardown: func(ctx context.Context) error {
			return nil
		},
	}

	o.s.Append(createTeam)

	return nil
}

// CreateRepo adds an action to the scenario to create a repo with the given name and visibility for the org.
func (o *Org) CreateRepo(name string, public bool) any {
	action := &Action{
		Name: fmt.Sprintf("repo:create:%s", name),
		Apply: func(ctx context.Context) error {
			return nil
		},
		Teardown: func(ctx context.Context) error {
			return nil
		},
	}
	o.s.Append(action)

	return nil
}

// CreateRepoFork adds an action to the scenario to fork a target repo into the org.
func (o *Org) CreateRepoFork(target string) any {
	action := &Action{
		Name: fmt.Sprintf("repo:fork:%s", target),
		Apply: func(ctx context.Context) error {
			return nil
		},
		Teardown: func(ctx context.Context) error {
			return nil
		},
	}
	o.s.Append(action)

	return nil
}
