package codehost_testing

import (
	"context"
	"fmt"
	"strings"

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
func (o *Org) CreateTeam(name string) *Team {
	baseTeam := &Team{
		s:    o.s,
		org:  o,
		name: name,
	}

	action := &Action{
		Name: "org:team:create:" + name,
		Apply: func(ctx context.Context) error {
			name := fmt.Sprintf("team-%s-%s", name, o.s.id)
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			team, err := o.s.client.CreateTeam(ctx, org, name)
			if err != nil {
				return err
			}
			baseTeam.name = team.GetName()
			return nil
		},
		Teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}
			return o.s.client.DeleteTeam(ctx, org, baseTeam.name)
		},
	}
	o.s.Append(action)

	return baseTeam
}

// CreateRepo adds an action to the scenario to create a repo with the given name and visibility for the org.
func (o *Org) CreateRepo(name string, public bool) *Repo {
	baseRepo := &Repo{
		s:    o.s,
		org:  o,
		name: name,
	}
	action := &Action{
		Name: fmt.Sprintf("repo:create:%s", name),
		Apply: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			var repoName string
			parts := strings.Split(name, "/")
			if len(parts) >= 2 {
				repoName = parts[1]
			} else {
				return errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			repo, err := o.s.client.CreateRepo(ctx, org, repoName, public)
			if err != nil {
				return err
			}

			baseRepo.name = repo.GetFullName()
			return nil
		},
		Teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			repo, err := baseRepo.get(ctx)
			if err != nil {
				return err
			}

			return o.s.client.DeleteRepo(ctx, org, repo)
		},
	}
	o.s.Append(action)

	return baseRepo
}

// CreateRepoFork adds an action to the scenario to fork a target repo into the org.
//
// NOTE: This method actually adds two actions to the scenario. One which performs the Fork and a subsequent
// action which waits till the forked repo exists on GitHub.
func (o *Org) CreateRepoFork(target string) *Repo {
	baseRepo := &Repo{
		s:    o.s,
		org:  o,
		name: target,
	}
	action := &Action{
		Name: fmt.Sprintf("repo:fork:%s", target),
		Apply: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			var owner, repoName string
			parts := strings.Split(target, "/")
			if len(parts) >= 2 {
				owner = parts[0]
				repoName = parts[1]
			} else {
				return errors.Newf("incorrect repo format for %q - expecting {owner}/{name}")
			}

			err = o.s.client.ForkRepo(ctx, org, owner, repoName)
			if err != nil {
				return err
			}

			// Wait till fork has synced
			baseRepo.name = repoName
			return nil
		},
		Teardown: func(ctx context.Context) error {
			org, err := o.get(ctx)
			if err != nil {
				return err
			}

			repo, err := baseRepo.get(ctx)
			if err != nil {
				return err
			}

			return o.s.client.DeleteRepo(ctx, org, repo)
		},
	}
	o.s.Append(action)
	baseRepo.WaitTillExists()

	return baseRepo
}
